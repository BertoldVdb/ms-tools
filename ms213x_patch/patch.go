package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"

	_ "embed"
)

type patcher struct {
	image []byte

	hookOffset uint16
}

//go:embed asm/init.bin
var patchInit []byte

//go:embed asm/hook.bin
var patchHook []byte

//go:embed asm/finishf660.bin
var patchFinishf660 []byte

//go:embed asm/finishsig.bin
var patchFinishSig []byte

//go:embed asm/vsync.bin
var patchVSYNC []byte

//go:embed asm/readinfo.bin
var patchReadInfo []byte

//go:embed asm/readinfo2.bin
var patchReadInfo2 []byte

func (p *patcher) addCode(code []byte) uint16 {
	offs := len(p.image)
	p.image = append(p.image, code...)
	return uint16(offs)
}

func (p *patcher) detourCall(offset uint16, dest uint16) {
	a := p.image[offset]
	if a != 0x02 && a != 0x12 {
		panic("Not LJMP or LCALL")
	}

	var code [6]byte
	code[0] = 0x12 /* LCALL */
	binary.BigEndian.PutUint16(code[1:], dest)

	code[3] = 0x02                     /* LJMP */
	copy(code[4:], p.image[offset+1:]) /* Copy original address */

	binary.BigEndian.PutUint16(p.image[offset+1:], p.addCode(code[:]))
}

func (p *patcher) createHook(cmd uint8) uint16 {
	return p.addCode([]byte{
		0x78, cmd,
		0x02, byte(p.hookOffset >> 8), byte(p.hookOffset),
	})
}

func (p *patcher) replaceJump(offset, dest uint16) {
	p.image[offset] = 0x02
	binary.BigEndian.PutUint16(p.image[offset+1:], dest)
}

func (p *patcher) replaceCall(offset, dest uint16) {
	p.image[offset] = 0x12
	binary.BigEndian.PutUint16(p.image[offset+1:], dest)
}

func patch(in []byte) ([]byte, error) {
	/* Check if it is a file we know how to handle */
	hash := sha256.Sum256(in)
	hashStr := hex.EncodeToString(hash[:])
	if hashStr != "cc67f79a043da85dc8e6688a22111ade626e519e1ab549f110b3a06308190047" {
		return nil, fmt.Errorf("code hash %s is not supported", hashStr)
	}

	p := patcher{
		image: in,
	}

	/* Add init code */
	p.detourCall(0x4d48, p.addCode(patchInit))

	/* Add hook code and update its internal jump */
	p.hookOffset = p.addCode(patchHook)
	binary.BigEndian.PutUint16(p.image[p.hookOffset+8:], binary.BigEndian.Uint16(p.image[p.hookOffset+8:])+p.hookOffset)

	/* Add hook entry to main loop  */
	p.detourCall(0x4d70, p.createHook(0xEF)) /* Not in IRQ, but handled the same as the other USB commands */

	/* Add 0xEE, remove result codes 0xFF and 0xFE which are legitimate commands (function unclear, can still call them
	 * via hook if needed) */
	table, dflt := p.jumptableParse(0x1d9c)
	var newTable []jumptableEntry
	needInsert := true
	for _, b := range table {
		if b.Key > 0xEE && needInsert {
			needInsert = false
			newTable = append(newTable, jumptableEntry{
				Key:     0xEE,
				Address: p.createHook(0xEE),
			})
		}
		if b.Key >= 0xFE {
			break
		}
		newTable = append(newTable, b)
	}
	p.jumptableWrite(0x1d9c, newTable, dflt)

	/* Write 0xF660 also to safe place (0x7b10) */
	binary.BigEndian.PutUint16(p.image[0xbbb3:], 0x7b10)
	binary.BigEndian.PutUint16(p.image[0xbbbf:], 0x7b12)
	p.replaceJump(0xbbc3, p.addCode(patchFinishf660))

	/* Write signal info to safe place (0x7b14) */
	p.replaceJump(0xe9c6, p.addCode(patchFinishSig))

	/* Count frames */
	p.replaceCall(0xb208, p.addCode(patchVSYNC))

	/* Finally, add read results function */
	log.Printf("ReadInfo1 Offset: %02x", p.addCode(patchReadInfo))
	log.Printf("ReadInfo2 Offset: %02x", p.addCode(patchReadInfo2))

	return p.image, nil
}
