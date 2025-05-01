package store

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path"
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
	files  map[string]File
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
	return &Store{
		config: config,
		files:  make(map[string]File),
	}
}

// Saves file to disk temporarily and returns id/code to display to user.
func (s *Store) SaveFile(filename string, size int, r io.Reader) (string, error) {
	if size > MAX_SIZE {
		return "", fmt.Errorf("rejected '%s': file was too large (%d bytes)", filename, size)
	}

	id := s.newId()
	file := File{
		Name:     filename,
		ID:       id,
		Size:     size,
		Uploaded: time.Now(),
		Expires:  time.Now().Add(EXPIRATION),
	}

	if err := writeFile(s.path(file), r); err != nil {
		return "", err
	}

	// Collision can happen if a file is uploaded at the same time as another
	// is deleted and the n random numbers appended are identical.
	s.files[id] = file
	return id, nil
}

func (s *Store) GetFile(id string) (filename string, r io.ReadCloser, err error) {
	file, ok := s.files[id]
	if !ok {
		return "", nil, fmt.Errorf("no file with id=%s", id)
	}

	r, err = readFile(s.path(file))
	return file.Name, r, err
}

func (s *Store) newId() string {
	id := strconv.Itoa(len(s.files))
	for i := len(id); i < CODE_LEN; i++ {
		id += strconv.Itoa(rand.IntN(10))
	}

	return id
}

func (s *Store) path(f File) string {
	return path.Join(s.config.DumpDir, f.ID)
}

func (s *Store) Run(notif *notifier.Notifier) {
	done, finish := notif.Register()

	<-done
	finish()
}

func writeFile(path string, r io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func readFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
