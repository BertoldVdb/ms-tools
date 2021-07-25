package main

import (
	"fmt"

	"github.com/fatih/color"
)

func hexdump(offset int, data []byte, mark []bool) string {
	var result string
	red := color.New(color.FgRed)

	for len(data) > 0 {
		l := len(data)
		if l > 32 {
			l = 32
		}
		work := data[:l]
		data = data[l:]
		var workMark []bool
		if mark != nil {
			workMark = mark[:l]
			mark = mark[l:]
		}

		var workHex string
		var workAscii string
		for i := 0; i < 32; i++ {
			m := byte(0)
			valid := i < len(work)
			delta := false
			if valid {
				m = work[i]
				if workMark != nil && workMark[i] {
					delta = true
				}
			}

			if valid {
				if delta {
					workHex += red.Sprintf("%02x ", m)
				} else {
					workHex += fmt.Sprintf("%02x ", m)
				}

				if m < 32 || m > 126 {
					m = '.'
				}
				if delta {
					workAscii += red.Sprintf("%c", m)
				} else {
					workAscii += fmt.Sprintf("%c", m)
				}
			} else {
				workHex += "   "
				workAscii += " "
			}
			if i%8 == 7 {
				workHex += " "
			}
		}

		result += fmt.Sprintf("%08x  %s|%s|\n", offset, workHex, workAscii)
		offset += l
	}

	return result
}
