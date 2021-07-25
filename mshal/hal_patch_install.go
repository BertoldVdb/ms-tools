package mshal

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
)

func (h *HAL) patchAlloc(len int) int {
	addr := h.patchAllocAddr
	h.patchAllocAddr += len
	if h.config.LogFunc != nil {
		h.config.LogFunc(2, "Allocated %d bytes for patch at %04x", len, addr)
	}
	return addr
}

func (h *HAL) patchWriteWithTempFirstByte(region MemoryRegion, addr int, data []byte, firstByte byte) error {
	if len(data) == 0 {
		return nil
	}

	if h.config.LogFunc != nil {
		h.config.LogFunc(2, "Safe writing blob at %04x: %s", addr, hex.EncodeToString(data))
	}

	if err := WriteByte(region, addr, firstByte); err != nil {
		return err
	}

	if _, err := region.Access(true, addr+1, data[1:]); err != nil {
		return err
	}

	return WriteByte(region, addr, data[0])
}

func (h *HAL) patchWriteWithRET(region MemoryRegion, addr int, data []byte) error {
	return h.patchWriteWithTempFirstByte(region, addr, data, 0x22)
}

func patchTrampolineEncode(orig []byte, origAddr int, R0Value byte, hookAddr int) []byte {
	// ...orig...
	// LCALL origAddr
	// MOV   R0, #R0Value
	// LJMP  hookAddr

	trampolineOrig := []byte{
		0x12, byte(origAddr >> 8), byte(origAddr),
	}

	trampolineHook := []byte{
		0x78, R0Value,
		0x02, byte(hookAddr >> 8), byte(hookAddr),
	}

	result := orig
	if origAddr != 0 {
		result = append(result, trampolineOrig...)
	}
	result = append(result, trampolineHook...)

	return result
}

func (h *HAL) patchTrampolineInstall(ram MemoryRegion, replaceCode bool, addr int, R0value byte, hookAddr int) error {
	var trampoline []byte
	if replaceCode {
		replaceLen := 0
		var in [14]byte

		_, err := ram.Access(false, addr, in[:])
		if err != nil {
			return err
		}

		/* Can we patch this code? */
		if in[0] == 0x2 || in[0] == 0x12 || in[0] == 0x90 {
			replaceLen = 3
		} else if in[0] == 0xe5 && in[1] == 0x33 && in[2] == 0x30 {
			replaceLen = 14
		} else {
			return ErrorPatchFailed
		}

		trampoline = patchTrampolineEncode(in[:replaceLen], addr+replaceLen, R0value, hookAddr)
	} else {
		trampoline = patchTrampolineEncode(nil, 0, R0value, hookAddr)
	}

	trampolineAddr := h.patchAlloc(len(trampoline))

	if h.config.LogFunc != nil {
		h.config.LogFunc(2, "Writing trampoline at %04x: %s", trampolineAddr, hex.EncodeToString(trampoline))
	}

	if _, err := ram.Access(true, trampolineAddr, trampoline); err != nil {
		return err
	}

	return h.patchWriteWithRET(ram, addr, []byte{0x02, byte(trampolineAddr >> 8), byte(trampolineAddr)})
}

type blob struct {
	data  []byte
	reloc func(dataCopy []byte, addr int) (int, []byte)
}

//go:embed asm/hook_2106.bin
var codeCallgate2106 []byte

//go:embed asm/hook_2109.bin
var codeCallgate2109 []byte

func relocateCallgate(result []byte, addr int) (int, []byte) {
	if result[5] != 0x12 {
		panic("Offset 5 is not LCALL")
	}

	callAddr := binary.BigEndian.Uint16(result[6:])
	callAddr += uint16(addr)
	binary.BigEndian.PutUint16(result[6:], callAddr)

	return addr, result
}

//go:embed asm/gpio.bin
var codeGpio []byte

//go:embed asm/code.bin
var codeMOVC []byte

//go:embed asm/i2cRead2109.bin
var codei2cRead []byte

var installBlobs2106 = []blob{
	{
		data:  codeCallgate2106,
		reloc: relocateCallgate,
	}, {
		data: codeGpio,
	}, {
		data: codeMOVC,
	}}

