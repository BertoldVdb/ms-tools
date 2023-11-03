package mshal

import (
	"bytes"

	"github.com/BertoldVdb/ms-tools/gohid"
)

type HAL struct {
	dev gohid.HIDDevice

	deviceType      int
	deviceTypeExtra int
	eepromSize      int

	patchAllocAddr              int
	patchCallAddrsExternalStart int
	patchCallAddrs              []int

	patchInstalled bool
	patchCanCall   bool

	config           HALConfig
	ms2130spiEnabled int
}

type LogFunc func(level int, format string, param ...interface{})

type HALConfig struct {
	EEPromSize int

	PatchTryInstall         bool
	PatchIgnoreUserFirmware bool
	PatchProbeEEPROM        bool
	PatchBlobs              []CodeBlob

	LogFunc LogFunc
}

func New(dev gohid.HIDDevice, config HALConfig) (*HAL, error) {
	h := &HAL{
		dev:    dev,
		config: config,

		ms2130spiEnabled: -1,
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
		if err != nil {
			return nil, err
		}
		h.deviceTypeExtra = int(chipType)

	case 0xa7:
		h.deviceType = 2109

	case 0xff: /* TODO: Find a better ID register, as this will likely match many devices */
		h.deviceType = 2107

	case 0x00: /* TODO: Find a better ID register, as this will likely match many devices */
		h.deviceType = 2130
		config.PatchTryInstall = false

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

	h.eepromSize = config.EEPromSize

	if h.eepromSize == 0 && config.PatchProbeEEPROM {
		h.eepromSize, err = h.patchEepromDetectSize()
		if err != nil {
			if h.deviceType != 2130 {
				h.eepromSize = 2048
			} else {
				h.eepromSize = 64 * 1024
			}
			h.config.LogFunc(1, "Failed to detect EEPROM: %v", err)
		}
	}

	if h.deviceType == 2106 && h.eepromSize > 2048 {
		h.eepromSize = 2048
	}

	h.config.LogFunc(1, "Assumed EEPROM Size: %d", h.eepromSize)

	/* MS2130 can be running code from flash that is pre-patched. This is a hack to allow
	 * using that even withtout offset discovery */
	if h.deviceType == 2130 {
		var id [4]byte
		if _, err := xdata.Access(false, 0x7b00, id[:]); err != nil {
			return nil, err
		}
		if bytes.Equal(id[:], []byte("BVDB")) {
			h.patchCanCall = true
			h.config.LogFunc(1, "MS213x firmware supports calling functions")
		}
	}

	return h, nil
}

type MemoryRegionNameType string

const (
	MemoryRegionCODE             MemoryRegionNameType = "CODE"
	MemoryRegionRAM              MemoryRegionNameType = "RAM"
	MemoryRegionIRAM             MemoryRegionNameType = "IRAM"
	MemoryRegionSFR              MemoryRegionNameType = "SFR"
	MemoryRegionEEPROM           MemoryRegionNameType = "EEPROM"
	MemoryRegionUserConfig       MemoryRegionNameType = "USERCONFIG"
	MemoryRegionUserRAM          MemoryRegionNameType = "USERRAM"
	MemoryRegionRegisters2106TVD MemoryRegionNameType = "TVDREGS"
	MemoryRegionB7_0             MemoryRegionNameType = "B7_0"
	MemoryRegionB7_1             MemoryRegionNameType = "B7_1"
	MemoryRegionB9               MemoryRegionNameType = "B9"
	MemoryRegionFLASH            MemoryRegionNameType = "FLASH"
)

type HookNameType string

func (h *HAL) GetDeviceType() string {
	if h.deviceType == 2106 {
		if h.deviceTypeExtra != 0 {
			return "MS2106s"
		}
		return "MS2106"
	} else if h.deviceType == 2130 {
		return "MS2130"
	} else if h.deviceType == 2107 {
		return "MS2107"
	}

	return "MS2109"
}
