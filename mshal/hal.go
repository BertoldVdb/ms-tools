package mshal

import (
	"github.com/sstallion/go-hid"
)

type HAL struct {
	dev *hid.Device

	deviceType      int
	deviceTypeExtra int
	eepromSize      int

	patchAllocAddr int
	patchCallAddrs []int

	patchInstalled bool

	config HALConfig
}

type LogFunc func(level int, format string, param ...interface{})

type HALConfig struct {
	EEPromSize int

	PatchTryInstall         bool
	PatchIgnoreUserFirmware bool
	PatchProbeEEPROM        bool

	LogFunc LogFunc
}

func New(dev *hid.Device, config HALConfig) (*HAL, error) {
	h := &HAL{
		dev:    dev,
		config: config,
	}

	xdata := h.MemoryRegionGet(MemoryRegionRAM)
	/* This is a value that is set by the ROM, so we can ID the chip from it */
	chipType, err := ReadByte(xdata, 0xF800)
	if err != nil {
		return nil, err
	}

	switch chipType {
	case 0x6a:
		h.deviceType = 2106

		chipType, err = ReadByte(xdata, 0x35)
		h.deviceTypeExtra = int(chipType)

	case 0xa7:
		h.deviceType = 2109
	default:
		return nil, ErrorUnknownDevice
	}

	if h.config.LogFunc != nil {
		h.config.LogFunc(1, "Detected %s", h.GetDeviceType())
	}

	if config.PatchTryInstall {
		isNew, err := h.patchInstall()
		if err != nil {
			return nil, err
		}

		if h.config.LogFunc != nil {
			if isNew {
				h.config.LogFunc(1, "Patch installed")
			} else {
				h.config.LogFunc(1, "Patch already installed")
			}
		}

		h.patchInstalled = true
	}

	if h.eepromSize == 0 && config.PatchProbeEEPROM {
		h.eepromSize, err = h.patchEepromDetectSize()
		if err != nil {
			h.eepromSize = 2048
			h.config.LogFunc(1, "Failed to detect EEPROM: %v", err)
		}
	}

	if h.deviceType == 2106 && h.eepromSize > 2048 {
		h.eepromSize = 2048
	}

	h.config.LogFunc(1, "Assumed EEPROM Size: %d", h.eepromSize)

	return h, nil
}

type MemoryRegionNameType string

const (
	MemoryRegionCODE       MemoryRegionNameType = "CODE"
	MemoryRegionRAM        MemoryRegionNameType = "RAM"
	MemoryRegionIRAM       MemoryRegionNameType = "IRAM"
	MemoryRegionEEPROM     MemoryRegionNameType = "EEPROM"
	MemoryRegionUserConfig MemoryRegionNameType = "USERCONFIG"

	MemoryRegionUserRAM MemoryRegionNameType = "USERRAM"

	MemoryRegionRegisters2106TVD MemoryRegionNameType = "TVDREGS"
)

type HookNameType string

func (h *HAL) GetDeviceType() string {
	if h.deviceType == 2106 {
		if h.deviceTypeExtra != 0 {
			return "MS2106s"
		}
		return "MS2106"
	}

	return "MS2109"
}
