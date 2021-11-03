package model

import (
	"bytes"
	"crypto/rand"
	_ "database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/streams"
	"golang.org/x/crypto/argon2"
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
	IsEmpty bool   `json:"isEmpty"`
	Mode    int32  `json:"mode"`
	ModTime int64  `json:"modTime"`
}

type AggloRef struct {
	AggloID string `json:"aggloId"`
	Offset  int64  `json:"offset"`
}

type Blob struct {
	Hash     string    `json:"hash"`
	Size     int64     `json:"size"`
	BlobSize int64     `json:"blobSize"`
	AggloRef *AggloRef `json:"aggloRef"`
}

type Agglo struct {
	ID   string `json:"id"`
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

type Model struct {
	KDFSalt    []byte
	Key4Hashes []byte   `json:"key4hashes"`
	Key4Enc    []byte   `json:"key4enc"`
	Snaps      []Snap   `json:"snaps"`
	Items      []Item   `json:"items"`
	Blobs      []Blob   `json:"blobs"`
	Agglos     []Agglo  `json:"agglos"`
	Roots      []string `json:"roots"`
}

const kdfSaltSize = 16

var magicNumber = []byte("SNPMDL1")

const ModelFileName = "snapkup.dat"

func NewModel() (modl *Model, err error) {
	var ret Model

	ret.KDFSalt = make([]byte, kdfSaltSize)
	if _, err := rand.Read(ret.KDFSalt); err != nil {
		return nil, err
	}

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

func LoadModel(pwd string, dir string) (modl *Model, err error) {
	fPath := path.Join(dir, ModelFileName)

	f, errOpening := os.Open(fPath)
	if errOpening != nil {
		return nil, errOpening
	}
	defer f.Close()

	wannabeMagicNumber := make([]byte, len(magicNumber))
	if _, errReadingMagicNumber := f.Read(wannabeMagicNumber); errReadingMagicNumber != nil {
		return nil, errReadingMagicNumber
	}
	if bytes.Compare(magicNumber, wannabeMagicNumber) != 0 {
		return nil, errors.New("wrong magic number")
	}

	kdfSalt := make([]byte, kdfSaltSize)
	if _, errRdr := f.Read(kdfSalt); errRdr != nil {
		return nil, errRdr
	}
	key := argon2.Key([]byte(pwd), kdfSalt, 3, 32*1024, 4, 32)

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
	ret.KDFSalt = kdfSalt

	return &ret, nil
}

func SaveModel(pwd string, dir string, modl Model) error {
	key := argon2.Key([]byte(pwd), modl.KDFSalt, 3, 32*1024, 4, 32)

	fPath := path.Join(dir, ModelFileName)

	f, err := os.Create(fPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, errWrt := f.Write(magicNumber); errWrt != nil {
		return errWrt
	}
	if _, errWrt := f.Write(modl.KDFSalt); errWrt != nil {
		return errWrt
	}

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
