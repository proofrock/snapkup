package snap

import (
	"encoding/hex"
	"io"
	"os"
	"strings"

	"github.com/dchest/siphash"
	"github.com/proofrock/snapkup/model"
)

func findSnap(modl *model.Model, snap int) int {
	for i, snp := range modl.Snaps {
		if snp.Id == snap {
			return i
		}
	}

	return -1
}

const bufSize = 1024 * 32 // 32Kb

func fileHash(path string, key []byte) (string, error) {
	source, errOpening := os.Open(path)
	if errOpening != nil {
		return "", errOpening
	}
	defer source.Close()

	hasher := siphash.New128(key)
	buf := make([]byte, bufSize)
	for {
		n, errHashingFile := source.Read(buf)
		if errHashingFile != nil && errHashingFile != io.EOF {
			return "", errHashingFile
		}
		if n == 0 {
			break
		}

		if _, errWritingHash := hasher.Write(buf[:n]); errWritingHash != nil {
			return "", errWritingHash
		}
	}

	ret := hasher.Sum(nil)

	return strings.ToLower(hex.EncodeToString(ret)), nil
}
