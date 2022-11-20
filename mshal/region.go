package mshal

import "errors"

type MemoryRegion interface {
	GetLength() int
	Access(write bool, addr int, buf []byte) (int, error)
	GetParent() (MemoryRegion, int)
	GetName() MemoryRegionNameType
	GetAlignment() int
}

type regionCompleteIO struct {
	MemoryRegion
}

func regionWrapCompleteIO(parent MemoryRegion) MemoryRegion {
	return regionCompleteIO{
		MemoryRegion: parent,
	}
}

func (m regionCompleteIO) Access(write bool, addr int, buf []byte) (int, error) {
	align := m.GetAlignment()
	if addr&(align-1) != 0 {
		return 0, errors.New("address alignment has been violated")
	} else if write && len(buf)%align != 0 {
		return 0, errors.New("data alignment has been violated")
	}

	total := 0
	for len(buf) > 0 {
		n, err := m.MemoryRegion.Access(write, addr+total, buf)
		total += n
		buf = buf[n:]

		if err != nil || n == 0 {
			return total, err
		}
	}

	return total, nil
}

func WriteByte(m MemoryRegion, addr int, value byte) error {
	_, err := m.Access(true, addr, []byte{value})
	return err
}

func ReadByte(m MemoryRegion, addr int) (byte, error) {
	var buf [1]byte
	_, err := m.Access(false, addr, buf[:])
	return buf[0], err
}

type regionPartial struct {
	parent MemoryRegion
	offset int
	length int
	name   MemoryRegionNameType
}

func regionWrapPartial(name MemoryRegionNameType, parent MemoryRegion, offset int, length int) MemoryRegion {
	return regionPartial{
		parent: parent,
		offset: offset,
		length: length,
		name:   name,
	}
}

func (h regionPartial) GetName() MemoryRegionNameType {
	return h.name
}

func (h regionPartial) GetLength() int {
	return h.length
}

func (h regionPartial) GetParent() (MemoryRegion, int) {
	return h.parent, h.offset
}

func (h regionPartial) GetAlignment() int {
	return h.parent.GetAlignment()
}

func (h regionPartial) Access(write bool, addr int, buf []byte) (int, error) {
	if len(buf)+addr > h.length {
		if addr > h.length {
			return 0, nil
		}
		buf = buf[:h.length-addr]
	}

	return h.parent.Access(write, h.offset+addr, buf)
}

func RecursiveGetParentAddress(region MemoryRegion, offset int) (MemoryRegion, int) {
	for {
		var parentOffset int
		prevRegion := region
		region, parentOffset = region.GetParent()

		offset += parentOffset

		if region == nil {
			return prevRegion, offset
		}
	}
}
