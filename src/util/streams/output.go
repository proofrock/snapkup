package streams

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"github.com/DataDog/zstd"
	"golang.org/x/crypto/chacha20poly1305"
)

type OutputStream struct {
	underlying      io.Writer
	key             []byte
	nonce           []byte
	previousAuthTag []byte
	neverCompress   bool
	chunkSize       int
	chunkNum        uint32
	chunk           []byte
	index           int
	finished        bool
}

func NewOS(key []byte, chunkSize int, neverCompress bool, w io.Writer) (*OutputStream, error) {
	if _, errWritingMagicNumber := w.Write(magicNumber); errWritingMagicNumber != nil {
		return nil, errWritingMagicNumber
	}

	nonce := make([]byte, nonceSize)
	if _, errGenNonce := rand.Read(nonce); errGenNonce != nil {
		return nil, errGenNonce
	}

	if _, errWritingNonce := w.Write(nonce); errWritingNonce != nil {
		return nil, errWritingNonce
	}

	return &OutputStream{w, key, nonce, nil, neverCompress, chunkSize, 0, make([]byte, chunkSize), 0, false}, nil
}

func (os *OutputStream) process() error {
	var compressed []byte
	var isCompressed []byte
	if os.neverCompress {
		compressed = os.chunk[:os.index]
		isCompressed = mnUncompressed
	} else {
		var errCompressing error
		compressed, errCompressing = zstd.CompressLevel(nil, os.chunk[:os.index], zLevel)
		if errCompressing != nil {
			return errCompressing
		}
		if len(compressed) >= os.index {
			compressed = os.chunk[:os.index]
			isCompressed = mnUncompressed
		} else {
			isCompressed = mnCompressed
		}
	}

	aead, errAEAD := chacha20poly1305.NewX(os.key)
	if errAEAD != nil {
		return errAEAD
	}

	if os.previousAuthTag == nil {
		os.previousAuthTag = make([]byte, aead.Overhead())
	}

	derivedNonce := xor(os.nonce, uint32ToBytes(os.chunkNum))
	os.chunkNum++

	encLen := len(compressed) + aead.Overhead()
	encrypted := make([]byte, encLen)

	encrypted = aead.Seal(nil, derivedNonce, compressed, os.previousAuthTag)
	if errWritingLen := binary.Write(os.underlying, binary.LittleEndian, int64(encLen+nonceSize)); errWritingLen != nil {
		return errWritingLen
	}
	if _, errWritingZ := os.underlying.Write(isCompressed); errWritingZ != nil {
		return errWritingZ
	}
	if _, errWritingData := os.underlying.Write(encrypted); errWritingData != nil {
		return errWritingData
	}
	os.previousAuthTag = encrypted[len(encrypted)-aead.Overhead():]
	return nil
}

func (os *OutputStream) Write(p []byte) (n int, err error) {
	if os.finished {
		return 0, errors.New("stream already closed")
	}

	for i := 0; i < len(p); i++ {
		os.chunk[os.index] = p[i]
		os.index++
		if os.index == os.chunkSize {
			if errProcessing := os.process(); errProcessing != nil {
				return 0, errProcessing
			}
			os.index = 0
		}
	}

	return len(p), nil
}

func (os *OutputStream) Close() (err error) {
	if os.finished {
		return nil
	}

	if os.index > 0 {
		if errProcessing := os.process(); errProcessing != nil {
			return errProcessing
		}
	}

	os.finished = true

	return nil
}
