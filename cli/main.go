package main

import (
	"fmt"
	"os"

	"github.com/BertoldVdb/ms-tools/mshal"
	"github.com/alecthomas/kong"
	"github.com/sstallion/go-hid"
)

type Context struct {
	dev *hid.Device
	hal *mshal.HAL
}

var CLI struct {
	VID      int    `optional type:"hex" help:"The USB Vendor ID." default:534d`
	PID      int    `optional type:"hex" help:"The USB Product ID."`
	Serial   string `optional help:"The USB Serial."`
	RawPath  string `optional help:"The USB Device Path."`
	LogLevel int    `optional help:"Higher values give more output."`

	NoPatch    bool `optional help:"Do not attempt to patch running firmware."`
	EEPROMSize int  `optional help:"Specify EEPROM size to skip autodetection."`
	NoFirmware bool `optional help:"Do not use firmware in EEPROM."`

	ListDev ListHIDCmd `cmd help:"List devices."`

	ListRegions MEMIOListRegions  `cmd help:"List available memory regions."`
	Read        MEMIOReadCmd      `cmd help:"Read and dump memory."`
	Write       MEMIOWriteCmd     `cmd help:"Write value to memory."`
	WriteFile   MEMIOWriteFileCmd `cmd help:"Write file to memory."`

	DumpROM DumpROM `cmd help:"Dump ROM (code) to file by uploading custom code."`

	I2CScan     I2CScan     `cmd name:"i2c-scan" help:"Scan I2C bus and show discovered devices."`
	I2CTransfer I2CTransfer `cmd name:"i2c-txfr" help:"Perform I2C transfer."`

	GPIOSet GPIOSet `cmd name:"gpio-set" help:"Set GPIO pin value and direction."`
	GPIOGet GPIOGet `cmd name:"gpio-get" help:"Get GPIO values."`
}

func main() {
	k, err := kong.New(&CLI,
		kong.NamedMapper("int", intMapper{}),
		kong.NamedMapper("hex", intMapper{base: 16}))
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, err := k.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	hid.Init()
	defer hid.Exit()

	c := &Context{}
	if ctx.Command() != "list-dev" {
		dev, err := OpenDevice()
		if err != nil {
			fmt.Println("Failed to open device", err)
			return
		}
		defer dev.Close()

		c.dev = dev
		config := mshal.HALConfig{
			PatchTryInstall: !CLI.NoPatch,

			PatchProbeEEPROM: true,
			EEPromSize:       CLI.EEPROMSize,

			PatchIgnoreUserFirmware: CLI.NoFirmware,

			LogFunc: func(level int, format string, param ...interface{}) {
				if level > CLI.LogLevel {
					return
				}
				str := fmt.Sprintf(format, param...)
				fmt.Printf("HAL(%d): %s\n", level, str)
			},
		}

		c.hal, err = mshal.New(dev, config)
		if err != nil {
			fmt.Println("Failed to create HAL", err)
			return
		}
	}

	err = ctx.Run(c)
	ctx.FatalIfErrorf(err)
}
