package main

import "encoding/binary"

type jumptableEntry struct {
	Key     uint8
	Address uint16
}

func (p *patcher) jumptableParse(offset uint16) ([]jumptableEntry, uint16) {
	var results []jumptableEntry

	lastKey := -1
	for {
		addr := binary.BigEndian.Uint16(p.image[offset:])
		offset += 2
		if addr > 0 {
			key := int(p.image[offset])
			offset++

			if key < lastKey {
				panic("Parse error")
			}
			lastKey = key

			results = append(results, jumptableEntry{
				Key:     uint8(key),
				Address: addr,
			})
		} else {
			return results, binary.BigEndian.Uint16(p.image[offset:])
		}
	}
}

func (p *patcher) jumptableWrite(offset uint16, entries []jumptableEntry, dflt uint16) {
	for _, m := range entries {
		binary.BigEndian.PutUint16(p.image[offset:], m.Address)
		p.image[offset+2] = m.Key
		offset += 3
	}
	binary.BigEndian.PutUint16(p.image[offset:], 0)
	binary.BigEndian.PutUint16(p.image[offset+2:], dflt)
}
