package transport

import (
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	SharedMemMagic   = "V2E-SHRM"
	SharedMemVersion = 1
	SharedMemMinSize     = 4096
	SharedMemDefaultSize = 64 * 1024
	SharedMemMaxSize     = 16 * 1024 * 1024
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

	dataLen := uint32(len(data))
	if dataLen > shm.header.Capacity {
		return fmt.Errorf("data too large: %d bytes", len(data))
	}

	available := shm.availableSpaceLocked()
	if dataLen > available {
		return fmt.Errorf("ring buffer full: %d bytes needed, %d available", dataLen, available)
	}

	headerSize := uint32(unsafe.Sizeof(SharedMemHeader{}))
	dataOffset := shm.header.WritePos % shm.header.Capacity
	writeStart := headerSize + dataOffset

	endPos := dataOffset + dataLen
	if endPos > shm.header.Capacity {
		firstPart := shm.header.Capacity - dataOffset
		secondPart := dataLen - firstPart

		copy(shm.data[writeStart:writeStart+firstPart], data[:firstPart])

		copy(shm.data[headerSize:headerSize+secondPart], data[firstPart:])
	} else {
		copy(shm.data[writeStart:writeStart+dataLen], data)
	}

	shm.header.WritePos += dataLen

	return nil
}

func (shm *SharedMemory) availableSpaceLocked() uint32 {
	used := shm.header.WritePos - shm.header.ReadPos
	if used >= shm.header.Capacity {
		return 0
	}
	return shm.header.Capacity - used
}

func (shm *SharedMemory) Read(dst []byte) (int, error) {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	if shm.closed {
		return 0, fmt.Errorf("shared memory closed")
	}

	available := shm.header.WritePos - shm.header.ReadPos
	if available == 0 {
		return 0, nil
	}

	readLen := uint32(len(dst))
	if readLen > available {
		return 0, fmt.Errorf("not enough data available: %d bytes requested, %d available", len(dst), available)
	}

	headerSize := uint32(unsafe.Sizeof(SharedMemHeader{}))
	dataOffset := shm.header.ReadPos % shm.header.Capacity
	readStart := headerSize + dataOffset

	endPos := dataOffset + readLen
	if endPos > shm.header.Capacity {
		firstPart := shm.header.Capacity - dataOffset
		secondPart := readLen - firstPart

		copy(dst[:firstPart], shm.data[readStart:readStart+firstPart])

		copy(dst[firstPart:], shm.data[headerSize:headerSize+secondPart])
	} else {
		copy(dst, shm.data[readStart:readStart+readLen])
	}

	shm.header.ReadPos += readLen

	return int(readLen), nil
}

func (shm *SharedMemory) IsClosed() bool {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	return shm.closed
}

func (shm *SharedMemory) Available() uint32 {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	return shm.availableSpaceLocked()
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
	shm.mu.Lock()
	defer shm.mu.Unlock()

	if shm.closed {
		return fmt.Errorf("shared memory closed")
	}

	if shm.fd < 0 {
		return fmt.Errorf("invalid file descriptor")
	}

	socketFd := int(conn.Fd())

	rights := unix.UnixRights(int(shm.fd))
	rightsBytes := append([]byte(nil), rights...)

	if len(rightsBytes) == 0 {
		return fmt.Errorf("failed to serialize rights")
	}

	data := []byte("FD")
	oob := append([]byte(nil), rights...)

	err := syscall.Sendmsg(socketFd, data, oob, 0)
	if err != nil {
		return fmt.Errorf("sendmsg failed: %w", err)
	}

	return nil
}

func RecvFd(conn *os.File) (*SharedMemory, error) {
	socketFd := int(conn.Fd())

	buf := make([]byte, 32)
	oob := make([]byte, unix.CmsgSpace(1))

	n, oobn, flags, err := syscall.Recvmsg(socketFd, buf, oob, 0)
	if err != nil {
		return nil, fmt.Errorf("recvmsg failed: %w", err)
	}

	if flags&syscall.MSG_TRUNC != 0 {
		return nil, fmt.Errorf("message truncated")
	}

	if flags&syscall.MSG_CTRUNC != 0 {
		return nil, fmt.Errorf("control message truncated")
	}

	if n == 0 && oobn == 0 {
		return nil, fmt.Errorf("no data received")
	}

	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, fmt.Errorf("parse socket control failed: %w", err)
	}

	for _, msg := range msgs {
		if msg.Header.Type == syscall.SCM_RIGHTS {
			fds, err := syscall.ParseUnixRights(&msg)
			if err != nil {
				return nil, fmt.Errorf("parse unix rights failed: %w", err)
			}

			if len(fds) == 0 {
				return nil, fmt.Errorf("no file descriptor received")
			}

			receivedFd := fds[0]

			var stat syscall.Stat_t
			if err := syscall.Fstat(receivedFd, &stat); err != nil {
				syscall.Close(receivedFd)
				return nil, fmt.Errorf("failed to stat shared memory: %w", err)
			}
			size := int(stat.Size)

			data, err := unix.Mmap(receivedFd, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
			if err != nil {
				syscall.Close(receivedFd)
				return nil, fmt.Errorf("failed to mmap shared memory: %w", err)
			}

			shm := &SharedMemory{
				fd:       receivedFd,
				data:     data,
				isServer: false,
				memFd:    os.NewFile(uintptr(receivedFd), "v2e-shmem-recv"),
			}

			shm.header = (*SharedMemHeader)(unsafe.Pointer(&data[0]))

			if err := shm.validateHeader(); err != nil {
				shm.Close()
				return nil, fmt.Errorf("invalid shared memory header: %w", err)
			}

			return shm, nil
		}
	}

	return nil, fmt.Errorf("no SCM_RIGHTS message received")
}

func alignToPage(size uint64) uint64 {
	pageSize := uint64(os.Getpagesize())
	return ((size + pageSize - 1) / pageSize) * pageSize
}
