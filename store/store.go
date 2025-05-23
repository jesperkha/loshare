package store

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/jesperkha/loshare/config"
	"github.com/jesperkha/notifier"
)

const (
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB

	// Max size of uploaded files in bytes
	MAX_SIZE = 2 * GB

	// Length of file codes
	CODE_LEN = 6

	// How long a file is kept before deletion
	EXPIRATION = time.Minute * 10

	// How often to poll and purge expired files
	REFRESH_RATE = time.Minute

	// Maxiumum size of temporary store before rejecting files
	MAX_STORE_SIZE = 4 * GB
)

type Store struct {
	files     map[string]File
	config    *config.Config
	totalSize int

	// Use generic writer and reader function to allow easy swapping
	fileWriter func(path string, r io.Reader) error
	fileReader func(path string) (io.ReadCloser, error)
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
		config:     config,
		files:      make(map[string]File),
		fileWriter: writeFile,
		fileReader: readFile,
	}
}

// Initialize store, clearing and creating dump directory.
func (s *Store) Init() {
	os.RemoveAll(s.config.DumpDir)
	os.Mkdir(s.config.DumpDir, os.ModePerm)
}

// Saves file to disk temporarily and returns id/code to display to user.
func (s *Store) SaveFile(filename string, size int, r io.Reader) (string, error) {
	if size > MAX_SIZE {
		return "", fmt.Errorf("file is too large: %s", filename)
	}

	if s.totalSize+size >= MAX_STORE_SIZE {
		return "", fmt.Errorf("file storage is full or file would overflow cap: %s", filename)
	}

	id := s.newId()
	file := File{
		Name:     filename,
		ID:       id,
		Size:     size,
		Uploaded: time.Now(),
		Expires:  time.Now().Add(EXPIRATION),
	}

	if err := s.fileWriter(s.path(file), r); err != nil {
		return "", err
	}

	// Collision can happen if a file is uploaded at the same time as another
	// is deleted and the n random numbers appended are identical.
	s.files[id] = file
	s.totalSize += file.Size
	return id, nil
}

func (s *Store) GetFile(id string) (filename string, r io.ReadCloser, err error) {
	file, ok := s.files[id]
	if !ok {
		return "", nil, fmt.Errorf("no file with id=%s", id)
	}

	r, err = s.fileReader(s.path(file))
	return file.Name, r, err
}

// Run starts a background process to remove expired files.
func (s *Store) Run(notif *notifier.Notifier) {
	done, finish := notif.Register()
	tick := time.NewTicker(REFRESH_RATE)

	for {
		select {
		case <-tick.C:
			if err := s.removeExpiredFiles(); err != nil {
				log.Println(err)
			}

		case <-done:
			finish()
			return
		}
	}
}

func (s *Store) removeExpiredFiles() error {
	for _, file := range s.files {
		if file.Expires.Before(time.Now()) {
			if err := os.Remove(s.path(file)); err != nil {
				return err
			}

			delete(s.files, file.ID)
			log.Printf("removed file id=%s", file.ID)
		}
	}

	return nil
}

func (s *Store) newId() string {
	id := ""
	for range CODE_LEN {
		id += strconv.Itoa(rand.IntN(10))
	}
	return id
}

func (s *Store) path(f File) string {
	return path.Join(s.config.DumpDir, f.ID)
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
