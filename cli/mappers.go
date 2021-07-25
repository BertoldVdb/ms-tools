package main

import (
	"reflect"
	"strconv"

	"github.com/alecthomas/kong"
)

type intMapper struct {
	base int
}

func (h intMapper) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	var value string
	err := ctx.Scan.PopValueInto("hex", &value)
	if err != nil {
		return err
	}
	i, err := strconv.ParseInt(value, h.base, 64)
	if err != nil {
		return err
	}
	target.SetInt(i)
	return nil
}