var installBlobs2109 = []blob{
	{
		data:  codeCallgate2109,
		reloc: relocateCallgate,
	}, {
		data: codeGpio,
	}, {
		data: codeMOVC,
	}, {
		data: codei2cRead,
	},
}

func (h *HAL) EEPROMReloadUser() error {
	if h.config.LogFunc != nil {
		h.config.LogFunc(1, "Reloading EEPROM code")
	}

	ram := h.MemoryRegionGet(MemoryRegionRAM)
	userConfig := h.MemoryRegionGet(MemoryRegionUserConfig)

	addr, _, err := h.patchHookGet(userConfig, true)
	if err != nil {
		return err
	}

	/* Write RET and enable callback */
	if err := h.patchHookConfigure(userConfig, true, false); err != nil {
		return err
	}

	loadEEPROM := []byte{0x02, 0x12, 0x82}
	if h.deviceType == 2109 {
		loadEEPROM = []byte{0x02, 0x5f, 0x19}
	}

	/* Reload EEPROM from IRQ context */
	if err := h.patchWriteWithRET(ram, addr, loadEEPROM); err != nil {
		return err
	}

	return h.patchHookConfigure(userConfig, true, true)
}

func (h *HAL) EEPROMIgnoreUser() error {
	if h.config.LogFunc != nil {
		h.config.LogFunc(1, "Unloading EEPROM code")
	}

	userConfig := h.MemoryRegionGet(MemoryRegionUserConfig)

	ff := bytes.Repeat([]byte{0xFF}, userConfig.GetLength())
	ff[4] = 0
	ff[5] = 0

	_, err := userConfig.Access(true, 0, ff)
	return err
}

func (h *HAL) EEPROMIsLoaded() (bool, int, error) {
	userconfig := h.MemoryRegionGet(MemoryRegionUserConfig)

	var hdr [4]byte
	if _, err := userconfig.Access(false, 0, hdr[:]); err != nil {
		return false, 0, err
	}

	eepromLen := int(binary.BigEndian.Uint16(hdr[2:]))

	if h.deviceType == 2106 {
		return hdr[0] == 0x5a && hdr[1] == 0xa5, eepromLen, nil
	}

	if hdr[0] == 0xa5 && hdr[1] == 0x5a {
		return true, eepromLen, nil
	}

	/* 2109 can also use 16bit eeproms and they have a different header */
	return hdr[0] == 0x96 && hdr[1] == 0x69, eepromLen, nil
}

func (h *HAL) patchHookGet(loc MemoryRegion, inIRQ bool) (int, bool, error) {
	if h.deviceType == 2106 {
		if inIRQ {
			value, err := ReadByte(loc, 0x9)
			return 0xc4a0, value == 0x96, err
		} else {
			value, err := ReadByte(loc, 0x5)
			return 0xc420, value == 0x5a, err
		}
	}

	value, err := ReadByte(loc, 0x4)
	if err != nil {
		return 0, false, err
	}

	if inIRQ {
		return 0xcc20, value&4 > 0, nil
	}

	return 0xcc00, value&1 > 0, nil
}
func (h *HAL) patchHookConfigure(loc MemoryRegion, inIRQ bool, enable bool) error {
	if h.config.LogFunc != nil {
		h.config.LogFunc(2, "Configuring userhook: inIRQ=%v, enable=%v", inIRQ, enable)
	}

	if h.deviceType == 2106 {
		value := byte(0)
		if inIRQ {
			if enable {
				value = 0x96
			}
			return WriteByte(loc, 0x9, value)
		} else {
			if enable {
				value = 0x5a
			}
			return WriteByte(loc, 0x5, value)
		}
	}

	value, err := ReadByte(loc, 0x4)
	if err != nil {
		return err
	}

	if inIRQ {
		value &= ^byte(4)
		if enable {
			value |= 4
		}
	} else {
		value &= ^byte(1)
		if enable {
			value |= 1
		}
	}

	return WriteByte(loc, 0x4, value)
}

func (h *HAL) patchInitAlloc(userConfig MemoryRegion) (bool, error) {
	userCodePresent, userCodeLen, err := h.EEPROMIsLoaded()
	if err != nil {
		return userCodePresent, err
	}
	if !userCodePresent {
		userCodeLen = 256
	}
	_, userOffset := RecursiveGetParentAddress(userConfig, userConfig.GetLength())

	h.patchAllocAddr = userOffset + userCodeLen
	return userCodePresent, nil
}

