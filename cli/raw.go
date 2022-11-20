package main

import (
	"encoding/hex"
	"fmt"
)

type RawCmd struct {
	Data string `arg name:"value" help:"Value to write."`
}

func (w RawCmd) Run(c *Context) error {
	buf, err := hex.DecodeString(w.Data)
	if err != nil {
		return err
	}

	out, err := c.hal.ROMExchangeReport(buf)
	if err != nil {
		return err
	}

	fmt.Println("Raw command results:", hex.EncodeToString(buf), "->", hex.EncodeToString(out))
	return nil
}
