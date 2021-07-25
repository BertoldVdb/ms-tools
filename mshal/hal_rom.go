package mshal

import (
	"bytes"
	"encoding/hex"
	"time"
)

type cbApplyParamType func(h *HAL, out []byte) error
type cbPostExchangeType func(h *HAL, addr int, buf []byte) error

type romCommand struct {
	id byte

	is16bit    bool
	isWrite    bool
	maxPayload int

	cbApplyParam   cbApplyParamType
	cbPostExchange cbPostExchangeType
}

func romCommandMake(id byte, is16bit bool, isWrite bool) romCommand {
	return romCommand{
		id:      id,
		is16bit: is16bit,
		isWrite: isWrite,
	}
}

func romCommandMakeReadWrite(id byte, is16bit bool) (romCommand, romCommand) {
	return romCommandMake(id, is16bit, false), romCommandMake(id+1, is16bit, true)
}

func (h *HAL) romExchangeReport(out [9]byte, checkLen int) ([9]byte, error) {
	var in [9]byte

	if h.config.LogFunc != nil {
		h.config.LogFunc(3, "ROMOut:   %s", hex.EncodeToString(out[:]))
	}

	if _, err := h.dev.SendFeatureReport(out[:]); err != nil {
		return in, err
	}

	_, err := h.dev.GetFeatureReport(in[:])
	if err != nil {
		return in, err
	}

	if h.config.LogFunc != nil {
		h.config.LogFunc(3, "ROMIn:    %s", hex.EncodeToString(in[:]))
	}

	if !bytes.Equal(out[:checkLen], in[:checkLen]) {
		return in, ErrorInvalidResponse
	}
	return in, nil
}

func (h *HAL) romProtocolMakeHeader(cmd byte, is16bit bool, addr int) ([9]byte, int) {
	var out [9]byte

	out[0] = 0
	out[1] = byte(cmd)

	index := 2
	if is16bit {
		out[index] = byte(addr >> 8)
		index++
	}
	out[index] = byte(addr)
	index++

	return out, index
}

func (h *HAL) romProtocolReadReply(buf []byte, in []byte, index int, maxReply int) int {
	if len(buf) > maxReply {
		buf = buf[:maxReply]
	}
	return copy(buf, in[index:])

}

func (h *HAL) romProtocolWritePayload(buf []byte, out []byte, index int, maxLen int) int {
	if len(buf) > maxLen {
		buf = buf[:maxLen]
	}
	return copy(out[index:], buf)
}

func (h *HAL) romProtocolExec(cmd romCommand, addr int, buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	maxPayload := cmd.maxPayload
	if maxPayload <= 0 {
		maxPayload = 1
	}

	out, index := h.romProtocolMakeHeader(cmd.id, cmd.is16bit, addr)

	txrLen := 0
	if cmd.isWrite {
		txrLen = h.romProtocolWritePayload(buf, out[:], index, maxPayload)
	}

	if cmd.cbApplyParam != nil {
		if err := cmd.cbApplyParam(h, out[:]); err != nil {
			return 0, err
		}
	}

	in, err := h.romExchangeReport(out, index)
	if err != nil {
		return 0, err
	}

	if !cmd.isWrite {
		txrLen = h.romProtocolReadReply(buf, in[:], index, maxPayload)
	}

	if cmd.cbPostExchange != nil {
		if err := cmd.cbPostExchange(h, addr, buf); err != nil {
			return 0, err
		}
	}

	return txrLen, nil
}

type halROMMemoryRegion struct {
	hal          *HAL
	readCommand  *romCommand
	writeCommand *romCommand
	baseAddr     int
	length       int
	name         MemoryRegionNameType
}

func (h halROMMemoryRegion) GetName() MemoryRegionNameType {
	return h.name
}

func (h halROMMemoryRegion) GetLength() int {
	return h.length
}

func (h halROMMemoryRegion) GetParent() (MemoryRegion, int) {
	return nil, 0
}

func (h halROMMemoryRegion) Access(write bool, addr int, buf []byte) (int, error) {
	if addr > h.length {
		return 0, nil
	}
	if addr+len(buf) > h.length {
		buf = buf[:h.length-addr]
	}

	if write {
		if h.writeCommand == nil {
			return 0, ErrorWriteNotAllowed
		}
		return h.hal.romProtocolExec(*h.writeCommand, h.baseAddr+addr, buf)
	}

	if h.writeCommand == nil {
		return 0, ErrorWriteNotAllowed
	}
	return h.hal.romProtocolExec(*h.readCommand, h.baseAddr+addr, buf)
}

func (h *HAL) romMemoryRegionMake(name MemoryRegionNameType, baseAddr int, length int, read *romCommand, write *romCommand) MemoryRegion {
	return regionWrapCompleteIO(halROMMemoryRegion{
		hal:          h,
		baseAddr:     baseAddr,
		length:       length,
		readCommand:  read,
		writeCommand: write,
		name:         name,
	})
}

func romEepromV2HandleTwoByteAddress(h *HAL, out []byte) error {
	if h.eepromSize > 2048 {
		out[8] = 1
	}
	return nil
}

func romEepromVerify(region MemoryRegion) cbPostExchangeType {
	return func(h *HAL, addr int, buf []byte) error {
		if buf[0] == 0 {
			/* The chip returns 0 if there is no I2C response, so we just have to wait */
			time.Sleep(15 * time.Millisecond)
			return nil
		}

		var tmp [1]byte
		for i := 0; i < 25; i++ {
			if _, err := region.Access(false, addr, tmp[:]); err != nil {
				return err
			}
			if tmp[0] == buf[0] {
				return nil
			}
		}
		return ErrorTimeout
	}
}