func (h *HAL) patchInstall() (bool, error) {
	installBlobs := installBlobs2106
	if h.deviceType == 2109 {
		installBlobs = installBlobs2109
	}

	ram := h.MemoryRegionGet(MemoryRegionRAM)
	userConfig := h.MemoryRegionGet(MemoryRegionUserConfig)

	h.patchCallAddrs = make([]int, len(installBlobs))

	/* Calculate checksum of blobs */
	crc := crc32.New(crc32.IEEETable)
	for _, m := range installBlobs {
		crc.Write(m.data)
	}
	sum := crc.Sum(nil)

	if h.config.PatchIgnoreUserFirmware {
		sum[0] = ^sum[0]
	}

	if _, err := h.patchInitAlloc(userConfig); err != nil {
		return false, err
	}

	/* Read current stored patch info */
	sumBlock := make([]byte, len(sum)+2*len(installBlobs))
	sumBlockAddr := h.patchAlloc(len(sumBlock))

	if _, err := ram.Access(false, sumBlockAddr, sumBlock); err != nil {
		return false, err
	}

	/* Is this chip already patched? */
	if bytes.Equal(sumBlock[:4], sum) {
		sumBlock = sumBlock[4:]
		for i := range installBlobs {
			h.patchCallAddrs[i] = int(binary.BigEndian.Uint16(sumBlock))
			sumBlock = sumBlock[2:]
		}
		return false, nil
	}

	if !h.config.PatchIgnoreUserFirmware {
		/* Reload eeprom to unpatch */
		if err := h.EEPROMReloadUser(); err != nil {
			return true, err
		}
	} else {
		/* Remove EEPROM data */
		if err := h.EEPROMIgnoreUser(); err != nil {
			return false, err
		}
	}

	userCodePresent, err := h.patchInitAlloc(userConfig)
	if err != nil {
		return false, err
	}

	sumBlock = make([]byte, len(sum)+2*len(installBlobs))
	sumBlockAddr = h.patchAlloc(len(sumBlock))
	copy(sumBlock, sum)

	/* Install all blobs */
	for i, m := range installBlobs {
		data := m.data

		loadAddr := h.patchAlloc(len(data))
		callAddr := loadAddr

		if m.reloc != nil {
			dataCopy := make([]byte, len(data))
			copy(dataCopy, data)
			callAddr, data = m.reloc(dataCopy, loadAddr)
		}

		if h.config.LogFunc != nil {
			h.config.LogFunc(2, "Writing blob at %04x: %s", loadAddr, hex.EncodeToString(data))
		}

		_, err := ram.Access(true, loadAddr, data)
		if err != nil {
			return true, err
		}

		h.patchCallAddrs[i] = callAddr
	}

	/* Check current state */
	addrIrq, enableIrq, err := h.patchHookGet(userConfig, true)
	if err != nil {
		return true, err
	}
	addrNorm, enableNorm, err := h.patchHookGet(userConfig, false)
	if err != nil {
		return true, err
	}

	if userCodePresent && (!enableIrq || !enableNorm) {
		return true, ErrorPatchFailed
	}

	/* Install trampolines to callgate */
	if err := h.patchTrampolineInstall(ram, userCodePresent, addrIrq, 0xee, h.patchCallAddrs[0]); err != nil {
		return true, err
	}

	if err := h.patchTrampolineInstall(ram, userCodePresent, addrNorm, 0xef, h.patchCallAddrs[0]); err != nil {
		return true, err
	}

	/* Enable callbacks */
	if !userCodePresent {
		if err := h.patchHookConfigure(userConfig, true, true); err != nil {
			return true, err
		}
		if err := h.patchHookConfigure(userConfig, false, true); err != nil {
			return true, err
		}
	}

	/* Write patch sumblock */
	for i := range installBlobs {
		binary.BigEndian.PutUint16(sumBlock[4+(2*i):], uint16(h.patchCallAddrs[i]))
	}

	return true, h.patchWriteWithTempFirstByte(ram, sumBlockAddr, sumBlock, sumBlock[0]-1)
}
