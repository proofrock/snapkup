package model

import (
	"os"
	"testing"
)

func TestRW(t *testing.T) {
	key := make([]byte, 32)

	modl, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}

	err = SaveModel(key, ".", *modl)
	defer os.Remove(modelFileName)

	_, err = LoadModel(key, ".")
	if err != nil {
		t.Fatal(err)
	}
}
