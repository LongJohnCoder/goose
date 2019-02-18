package filesys

import (
	"fmt"
	"path"
	"sync"
)

type fileMode uint8

const (
	readMode = fileMode(iota)
	appendMode
)

func (m fileMode) String() string {
	switch m {
	case readMode:
		return "readMode"
	case appendMode:
		return "appendMode"
	}
	return "invalidMode"
}

type MemFs struct {
	m sync.Mutex
	// fd -> data
	// (note that fds and inodes overlap)
	// (note also that there are no directory fds, so these are all files)
	inodes map[int][]byte
	// filename -> inode
	dirents map[string]int
	// solely for catching misuse we track open files
	openFiles map[int]fileMode
}

func NewMemFs() *MemFs {
	return &MemFs{
		inodes:    make(map[int][]byte),
		dirents:   make(map[string]int),
		openFiles: make(map[int]fileMode),
	}
}

func (fs MemFs) nextFd() int {
	return len(fs.inodes) + 1
}

func (fs *MemFs) Create(fname string) File {
	fs.m.Lock()
	defer fs.m.Unlock()
	fd := fs.nextFd()
	fs.inodes[fd] = nil
	fs.dirents[fname] = fd
	fs.openFiles[fd] = appendMode
	return File(fd)
}

func (fs MemFs) checkMode(f File, mode fileMode) int {
	actual, ok := fs.openFiles[f.fd()]
	if !ok {
		panic(fmt.Errorf("use of unopened file %d", f.fd()))
	}
	if actual != mode {
		panic(fmt.Errorf("attempt to use file using %s != %s", mode, actual))
	}
	return f.fd()
}

func (fs *MemFs) Append(f File, data []byte) {
	fs.m.Lock()
	defer fs.m.Unlock()
	fd := fs.checkMode(f, appendMode)
	fs.inodes[fd] = append(fs.inodes[fd], data...)
}

func (fs *MemFs) Close(f File) {
	fs.m.Lock()
	defer fs.m.Unlock()
	if _, ok := fs.openFiles[f.fd()]; !ok {
		panic(fmt.Errorf("close of unopened fd %d", f.fd()))
	}
	delete(fs.openFiles, f.fd())
}

func (fs *MemFs) Open(fname string) File {
	fs.m.Lock()
	defer fs.m.Unlock()
	fname = path.Clean(fname)
	fd, ok := fs.dirents[fname]
	if !ok {
		panic(fmt.Errorf("file %s does not exist", fname))
	}
	fs.openFiles[fd] = readMode
	return File(fd)
}

func (fs *MemFs) ReadAt(f File, offset uint64, length uint64) []byte {
	fs.m.Lock()
	defer fs.m.Unlock()
	fd := fs.checkMode(f, readMode)
	data := fs.inodes[fd]
	if offset >= uint64(len(data)) {
		return nil
	}
	// copy:
	// (1) makes the returned data independent
	// (2) automatically truncates to the smaller buffer
	p := make([]byte, length)
	n := copy(p, data[offset:])
	return p[:n]
}

func (fs *MemFs) Delete(fname string) {
	fs.m.Lock()
	defer fs.m.Unlock()
	delete(fs.dirents, fname)
	// NOTE: we don't actually garbage collect unreachable files
}

func (fs *MemFs) AtomicCreate(fname string, data []byte) {
	fs.m.Lock()
	defer fs.m.Unlock()
	fd := fs.nextFd()
	p := make([]byte, len(data))
	copy(p, data)
	fs.inodes[fd] = p
	fs.dirents[fname] = fd
}

func (fs *MemFs) List() (names []string) {
	fs.m.Lock()
	defer fs.m.Unlock()
	for n := range fs.dirents {
		names = append(names, n)
	}
	return
}
