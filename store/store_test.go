package store

import (
	"io"
	"testing"

	"github.com/jesperkha/loshare/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Port:    "80",
		DumpDir: ".dump",
	}
}

func TestMaxFileSize(t *testing.T) {
	store := New(testConfig())

	_, err := store.SaveFile("foo", MAX_SIZE+1, nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestMaxStoreSize(t *testing.T) {
	store := New(testConfig())

	size := MAX_STORE_SIZE / 2
	store.totalSize = size

	_, err := store.SaveFile("foo", size+1, nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestIdGen(t *testing.T) {
	m := map[string]struct{}{}
	store := New(testConfig())

	// Mock writer for test
	store.fileWriter = func(path string, r io.Reader) error {
		return nil
	}

	// Duplicate id in as little as 100 draws should be improbable enough to test
	for range 100 {
		id, err := store.SaveFile("", 0, nil)
		if err != nil {
			t.Error(err)
		}

		if _, ok := m[id]; ok {
			t.Errorf("duplicate id=%s", id)
		}

		m[id] = struct{}{}
	}
}
