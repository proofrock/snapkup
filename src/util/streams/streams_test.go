package streams

import (
	"bytes"
	"crypto/rand"
	"io"
	rnd "math/rand"
	"os"
	"testing"
	"time"
)

func init() {
	println("Init random seed")
	rnd.Seed(time.Now().Unix())
}

const MEGA = 1024 * 1024

func TestRW(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	dataLen := 33 * MEGA

	data := make([]byte, dataLen)
	rnd.Read(data)

	defer func() {
		os.Remove("ciabo")
	}()

	func() {
		f, _ := os.Create("ciabo")
		defer f.Close()

		ous, err := NewOS(key, 16*MEGA, false, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := ous.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = ous.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}()

	data2 := make([]byte, dataLen, dataLen)

	func() {
		f, _ := os.Open("ciabo")
		defer f.Close()

		is, err := NewIS(key, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := is.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = is.Read(data2)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if bytes.Compare(data, data2) != 0 {
		t.Fatal("Data are different")
	}
}

func TestNoRandom(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	dataLen := 33 * MEGA

	data := make([]byte, dataLen)

	defer func() {
		os.Remove("ciano")
	}()

	func() {
		f, _ := os.Create("ciano")
		defer f.Close()

		ous, err := NewOS(key, 16*MEGA, false, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := ous.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = ous.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}()

	data2 := make([]byte, dataLen, dataLen)

	func() {
		f, _ := os.Open("ciano")
		defer f.Close()

		is, err := NewIS(key, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := is.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = is.Read(data2)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if bytes.Compare(data, data2) != 0 {
		t.Fatal("Data are different")
	}
}

func TestNoZ(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	dataLen := 33 * MEGA

	data := make([]byte, dataLen)
	rnd.Read(data)

	defer func() {
		os.Remove("ciaco")
	}()

	func() {
		f, _ := os.Create("ciaco")
		defer f.Close()

		ous, err := NewOS(key, 16*MEGA, true, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := ous.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = ous.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}()

	data2 := make([]byte, dataLen, dataLen)

	func() {
		f, _ := os.Open("ciaco")
		defer f.Close()

		is, err := NewIS(key, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := is.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = is.Read(data2)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if bytes.Compare(data, data2) != 0 {
		t.Fatal("Data are different")
	}
}

func TestEOF(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	defer func() {
		os.Remove("ciaao")
	}()

	func() {
		dataLen := 2050
		data := make([]byte, dataLen)
		rnd.Read(data)

		f, _ := os.Create("ciaao")
		defer f.Close()

		os, err := NewOS(key, 1020, false, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := os.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = os.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}()

	func() {
		dataLen := 2000
		data2 := make([]byte, dataLen, dataLen)

		f, _ := os.Open("ciaao")
		defer f.Close()

		is, err := NewIS(key, f)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := is.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		n, err := is.Read(data2)
		if err != nil {
			t.Fatal(err)
		}
		if n != dataLen {
			t.Fatal("First swipe should be", dataLen)
		}

		n, err = is.Read(data2)
		if err != nil {
			t.Fatal(err)
		}
		if n != 50 {
			t.Fatal("Second swipe should be", 50)
		}

		n, err = is.Read(data2)
		if err != io.EOF {
			t.Fatal("Data should be finished")
		}
	}()
}
