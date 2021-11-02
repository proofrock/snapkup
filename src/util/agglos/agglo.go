package agglos

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"io"
	"math/rand"
	"os"
	"path"
	"strings"
)

func Plan(modl *model.Model, treshold, target int64) (agglos map[string][]string, blobs map[string]model.AggloRef, err error) {
	agglos = make(map[string][]string)
	blobs = make(map[string]model.AggloRef)

	var toAgglo []model.Blob
	for _, blob := range modl.Blobs {
		if blob.AggloRef == nil && blob.BlobSize <= treshold {
			toAgglo = append(toAgglo, blob)
		}
	}
	if len(toAgglo) == 0 {
		return nil, nil, nil
	}

	// FIXME non-deterministic
	rand.Shuffle(len(toAgglo), func(i, j int) { toAgglo[i], toAgglo[j] = toAgglo[j], toAgglo[i] })

	aggloId := genAggloId(16)
	var blobsForAgglo []string
	var offset int64 = 0
	for _, blob := range toAgglo {
		blobs[blob.Hash] = model.AggloRef{AggloID: aggloId, Offset: offset}
		blobsForAgglo = append(blobsForAgglo, blob.Hash)
		offset += blob.BlobSize
		if offset > target {
			agglos[aggloId] = blobsForAgglo
			blobsForAgglo = make([]string, 0)
			aggloId = genAggloId(16)
			offset = 0
		}
	}
	if len(blobsForAgglo) > 0 {
		agglos[aggloId] = blobsForAgglo
	}

	return
}

var magicNumber = []byte("SNP+1")

func Apply(modl *model.Model, bkpDir string, agglos map[string][]string, blobs map[string]model.AggloRef) error {
	for aggloId, blobsForAgglo := range agglos {
		dstPath := path.Join(bkpDir, aggloId[1:2], aggloId)
		dst, errCreatingAgglo := os.Create(dstPath)
		if errCreatingAgglo != nil {
			return errCreatingAgglo
		}
		defer dst.Close()

		if _, errWrt := dst.Write(magicNumber); errWrt != nil {
			return errWrt
		}

		for _, blobToAdd := range blobsForAgglo {
			src, errOpeningBlob := os.Open(path.Join(bkpDir, blobToAdd[0:1], blobToAdd))
			if errOpeningBlob != nil {
				return errOpeningBlob
			}
			defer src.Close()
			if _, errCopying := io.Copy(dst, src); errCopying != nil {
				return errCopying
			}
			if errClosingSrc := src.Close(); errClosingSrc != nil {
				return errClosingSrc
			}
		}

		if errClosingDest := dst.Close(); errClosingDest != nil {
			return errClosingDest
		}

		hash, errHashing := util.FileHash(dstPath, modl.Key4Hashes)
		if errHashing != nil {
			return errHashing
		}

		stats, errStatsing := os.Stat(dstPath)
		if errStatsing != nil {
			return errStatsing
		}

		modl.Agglos = append(modl.Agglos, model.Agglo{ID: aggloId, Hash: hash, Size: stats.Size()})
	}

	for _, blobsForAgglo := range agglos {
		for _, blobToDel := range blobsForAgglo {
			fpBTD := path.Join(bkpDir, blobToDel[0:1], blobToDel)
			if errDeleting := os.Remove(fpBTD); errDeleting != nil {
				fmt.Fprintf(os.Stderr, "ERROR: deleting file %s; %v\n", fpBTD, errDeleting)
			}
		}
	}

	for idx, blob := range modl.Blobs {
		if nuAgglo, isThere := blobs[blob.Hash]; isThere {
			modl.Blobs[idx].AggloRef = &nuAgglo
		}
	}

	return nil
}

type AggloInputStream struct {
	underlying io.Reader
	offset     int64
	size       int64
	read       int64
}

func NewAIS(offset, size int64, r io.ReadSeeker) (*AggloInputStream, error) {
	wannabeMagicNumber := make([]byte, len(magicNumber))
	if _, errReadingMagicNumber := r.Read(wannabeMagicNumber); errReadingMagicNumber != nil {
		return nil, errReadingMagicNumber
	}
	if bytes.Compare(magicNumber, wannabeMagicNumber) != 0 {
		return nil, errors.New("wrong magic number")
	}

	toSeek := offset + int64(len(magicNumber))
	seekd, errSeeking := r.Seek(toSeek, io.SeekStart)
	if errSeeking != nil {
		return nil, errSeeking
	} else if seekd != toSeek {
		return nil, errors.New("agglo file too short")
	}

	return &AggloInputStream{r, offset, size, 0}, nil
}

func (ais *AggloInputStream) Read(p []byte) (int, error) {
	remaining := ais.size - ais.read
	if remaining <= 0 {
		return 0, io.EOF
	}

	read, errReading := ais.underlying.Read(p)
	if errReading != nil {
		return 0, errReading
	}

	if int64(read) <= remaining {
		ais.read += int64(read)
		return read, nil
	} else {
		ais.read = ais.size
		return int(remaining), nil
	}
}

func SplitAll(modl *model.Model, bkpDir string) error {
	for idx, blob := range modl.Blobs {
		if blob.AggloRef == nil {
			continue
		}

		aggloRef := *blob.AggloRef

		src, errOpening := os.Open(path.Join(bkpDir, aggloRef.AggloID[1:2], aggloRef.AggloID))
		if errOpening != nil {
			return errOpening
		}
		defer src.Close()

		srcPiece, errOpeningPieceReader := NewAIS(aggloRef.Offset, blob.BlobSize, src)
		if errOpeningPieceReader != nil {
			return errOpeningPieceReader
		}

		dst, errCreating := os.Create(path.Join(bkpDir, blob.Hash[0:1], blob.Hash))
		if errCreating != nil {
			return errCreating
		}
		defer dst.Close()

		if _, errCopying := io.Copy(dst, srcPiece); errCopying != nil {
			return errCopying
		}

		// TODO verify file

		modl.Blobs[idx].AggloRef = nil

		if errClosing := src.Close(); errClosing != nil {
			return errClosing
		}
	}

	for _, agglo := range modl.Agglos {
		if errDeleting := os.Remove(path.Join(bkpDir, agglo.ID[1:2], agglo.ID)); errDeleting != nil {
			fmt.Fprintf(os.Stderr, "ERROR: deleting file %s; %v\n", agglo.ID, errDeleting)
		}
	}

	modl.Agglos = make([]model.Agglo, 0)

	return nil
}

func genAggloId(lenBytes int) string {
	bytess := make([]byte, lenBytes)
	rand.Read(bytess)
	return "+" + strings.ToLower(hex.EncodeToString(bytess))
}
