package transport

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	SharedMemMagic       = "V2E-SHRM"
	SharedMemVersion     = 1
	SharedMemMinSize     = 4096
	SharedMemDefaultSize = 64 * 1024        // 64KB default
	SharedMemMaxSize     = 16 * 1024 * 1024 // 16MB max
)

type SharedMemHeader struct {
	Magic    [8]byte
	Version  uint16
	Flags    uint16
	Reserved uint32
	WritePos uint32
	ReadPos  uint32
	Size     uint32
	Capacity uint32
}

type SharedMemory struct {
	fd       int
	data     []byte
	header   *SharedMemHeader
	mu       sync.Mutex
	closed   bool
	isServer bool
	memFd    *os.File
}

type SharedMemConfig struct {
	Size     uint32
	IsServer bool
}

func NewSharedMemory(config SharedMemConfig) (*SharedMemory, error) {
	if config.Size < SharedMemMinSize {
		config.Size = SharedMemDefaultSize
	}
	if config.Size > SharedMemMaxSize {
		config.Size = SharedMemMaxSize
	}

	size := int(alignToPage(uint64(config.Size)))

	fd, err := unix.MemfdCreate("v2e-shmem", unix.MFD_CLOEXEC)
	if err != nil {
		return nil, fmt.Errorf("failed to create memfd: %w", err)
	}

	if err := unix.Ftruncate(fd, int64(size)); err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("failed to set memfd size: %w", err)
	}

	data, err := unix.Mmap(fd, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("failed to mmap shared memory: %w", err)
	}

	shm := &SharedMemory{
		fd:       fd,
		data:     data,
		isServer: config.IsServer,
		memFd:    os.NewFile(uintptr(fd), "v2e-shmem"),
	}

	shm.header = (*SharedMemHeader)(unsafe.Pointer(&data[0]))

	if config.IsServer {
		shm.initializeHeader(size)
	} else {
		if err := shm.validateHeader(); err != nil {
			shm.Close()
			return nil, fmt.Errorf("invalid shared memory header: %w", err)
		}
	}

	return shm, nil
}

func (shm *SharedMemory) initializeHeader(size int) {
	copy(shm.header.Magic[:], SharedMemMagic)
	shm.header.Version = SharedMemVersion
	shm.header.Size = uint32(size)
	shm.header.Capacity = uint32(size - int(unsafe.Sizeof(SharedMemHeader{})))
	shm.header.WritePos = 0
	shm.header.ReadPos = 0
}

func (shm *SharedMemory) validateHeader() error {
	magic := string(shm.header.Magic[:])
	if magic != SharedMemMagic {
		return fmt.Errorf("invalid magic: got %s, want %s", magic, SharedMemMagic)
	}
	if shm.header.Version != SharedMemVersion {
		return fmt.Errorf("unsupported version: %d", shm.header.Version)
	}
	return nil
}

func (shm *SharedMemory) Write(data []byte) error {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	if shm.closed {
		return fmt.Errorf("shared memory closed")
	}

	if len(data) == 0 {
		return nil
	}

	if uint32(len(data)) > shm.header.Capacity {
		return fmt.Errorf("data too large: %d bytes", len(data))
	}

	headerSize := uint32(unsafe.Sizeof(SharedMemHeader{}))
	writePos := shm.header.WritePos + headerSize
	remaining := shm.header.Capacity - (shm.header.WritePos % shm.header.Capacity)

	if uint32(len(data)) > remaining {
		return fmt.Errorf("ring buffer full")
	}

	copy(shm.data[writePos:writePos+uint32(len(data))], data)
	shm.header.WritePos += uint32(len(data))

	return nil
}

func (shm *SharedMemory) Read(dst []byte) (int, error) {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	if shm.closed {
		return 0, fmt.Errorf("shared memory closed")
	}

	if shm.header.ReadPos == shm.header.WritePos {
		return 0, nil
	}

	headerSize := uint32(unsafe.Sizeof(SharedMemHeader{}))
	readPos := shm.header.ReadPos + headerSize
	available := shm.header.WritePos - shm.header.ReadPos

	if uint32(len(dst)) > available {
		return 0, fmt.Errorf("not enough data available")
	}

	copy(dst, shm.data[readPos:readPos+uint32(len(dst))])
	shm.header.ReadPos += uint32(len(dst))

	if shm.header.ReadPos == shm.header.WritePos {
		shm.header.ReadPos = 0
		shm.header.WritePos = 0
	}

	return len(dst), nil
}

func (shm *SharedMemory) IsClosed() bool {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	return shm.closed
}

func (shm *SharedMemory) Available() uint32 {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	return shm.header.Capacity - (shm.header.WritePos - shm.header.ReadPos)
}

func (shm *SharedMemory) BytesAvailable() uint32 {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	return shm.header.WritePos - shm.header.ReadPos
}

func (shm *SharedMemory) Fd() uintptr {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	return uintptr(shm.fd)
}

func (shm *SharedMemory) Size() uint32 {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	return shm.header.Size
}

func (shm *SharedMemory) Close() error {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	if shm.closed {
		return nil
	}

	shm.closed = true

	var err error
	if shm.data != nil {
		if unmapErr := unix.Munmap(shm.data); unmapErr != nil && err == nil {
			err = unmapErr
		}
		shm.data = nil
	}

	if shm.fd >= 0 {
		if closeErr := unix.Close(shm.fd); closeErr != nil && err == nil {
			err = closeErr
		}
		shm.fd = -1
	}

	return err
}

func (shm *SharedMemory) SendFd(conn *os.File) error {
	return fmt.Errorf("SendFd: fd passing not implemented")
}

func alignToPage(size uint64) uint64 {
	pageSize := uint64(os.Getpagesize())
	return ((size + pageSize - 1) / pageSize) * pageSize
}
