package model

import (
	"encoding/binary"
	"io"
)

func writeString(wr io.Writer, str string) error {
	buf := []byte(str)
	blen := uint32(len(buf))
	if err := binary.Write(wr, binary.LittleEndian, blen); err != nil {
		return err
	}
	if _, err := wr.Write(buf); err != nil {
		return err
	}
	return nil
}

func readString(rd io.Reader) (str string, err error) {
	var blen uint32
	if err = binary.Read(rd, binary.LittleEndian, &blen); err != nil {
		return
	}
	buf := make([]byte, blen)
	if _, err = rd.Read(buf); err != nil {
		return
	}
	return string(buf), nil
}
