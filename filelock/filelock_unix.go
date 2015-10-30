package filelock

import "syscall"

func (fl *filelock) Lock() error {
	fl.mux.Lock()
	defer fl.mux.Unlock()

	err := syscall.Flock(l.fd, syscall.LOCK_EX)
	if err == syscall.EWOULDBLOCK {
		return AlreadyLockedErr
	} else if err != nil {
		return err
	}
	return nil
}

func (fl *filelock) Unlock() error {
	fl.mux.Lock()
	defer fl.mux.Unlock()

	return syscall.Flock(l.fd, syscall.LOCK_UN)
}
