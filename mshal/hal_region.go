package mshal

import "strings"

func (h *HAL) memoryRegionXDATAIRAM(base int, len int) MemoryRegion {
	read, write := romCommandMakeReadWrite(0xb5, true)
	if h.deviceType == 2130 {
		read.maxPayload = 4
	}
	return h.romMemoryRegionMake(MemoryRegionRAM, base, len, 1, &read, &write)
}

func (h *HAL) memoryRegionSFR(base int, len int) MemoryRegion {
	read, write := romCommandMakeReadWrite(0xc5, false)
	return h.romMemoryRegionMake(MemoryRegionSFR, base, len, 1, &read, &write)
}

func (h *HAL) memoryRegionB7(base int, len int, index byte) MemoryRegion {
	indexFunc := func(h *HAL, out []byte) error {
		out[2] = index
		return nil
	}

	read, write := romCommandMakeReadWrite(0xb7, true)
	read.offset = 3
	read.maxPayload = 4
	read.cbApplyParam = indexFunc
	write.offset = 3
	write.maxPayload = 4
	write.cbApplyParam = indexFunc
	t := MemoryRegionB7_0
	if index > 0 {
		t = MemoryRegionB7_1
	}

	return h.romMemoryRegionMake(t, base, len, 4, &read, &write)
}

func (h *HAL) memoryRegionB9(base int, len int) MemoryRegion {
	// Note: If the first byte of the request is not 0, the previous address is reused
	read, write := romCommandMakeReadWrite(0xb9, false)
	read.offset = 3
	read.addrShift = 1
	read.maxPayload = 2
	write.offset = 3
	write.addrShift = 1
	write.maxPayload = 2

	return h.romMemoryRegionMake(MemoryRegionB9, base, len, 2, &read, &write)
}

func (h *HAL) memoryRegionEEPROM() MemoryRegion {
	read, write := romCommandMakeReadWrite(0xe5, true)
	if h.deviceType != 2106 {
		read.maxPayload = 5
		read.cbApplyParam = romEepromV2HandleTwoByteAddress
	}
	region := h.romMemoryRegionMake(MemoryRegionEEPROM, 0, h.eepromSize, 1, &read, &write)

	/* The EEPROM takes time to write, so try to read again.
	   MS2109 has internal delay (quite long) */
	if h.deviceType != 2109 {
		write.cbPostExchange = romEepromVerify(region)
	}

	return region
}

func (h *HAL) memoryRegionRegisters2106TVD() MemoryRegion {
	read, write := romCommandMakeReadWrite(0xa5, false)
	return h.romMemoryRegionMake(MemoryRegionRegisters2106TVD, 0, 256, 1, &read, &write)
}

func (h *HAL) MemoryRegionList() []MemoryRegionNameType {
	list := []MemoryRegionNameType{
		MemoryRegionRAM,
		MemoryRegionIRAM,
		MemoryRegionEEPROM,
		MemoryRegionUserConfig,
	}

	if h.deviceType == 2106 || h.deviceType == 2109 {
		list = append(list, MemoryRegionUserRAM)
	}

	if h.deviceType == 2130 {
		list = append(list, MemoryRegionSFR)
		list = append(list, MemoryRegionB7_0)
		list = append(list, MemoryRegionB7_1)
		list = append(list, MemoryRegionB9)
		list = append(list, MemoryRegionFLASH)
	}

	if h.patchInstalled {
		list = append(list, MemoryRegionCODE)
	}

	if h.deviceType == 2106 {
		list = append(list, MemoryRegionRegisters2106TVD)
	}

	return list
}

func (h *HAL) MemoryRegionGet(name MemoryRegionNameType) MemoryRegion {
	t := MemoryRegionNameType(strings.ToUpper(string(name)))

	switch t {
	case MemoryRegionRAM:
		return h.memoryRegionXDATAIRAM(0, 0x10000)
	case MemoryRegionIRAM:
		return regionWrapPartial(MemoryRegionIRAM, h.MemoryRegionGet(MemoryRegionRAM), 0, 0x100)
	case MemoryRegionEEPROM:
		if h.patchInstalled {
			if h.config.LogFunc != nil {
				h.config.LogFunc(1, "Using patched EEPROM access")
			}

			return h.patchMakeEEPROMRegion()
		}

		return h.memoryRegionEEPROM()
	case MemoryRegionCODE:
		return h.patchMakeCodeRegion()
	}

	if h.deviceType == 2106 {
		switch t {
		case MemoryRegionUserRAM:
			return regionWrapPartial(MemoryRegionUserRAM, h.MemoryRegionGet(MemoryRegionRAM), 0xC000, 0x1000)

		case MemoryRegionUserConfig:
			return regionWrapPartial(MemoryRegionUserConfig, h.MemoryRegionGet(MemoryRegionUserRAM), 0x3F0, 0x10)

		case MemoryRegionRegisters2106TVD:
			if h.patchInstalled {
				if h.config.LogFunc != nil {
					h.config.LogFunc(1, "Using patched TVD access")
				}

				return h.patchMakeTVDRegion()
			}

			return h.memoryRegionRegisters2106TVD()
		}
	}

	if h.deviceType == 2109 {
		switch t {
		case MemoryRegionUserRAM:
			return regionWrapPartial(MemoryRegionUserRAM, h.MemoryRegionGet(MemoryRegionRAM), 0xC000, 0x2000)

		case MemoryRegionUserConfig:
			return regionWrapPartial(MemoryRegionUserConfig, h.MemoryRegionGet(MemoryRegionUserRAM), 0xBD0, 0x30)
		}
	}

	if h.deviceType == 2130 {
		switch t {
		case MemoryRegionSFR:
			return h.memoryRegionSFR(0x80, 0x80)

		case MemoryRegionB7_0:
			return h.memoryRegionB7(0, 65536, 0)

		case MemoryRegionB7_1:
			return h.memoryRegionB7(0, 65536, 1)

		case MemoryRegionB9:
			return h.memoryRegionB9(0, 512)

		case MemoryRegionUserConfig:
			return regionWrapPartial(MemoryRegionUserConfig, h.MemoryRegionGet(MemoryRegionRAM), 0x1FD0, 0x30)

		case MemoryRegionFLASH:
			return h.memoryRegionFlash()
		}
	}

	return nil
}
