package streams

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"github.com/DataDog/zstd"
	"golang.org/x/crypto/chacha20poly1305"
)

var magicNumber []byte = []byte("SNP1")

var mnCompressed []byte = []byte("Z")
var mnUncompressed []byte = []byte("N")

type OutputStream struct {
	underlying io.Writer
	key        []byte
	chunkSize  int
	zLevel     int
	chunkNum   uint32
	chunk      []byte
	index      int
	finished   bool
}

func NewOS(key []byte, chunkSize int, compressed bool, w io.Writer) (*OutputStream, error) {
	if _, errWritingMagicNumber := w.Write(magicNumber); errWritingMagicNumber != nil {
		return nil, errWritingMagicNumber
	}

	var zmg []byte
	var zLevel int
	if compressed {
		zmg = mnCompressed
		zLevel = 19
	} else {
		zmg = mnUncompressed
		zLevel = -1
	}

	if _, errWritingMagicNumber := w.Write(zmg); errWritingMagicNumber != nil {
		return nil, errWritingMagicNumber
	}

	return &OutputStream{w, key, chunkSize, zLevel, 0, make([]byte, chunkSize), 0, false}, nil
}

func (os *OutputStream) process() error {
	var compressed []byte
	if os.zLevel >= 0 {
		var errCompressing error
		compressed, errCompressing = zstd.CompressLevel(nil, os.chunk[:os.index], os.zLevel)
		if errCompressing != nil {
			return errCompressing
		}
	} else {
		compressed = os.chunk[:os.index]
	}

	aead, errAEAD := chacha20poly1305.NewX(os.key)
	if errAEAD != nil {
		return errAEAD
	}
	nonce := make([]byte, aead.NonceSize())
	if _, errGenIV := rand.Read(nonce); errGenIV != nil {
		return errGenIV
	}

	encLen := len(compressed) + aead.Overhead()
	encrypted := make([]byte, encLen)

	encrypted = aead.Seal(nil, nonce, compressed, uint32ToBytes(os.chunkNum))
	os.chunkNum++
	if errWritingLen := binary.Write(os.underlying, binary.LittleEndian, int64(encLen+len(nonce))); errWritingLen != nil {
		return errWritingLen
	}
	if _, errWritingNonce := os.underlying.Write(nonce); errWritingNonce != nil {
		return errWritingNonce
	}
	if _, errWritingData := os.underlying.Write(encrypted); errWritingData != nil {
		return errWritingData
	}
	return nil
}

func (os *OutputStream) Write(p []byte) (n int, err error) {
	if os.finished {
		return 0, errors.New("Stream already closed")
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
