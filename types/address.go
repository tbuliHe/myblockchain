package types

import "encoding/hex"

type Address [20]uint8

func (addr Address) ToSlice() []uint8 {
	b := make([]uint8, 20)
	copy(b, addr[:])
	return b
}

func (addr Address) String() string {
	return hex.EncodeToString(addr.ToSlice())
}

func AddressFromBytes(b []uint8) Address {
	if len(b) != 20 {
		panic("invalid address length")
	}
	var value [20]uint8
	copy(value[:], b)
	return Address(value)
}
