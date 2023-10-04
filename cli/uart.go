package main

import (
	"encoding/hex"
)

type UARTTx struct {
	Data string `arg name:"data" help:"Hex string to write to device"`
	Baud int    `optional help:"Data rate in bits per second" default:"57600"`
}

func (l *UARTTx) Run(c *Context) error {
	buf, err := hex.DecodeString(l.Data)
	if err != nil {
		return err
	}

	return c.hal.UARTTransmit(l.Baud, buf)
}
