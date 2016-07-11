package main

import (
	"io"
	"sync"
)

type sharedLock struct {
	sync.Locker

	held *struct {
		sync.Locker

		clients int
		locked  bool
	}
}

func newSharedLock(lock sync.Locker, clients int) *sharedLock {
	return &sharedLock{
		Locker: lock,

		held: &struct {
			sync.Locker

			clients int
			locked  bool
		}{
			Locker: &sync.Mutex{},

			clients: clients,
			locked:  false,
		},
	}
}

func (mutex *sharedLock) Lock() {
	mutex.held.Lock()
	defer mutex.held.Unlock()

	if !mutex.held.locked {
		mutex.Locker.Lock()

		mutex.held.locked = true
	}
}

func (mutex *sharedLock) Unlock() {
	mutex.held.Lock()
	defer mutex.held.Unlock()

	mutex.held.clients--

	if mutex.held.clients == 0 && mutex.held.locked {
		mutex.held.locked = false

		mutex.Locker.Unlock()
	}
}

type lockedWriter struct {
	writer io.WriteCloser

	lock sync.Locker
}

func newLockedWriter(
	writer io.WriteCloser,
	lock sync.Locker,
) *lockedWriter {
	return &lockedWriter{
		writer: writer,
		lock:   lock,
	}
}

func (writer *lockedWriter) Write(data []byte) (int, error) {
	writer.lock.Lock()

	return writer.writer.Write(data)
}

func (writer *lockedWriter) Close() error {
	writer.lock.Unlock()

	return writer.writer.Close()
}
