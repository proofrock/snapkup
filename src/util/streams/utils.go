package streams

import (
	"encoding/binary"
)

const nonceSize = 24

const zLevel = 19

var magicNumber = []byte("SNP1")

var mnCompressed = []byte("Z")
var mnUncompressed = []byte("N")

func uint32ToBytes(i uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, i)
	return bs
}

func xor(b1, b2 []byte) []byte {
	if len(b1) < len(b2) {
		b1, b2 = b2, b1
	}
	ret := make([]byte, len(b1))
	for i := 0; i < len(ret); i++ {
		if i < len(b2) {
			ret[i] = b1[i] ^ b2[i]
		} else {
			ret[i] = b1[i]
		}
	}
	return ret
}
