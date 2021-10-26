package streams

import "encoding/binary"

func uint32ToBytes(i uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, i)
	return bs
}

const zLevel = 19

var magicNumber []byte = []byte("SNP1")

var mnCompressed []byte = []byte("Z")
var mnUncompressed []byte = []byte("N")
