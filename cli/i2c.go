package main

import (
	"encoding/hex"
	"fmt"

	"github.com/BertoldVdb/ms-tools/mshal"
)

type I2CScan struct {
}

func (l *I2CScan) Run(c *Context) error {
	fmt.Printf("Detected I2C devices:\r\n   ")
	for i := 0; i < 16; i++ {
		fmt.Printf("%02X ", i)
	}
	for i := byte(0); i < 0x80; i++ {
		ok, err := c.hal.I2CTransfer(i, []byte{0}, nil)
		if err != nil {
			return err
		}

		if i&15 == 0 {
			fmt.Printf("\r\n%02x ", i)
		}

		if ok {
			fmt.Printf("%02X ", i)
		} else {
			fmt.Printf("-- ")
		}
	}
	fmt.Println()
	return nil
}

type I2CTransfer struct {
	Addr int `arg name:"addr" help:"I2C device address" type:"int"`

	Write string `optional help:"Hex string to write to device"`
	Read  int    `optional help:"Number of bytes to read back"`
}

func (l *I2CTransfer) Run(c *Context) error {
	wrBuf, err := hex.DecodeString(l.Write)
	if err != nil {
		return err
	}

	rdBuf := make([]byte, l.Read)
	ok, err := c.hal.I2CTransfer(byte(l.Addr), wrBuf, rdBuf)
	if err != nil {
		return err
	}
	if !ok {
		return mshal.ErrorNoAck
	}

	fmt.Println(hexdump(0, rdBuf, nil))
	return nil
}
