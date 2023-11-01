package ms213x

import (
	"encoding/binary"
	"errors"
	"fmt"
)

func calcSum(f []byte) uint16 {
	var csum uint16
	for _, m := range f {
		csum += uint16(m)
	}
	return csum
}

func work(f []byte, fix bool) error {
	if len(f) < 0x30+4 {
		return errors.New("file too short (hdr)")
	}

	if t := binary.BigEndian.Uint16(f); t != 0x5aa5 && t != 0x6996 && t != 0x3cc3 {
		return fmt.Errorf("unknown flash type: %x", t)
	}

	codeLen := int(binary.BigEndian.Uint16(f[2:]))
	end := 0x30 + codeLen
	if len(f) < end+4 {
		return errors.New("file too short (code)")
	}

	hdrSum := calcSum(f[2:12]) + calcSum(f[16:0x30])
	codeSum := calcSum(f[0x30:end])

	if !fix {
		if hdrImg := binary.BigEndian.Uint16(f[end:]); hdrSum != hdrImg {
			return fmt.Errorf("header checksum mismatch: %x != %x", hdrSum, hdrImg)
		} else if codeImg := binary.BigEndian.Uint16(f[end+2:]); codeSum != codeImg {
			return fmt.Errorf("code checksum mismatch: %x != %x", codeSum, codeImg)
		}
	} else {
		binary.BigEndian.PutUint16(f[end:], hdrSum)
		binary.BigEndian.PutUint16(f[end+2:], codeSum)
	}

	return nil
}

func CheckImage(f []byte) error {
	return work(f, false)
}

func FixImage(f []byte) {
	work(f, true)
}
