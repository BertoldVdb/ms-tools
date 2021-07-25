package mshal

import "strings"

func (h *HAL) memoryRegionXDATAIRAM(base int, len int) MemoryRegion {
	read, write := romCommandMakeReadWrite(0xb5, true)
	return h.romMemoryRegionMake(MemoryRegionRAM, base, len, &read, &write)
}

func (h *HAL) memoryRegionEEPROM() MemoryRegion {
	read, write := romCommandMakeReadWrite(0xe5, true)
	if h.deviceType == 2109 {
		read.maxPayload = 5
		read.cbApplyParam = romEepromV2HandleTwoByteAddress
	}
	region := h.romMemoryRegionMake(MemoryRegionEEPROM, 0, h.eepromSize, &read, &write)

	/* The EEPROM takes time to write, so try to read again.
	   MS2109 has internal delay (quite long) */
	if h.deviceType != 2109 {
		write.cbPostExchange = romEepromVerify(region)
	}

	return region
}

func (h *HAL) memoryRegionRegisters2106TVD() MemoryRegion {
	read, write := romCommandMakeReadWrite(0xa5, false)
	return h.romMemoryRegionMake(MemoryRegionRegisters2106TVD, 0, 256, &read, &write)
}

func (h *HAL) MemoryRegionList() []MemoryRegionNameType {
	list := []MemoryRegionNameType{
		MemoryRegionRAM,
		MemoryRegionIRAM,
		MemoryRegionEEPROM,
		MemoryRegionUserRAM,
		MemoryRegionUserConfig,
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

	return nil
}
