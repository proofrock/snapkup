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
	underlying io.ReadCloser
	key        []byte
	compressed bool
	chunkSize  int
	chunkNum   uint32
	chunk      []byte
	index      int
	finished   bool
}

func NewIS(key []byte, r io.ReadCloser) (*InputStream, error) {
	wannabeMagicNumber := make([]byte, len(magicNumber))
	if _, errReadingMagicNumber := r.Read(wannabeMagicNumber); errReadingMagicNumber != nil {
		return nil, errReadingMagicNumber
	}
	if bytes.Compare(magicNumber, wannabeMagicNumber) != 0 {
		return nil, errors.New("Wrong magic number")
	}
	wannabeMagicNumber = make([]byte, 1)
	if _, errReadingMagicNumber := r.Read(wannabeMagicNumber); errReadingMagicNumber != nil {
		return nil, errReadingMagicNumber
	}
	compressed := bytes.Compare(wannabeMagicNumber, mnCompressed) == 0
	return &InputStream{r, key, compressed, 0, 0, nil, 0, false}, nil
}

func (is *InputStream) unprocess() (finished bool, errDecrypting error) {
	var encSize int64
	if errReadingLen := binary.Read(is.underlying, binary.LittleEndian, &encSize); errReadingLen == io.EOF {
		return true, nil
	} else if errReadingLen != nil {
		return false, errReadingLen
	}

	aead, errAEAD := chacha20poly1305.NewX(is.key)
	if errAEAD != nil {
		return false, errAEAD
	}

	nonce := make([]byte, aead.NonceSize())
	if _, errReadingNonce := is.underlying.Read(nonce); errReadingNonce != nil {
		return false, errReadingNonce
	}

	enc := make([]byte, encSize-int64(aead.NonceSize()))
	if _, errReadingEnc := is.underlying.Read(enc); errReadingEnc != nil {
		return false, errReadingEnc
	}

	compressed, errDecrypting := aead.Open(nil, nonce, enc, uint32ToBytes(is.chunkNum))
	is.chunkNum++
	if errDecrypting != nil {
		return false, errDecrypting
	}

	var uncompressed []byte
	var errDecompressing error
	if is.compressed {
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
