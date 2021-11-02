package streams

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/DataDog/zstd"
	"golang.org/x/crypto/chacha20poly1305"
)

type InputStream struct {
	underlying io.Reader
	key        []byte
	nonce      []byte
	chunkSize  int
	chunkNum   uint32
	chunk      []byte
	index      int
	finished   bool
}

func NewIS(key []byte, r io.Reader) (*InputStream, error) {
	wannabeMagicNumber := make([]byte, len(magicNumber))
	if _, errReadingMagicNumber := r.Read(wannabeMagicNumber); errReadingMagicNumber != nil {
		return nil, errReadingMagicNumber
	}
	if bytes.Compare(magicNumber, wannabeMagicNumber) != 0 {
		return nil, errors.New("wrong magic number")
	}

	nonce := make([]byte, nonceSize)
	if _, errReadingNonce := r.Read(nonce); errReadingNonce != nil {
		return nil, errReadingNonce
	}

	return &InputStream{r, key, nonce, 0, 0, nil, 0, false}, nil
}

func (is *InputStream) unprocess() (finished bool, errDecrypting error) {
	var encSize int64
	if errReadingLen := binary.Read(is.underlying, binary.LittleEndian, &encSize); errReadingLen == io.EOF {
		return true, nil
	} else if errReadingLen != nil {
		return false, errReadingLen
	}

	z := make([]byte, 1)
	if _, errReadingMagicNumber := is.underlying.Read(z); errReadingMagicNumber != nil {
		return false, errReadingMagicNumber
	}
	isCompressed := z[0] == mnCompressed[0]

	aead, errAEAD := chacha20poly1305.NewX(is.key)
	if errAEAD != nil {
		return false, errAEAD
	}

	derivedNonce := xor(is.nonce, uint32ToBytes(is.chunkNum))
	is.chunkNum++

	enc := make([]byte, encSize-int64(nonceSize))
	if _, errReadingEnc := is.underlying.Read(enc); errReadingEnc != nil {
		return false, errReadingEnc
	}

	compressed, errDecrypting := aead.Open(nil, derivedNonce, enc, nil)
	if errDecrypting != nil {
		return false, errDecrypting
	}

	var uncompressed []byte
	var errDecompressing error
	if isCompressed {
		uncompressed, errDecompressing = zstd.Decompress(nil, compressed)
		if errDecompressing != nil {
			return false, errDecompressing
		}
	} else {
		uncompressed = compressed
	}

	is.chunk = uncompressed
	is.chunkSize = len(uncompressed)
	is.index = 0

	return false, nil
}

func (is *InputStream) Read(p []byte) (int, error) {
	if is.finished {
		return 0, io.EOF
	}

	var i int
	for i = 0; i < len(p); i++ {
		if is.chunkSize == is.index {
			if finished, errUnprocessing := is.unprocess(); errUnprocessing != nil {
				return 0, errUnprocessing
			} else if finished {
				is.finished = true
				return i, nil
			}
		}
		p[i] = is.chunk[is.index]
		is.index++
	}
	return i, nil
}

func (is *InputStream) Close() (err error) {
	if is.finished {
		return nil
	}

	is.finished = true
	return nil
}
