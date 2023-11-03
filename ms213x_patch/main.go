package main

import (
	"encoding/binary"
	"flag"
	"log"
	"os"

	"github.com/BertoldVdb/ms-tools/mshal/ms213x"
)

func main() {
	input := flag.String("input", "", "Input filename")
	output := flag.String("output", "/tmp/modified.bin", "Output filename")
	flag.Parse()

	in, err := os.ReadFile(*input)
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}

	if err := ms213x.CheckImage(in); err != nil {
		log.Fatalln("Failed to parse image:", in)
	}

	clen := binary.BigEndian.Uint16(in[2:])
	code := in[0x30:]
	code = code[:clen]

	code, err = patch(code)
	if err != nil {
		log.Fatalln("Failed to patch code:", err)
	}

	out := make([]byte, 0x30, 0x30+len(code)+4)
	copy(out, in)
	binary.BigEndian.PutUint16(out[2:], uint16(len(code)))

	out = append(out, code...)
	out = binary.BigEndian.AppendUint32(out, 0)
	ms213x.FixImage(out)

	os.WriteFile("/tmp/modcode.bin", code, 0644)

	if err := os.WriteFile(*output, out, 0644); err != nil {
		log.Fatalln("Failed to write output:", err)
	}
}
