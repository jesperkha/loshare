package store

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/jesperkha/loshare/config"
	"github.com/jesperkha/notifier"
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
	files  []File
	config *config.Config
}

type File struct {
	Name     string
	ID       string
	Size     int
	Uploaded time.Time
	Expires  time.Time
}

func New(config *config.Config) *Store {
	return &Store{config: config}
}

// Saves file to disk temporarily and returns id/code to display to user.
func (s *Store) SaveFile(filename string, size int) (string, error) {
	if size > MAX_SIZE {
		return "", fmt.Errorf("rejected '%s': file was too large (%d bytes)", filename, size)
	}

	id := s.newId()
	s.files = append(s.files, File{
		Name:     filename,
		ID:       id,
		Size:     size,
		Uploaded: time.Now(),
		Expires:  time.Now().Add(EXPIRATION),
	})

	return id, nil
}

func (s *Store) newId() string {
	id := strconv.Itoa(len(s.files))
	for i := len(id); i < CODE_LEN; i++ {
		id += strconv.Itoa(rand.IntN(10))
	}

	return id
}

func (s *Store) Run(notif *notifier.Notifier) {
	done, finish := notif.Register()

	<-done
	finish()
}
