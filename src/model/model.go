package model

import (
	"crypto/rand"
	_ "database/sql"
	"encoding/binary"
	"encoding/json"
	"os"
	"path"

	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/streams"
)

type Snap struct {
	Id        int    `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Label     string `json:"label"`
}

type Item struct {
	Path    string `json:"path"`
	Snap    int    `json:"snap"`
	Hash    string `json:"hash"`
	IsDir   bool   `json:"isDir"`
	Mode    int32  `json:"mode"`
	ModTime int64  `json:"modTime"`
}

type Blob struct {
	Hash     string `json:"hash"`
	Size     int64  `json:"size"`
	BlobSize int64  `json:"blobSize"`
}

type Root struct {
	Path string `json:"path"`
}

type Model struct {
	Key4Hashes []byte `json:"key4hashes"`
	Key4Enc    []byte `json:"key4enc"`
	Snaps      []Snap `json:"snaps"`
	Items      []Item `json:"items"`
	Blobs      []Blob `json:"blobs"`
	Roots      []Root `json:"roots"`
}

const modelFileName = "snapkup.dat"

func NewModel() (modl *Model, err error) {
	var ret Model
	ret.Key4Enc = make([]byte, 32)
	if _, err := rand.Read(ret.Key4Enc); err != nil {
		return nil, err
	}

	ret.Key4Hashes = make([]byte, 32)
	if _, err := rand.Read(ret.Key4Hashes); err != nil {
		return nil, err
	}

	return &ret, nil
}

func LoadModel(key []byte, dir string) (modl *Model, err error) {
	fPath := path.Join(dir, modelFileName)

	f, errOpening := os.Open(fPath)
	if errOpening != nil {
		return nil, errOpening
	}
	defer f.Close()

	is, errOpeningIS := streams.NewIS(key, f)
	if errOpeningIS != nil {
		return nil, errOpeningIS
	}
	defer is.Close()

	var lenMarshaled int32
	if errRdr := binary.Read(is, binary.LittleEndian, &lenMarshaled); errRdr != nil {
		return nil, errRdr
	}
	marshaled := make([]byte, lenMarshaled)
	if _, errRdr := is.Read(marshaled); errRdr != nil {
		return nil, errRdr
	}

	var ret Model
	if errUnmarshaling := json.Unmarshal(marshaled, &ret); errUnmarshaling != nil {
		return nil, errUnmarshaling
	}

	return &ret, nil
}

func SaveModel(key []byte, dir string, modl Model) error {
	fPath := path.Join(dir, modelFileName)

	f, err := os.Create(fPath)
	if err != nil {
		return err
	}
	defer f.Close()

	ous, errOpeningOS := streams.NewOS(key, util.ChunkSize, false, f)
	if errOpeningOS != nil {
		return errOpeningOS
	}
	defer ous.Close()

	marshaled, errMarshaling := json.Marshal(modl)
	if errMarshaling != nil {
		return errMarshaling
	}

	if errWrt := binary.Write(ous, binary.LittleEndian, int32(len(marshaled))); errWrt != nil {
		return errWrt
	}
	if _, errWrt := ous.Write(marshaled); errWrt != nil {
		return errWrt
	}

	if errClosing := ous.Close(); errClosing != nil {
		return errClosing
	}

	return nil
}
