package model

import (
	"os"
	"testing"
)

func TestRW(t *testing.T) {
	modl, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}

	err = SaveModel("key", ".", *modl)
	defer os.Remove(ModelFileName)

	_, err = LoadModel("key", ".")
	if err != nil {
		t.Fatal(err)
	}
}
