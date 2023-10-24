package main

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/sigurn/crc16"
)

type UARTTx struct {
	Data   string `arg name:"data" help:"Hex string to write to device"`
	Invert bool   `optional help:"Use RS232 polarity instead of TTL" default:"false"`
	Baud   int    `optional help:"Data rate in bits per second" default:"57600"`
}

func (l *UARTTx) Run(c *Context) error {
	buf, err := hex.DecodeString(l.Data)
	if err != nil {
		return err
	}

	return c.hal.UARTTransmit(l.Baud, buf, l.Invert)
}

type FlirTX struct {
	Func   int    `arg name:"func" type:"hex" help:"Function to call"`
	Params string `optional name:"data" help:"Parameters as hex string"`
}

func flirTauEncodeCommand(cmd byte, param []byte) []byte {
	var crcTab = crc16.MakeTable(crc16.CRC16_XMODEM)

	result := []byte{0x6E, 0, 0, cmd, 0, 0, 0, 0}
	binary.BigEndian.PutUint16(result[4:], uint16(len(param)))

	binary.BigEndian.PutUint16(result[6:], crc16.Update(0, result[:6], crcTab))

	result = append(result, param...)
	result = binary.BigEndian.AppendUint16(result, crc16.Update(0, result[8:], crcTab))
	return result
}

func (l *FlirTX) Run(c *Context) error {
	param, err := hex.DecodeString(l.Params)
	if err != nil {
		return err
	}

	return c.hal.UARTTransmit(56700, flirTauEncodeCommand(uint8(l.Func), param), true)
}
