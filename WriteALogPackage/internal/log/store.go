package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian // TODO: review big endian
)

const (
	// Number of bytes to store the record's length
	lenWidth = 8
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f), // TODO: practice using bufio package a bit
	}, nil
}

// Append persists the given bytes to the store.
// This returns the number of bytes written and the position where the store holds the record in its file.
// We write the length of the record so that, when we read the record, we know how many bytes to read.
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size // I'm assuming the inital size is 0 when we intialize the store
	// TODO: practice using encoding/binary package
	// 8 bytes = 8 bits * 8 (there's 8 bits in a byte);  therfore, 64 bits. We store the length of the
	// record, which will always take up 64 bits.
	// We need to take that into consideration when reading the message (pos + lenWidth)
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	w += lenWidth       // why are we returning the number of bytes written + lenWidth? (lenWidth is the number of bytes used to store the records length); why is this the next position of our next record; Because we are storing the length of the record and the record's message as well. The length of the record will always take up to 64 bits (a fixed length). We sum up 64 bits to the number of bytes used to stored the message
	s.size += uint64(w) // update the size, this updated value gives us the position of the next/soon to be added  record
	return uint64(w), pos, nil
}

// Returns the record stored at the given position
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil { // TODO: understand concept of flushing a buffer
		return nil, err
	}
	size := make([]byte, lenWidth) // Does two things: 1) checks to see if the record is less than lenWidth (8 bytes); 2) this returns the actual length of our record
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

// ReadAt reads len(p) bytes into p beginning at the off offset in the store's file
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
