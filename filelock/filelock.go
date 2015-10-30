package filelock

import (
	"errors"
	"os"
	"sync"
)

var AlreadyLockedErr = errors.New("already locked")

type FileLock interface {
	Name() string
	Lock() error
	Unlock() error
	Remove() error
}

type filelock struct {
	fd  int
	f   *os.File
	mux *sync.Mutex
}

func New(name string) (FileLock, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	flock := &filelock{
		int(f.Fd()),
		f,
	}
	return flock, nil
}

func (fl *filelock) Name() string {
	return l.f.Name()
}

func (fl *filelock) Remove() error {
	fl.mux.Lock()
	defer fl.mux.Unlock()

	err := l.f.Close()
	if err != nil {
		return err
	}
	return os.Remove(l.Name())
}
