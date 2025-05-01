package store

import (
	"math/rand/v2"
	"strconv"
	"time"
)

const (
	KB       = 1000
	MB       = 1000 * KB
	GB       = 1000 * MB
	MAX_SIZE = 4 * GB

	CODE_LEN   = 4
	EXPIRATION = time.Minute * 5
)

type Store struct {
	files []File
}

type File struct {
	Name     string
	ID       string
	Size     int
	Uploaded time.Time
	Expires  time.Time
}

func New() *Store {
	return &Store{}
}

// Saves file to disk temporarily and returns id/code to display to user.
func (m *Store) SaveFile(filename string, size int) string {
	id := m.newId()
	m.files = append(m.files, File{
		Name:     filename,
		ID:       id,
		Size:     size,
		Uploaded: time.Now(),
		Expires:  time.Now().Add(EXPIRATION),
	})

	return id
}

func (m *Store) newId() string {
	s := strconv.Itoa(len(m.files))
	for i := len(s); i < CODE_LEN; i++ {
		s += strconv.Itoa(rand.IntN(10))
	}

	return s
}
