package main

import (
	"errors"
	"fmt"
	"strconv"
)

type GPIOGet struct {
}

func (g *GPIOGet) Run(c *Context) error {
	value, isOutput, err := c.hal.GPIOUpdate(0, 0, 0, 0)
	if err != nil {
		return err
	}

	fmt.Printf("Pin:    76543210\nValue:  %08s\nOutput: %08s\n", strconv.FormatInt(int64(value), 2), strconv.FormatInt(int64(isOutput), 2))

	return nil
}

type GPIOSet struct {
	Command string `arg name:"command" help:"Set value: gpio=value (eg: 4=1). Set as input: gpio? (eg: 4?)" type:"string"`
}

func parseDigit(digit byte) int {
	if digit < '0' || digit > '9' {
		return 0
	}

	return int(digit - '0')
}

func (g *GPIOSet) Run(c *Context) error {
	if len(g.Command) == 3 && g.Command[1] == '=' {
		return c.hal.GPIOWrite(parseDigit(g.Command[0]), g.Command[2] != '0')
	} else if len(g.Command) == 2 && g.Command[1] == '?' {
		value, err := c.hal.GPIORead(parseDigit(g.Command[0]))
		if err != nil {
			return err
		}
		fmt.Println(value)
	} else {
		return errors.New("Invalid syntax")
	}
	return nil
}
