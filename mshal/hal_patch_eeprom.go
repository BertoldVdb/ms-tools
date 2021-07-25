package mshal

func (h *HAL) patchEepromDetectSize() (int, error) {
	eepromFound := 0
	for i := byte(0x50); i <= 0x57; i++ {
		ok, err := h.I2CTransfer(i, []byte{0}, nil)
		if err != nil {
			return 0, err
		}
		if !ok {
			break
		}
		eepromFound++
	}

	if eepromFound == 0 {
		return 0, nil
	}

	if eepromFound > 1 {
		return eepromFound * 256, nil
	}

	if h.deviceType == 2109 {
		/* If we find only one EEPROM we assume it is 16-bit addressable  since <256 byte EEPROMs
		   make no sense for this application. To actually test it you need to write to the chip :( */
		return 4096, nil
	}

	return 256, nil
}

func (h *HAL) patchEEPROMUnlock(unlock bool) error {
	if h.deviceType == 2109 {
		return h.GPIOWrite(5, !unlock)
	}

	return nil
}

type halPatchEEPROMMemoryRegion struct {
	hal *HAL
}

func (h halPatchEEPROMMemoryRegion) GetName() MemoryRegionNameType {
	return MemoryRegionEEPROM
}

func (h halPatchEEPROMMemoryRegion) GetLength() int {
	return h.hal.eepromSize
}

func (h halPatchEEPROMMemoryRegion) GetParent() (MemoryRegion, int) {
	return nil, 0
}

func (h halPatchEEPROMMemoryRegion) read(addr int, buf []byte) (int, error) {
	var ok bool
	var err error

	if h.GetLength() <= 2048 {
		ok, err = h.hal.I2CTransfer(0x50+byte(addr>>8), []byte{byte(addr)}, buf)
	} else {
		ok, err = h.hal.I2CTransfer(0x50, []byte{byte(addr >> 8), byte(addr)}, buf)
	}

	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrorNoAck
	}

	return len(buf), nil
}

func (h halPatchEEPROMMemoryRegion) write(addr int, buf []byte) (int, error) {
	if err := h.hal.patchEEPROMUnlock(true); err != nil {
		return 0, err
	}
	defer h.hal.patchEEPROMUnlock(false)

	/* We can write 16-byte pages */
	endOfPage := (addr + 16) / 16 * 16
	bytesRemaining := endOfPage - addr

	if len(buf) > bytesRemaining {
		buf = buf[:bytesRemaining]
	}

	var workBuf [16 + 2]byte
	copy(workBuf[2:], buf)
	workBuf[0] = byte(addr >> 8)
	workBuf[1] = byte(addr)
	wrBuf := workBuf[:len(buf)+2]

	var ok bool
	var err error

	if h.GetLength() <= 2048 {
		ok, err = h.hal.I2CTransfer(0x50+byte(addr>>8), wrBuf[1:], nil)
	} else {
		ok, err = h.hal.I2CTransfer(0x50, wrBuf, nil)
	}

	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrorNoAck
	}

	/* The EEPROM is working now, poll it to know when ACK received */
	for {
		_, err := h.read(0, []byte{0})
		if err == nil {
			break
		}
		if err != ErrorNoAck {
			return 0, err
		}
	}

	return len(buf), nil
}

func (h halPatchEEPROMMemoryRegion) Access(write bool, addr int, buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	if addr > h.GetLength() {
		return 0, nil
	}
	if addr+len(buf) > h.GetLength() {
		buf = buf[:h.GetLength()-addr]
	}

	if write {
		return h.write(addr, buf)
	}

	return h.read(addr, buf)
}

func (h *HAL) patchMakeEEPROMRegion() MemoryRegion {
	if !h.patchInstalled {
		return nil
	}

	return regionWrapCompleteIO(halPatchEEPROMMemoryRegion{hal: h})
}
