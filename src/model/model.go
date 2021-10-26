package model

import (
	"crypto/rand"
	_ "database/sql"
	"encoding/binary"
	"os"
	"path"

	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/streams"
)

type Snap struct {
	Id        uint32
	Timestamp int64
	Label     string
}

type Item struct {
	Path    string
	Snap    uint32
	Hash    string
	IsDir   bool
	Mode    uint32
	ModTime uint64
}

type Blob struct {
	Hash     string
	Size     uint64
	BlobSize uint64
}

type Root struct {
	Path string
}

type Model struct {
	Key4Hashes []byte
	Key4Enc    []byte
	Snaps      []Snap
	Items      []Item
	Blobs      []Blob
	Roots      []Root
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

	var ret Model

	ret.Key4Enc = make([]byte, 32)
	if _, errRdr := is.Read(ret.Key4Enc); errRdr != nil {
		return nil, errRdr
	}

	ret.Key4Hashes = make([]byte, 32)
	if _, errRdr := is.Read(ret.Key4Hashes); errRdr != nil {
		return nil, errRdr
	}

	var num int32

	if errRdr := binary.Read(is, binary.LittleEndian, &num); errRdr != nil {
		return nil, errRdr
	}
	for i := int32(0); i < num; i++ {
		var itm Snap
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.Id); errRdr != nil {
			return nil, errRdr
		}
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.Timestamp); errRdr != nil {
			return nil, errRdr
		}
		if str, errRdr := readString(is); errRdr != nil {
			return nil, errRdr
		} else {
			itm.Label = str
		}
		ret.Snaps = append(ret.Snaps, itm)
	}

	if errRdr := binary.Read(is, binary.LittleEndian, &num); errRdr != nil {
		return nil, errRdr
	}
	for i := int32(0); i < num; i++ {
		var itm Item
		if str, errRdr := readString(is); errRdr != nil {
			return nil, errRdr
		} else {
			itm.Path = str
		}
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.Snap); errRdr != nil {
			return nil, errRdr
		}
		if str, errRdr := readString(is); errRdr != nil {
			return nil, errRdr
		} else {
			itm.Hash = str
		}
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.IsDir); errRdr != nil {
			return nil, errRdr
		}
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.Mode); errRdr != nil {
			return nil, errRdr
		}
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.ModTime); errRdr != nil {
			return nil, errRdr
		}
		ret.Items = append(ret.Items, itm)
	}

	if errRdr := binary.Read(is, binary.LittleEndian, &num); errRdr != nil {
		return nil, errRdr
	}
	for i := int32(0); i < num; i++ {
		var itm Blob
		if str, errRdr := readString(is); errRdr != nil {
			return nil, errRdr
		} else {
			itm.Hash = str
		}
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.Size); errRdr != nil {
			return nil, errRdr
		}
		if errRdr := binary.Read(is, binary.LittleEndian, &itm.BlobSize); errRdr != nil {
			return nil, errRdr
		}
		ret.Blobs = append(ret.Blobs, itm)
	}

	if errRdr := binary.Read(is, binary.LittleEndian, &num); errRdr != nil {
		return nil, errRdr
	}
	for i := int32(0); i < num; i++ {
		var itm Root
		if str, errRdr := readString(is); errRdr != nil {
			return nil, errRdr
		} else {
			itm.Path = str
		}
		ret.Roots = append(ret.Roots, itm)
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

	if _, errWrt := ous.Write(modl.Key4Enc); errWrt != nil {
		return errWrt
	}

	if _, errWrt := ous.Write(modl.Key4Hashes); errWrt != nil {
		return errWrt
	}

	if errWrt := binary.Write(ous, binary.LittleEndian, int32(len(modl.Snaps))); errWrt != nil {
		return errWrt
	}
	for _, snap := range modl.Snaps {
		if errWrt := binary.Write(ous, binary.LittleEndian, snap.Id); errWrt != nil {
			return errWrt
		}
		if errWrt := binary.Write(ous, binary.LittleEndian, snap.Timestamp); errWrt != nil {
			return errWrt
		}
		if errWrt := writeString(ous, snap.Label); errWrt != nil {
			return errWrt
		}
	}

	if errWrt := binary.Write(ous, binary.LittleEndian, int32(len(modl.Items))); errWrt != nil {
		return errWrt
	}
	for _, item := range modl.Items {
		if errWrt := writeString(ous, item.Path); errWrt != nil {
			return errWrt
		}
		if errWrt := binary.Write(ous, binary.LittleEndian, item.Snap); errWrt != nil {
			return errWrt
		}
		if errWrt := writeString(ous, item.Hash); errWrt != nil {
			return errWrt
		}
		if errWrt := binary.Write(ous, binary.LittleEndian, item.IsDir); errWrt != nil {
			return errWrt
		}
		if errWrt := binary.Write(ous, binary.LittleEndian, item.Mode); errWrt != nil {
			return errWrt
		}
		if errWrt := binary.Write(ous, binary.LittleEndian, item.ModTime); errWrt != nil {
			return errWrt
		}
	}

	if errWrt := binary.Write(ous, binary.LittleEndian, int32(len(modl.Blobs))); errWrt != nil {
		return errWrt
	}
	for _, blob := range modl.Blobs {
		if errWrt := writeString(ous, blob.Hash); errWrt != nil {
			return errWrt
		}
		if errWrt := binary.Write(ous, binary.LittleEndian, blob.Size); errWrt != nil {
			return errWrt
		}
		if errWrt := binary.Write(ous, binary.LittleEndian, blob.BlobSize); errWrt != nil {
			return errWrt
		}
	}

	if errWrt := binary.Write(ous, binary.LittleEndian, int32(len(modl.Roots))); errWrt != nil {
		return errWrt
	}
	for _, root := range modl.Roots {
		if errWrt := writeString(ous, root.Path); errWrt != nil {
			return errWrt
		}
	}

	if errClosing := ous.Close(); errClosing != nil {
		return errClosing
	}

	return nil
}
