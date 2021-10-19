package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/manofthelionarmy/prolog/api/v1"
	"google.golang.org/protobuf/proto"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}
	var err error
	/// TODO: study file mode flags
	storeFile, err := os.OpenFile(path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}

	indexFile, err := os.OpenFile(path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}

	// Interesting, what does this block of code do and why?
	// Why do we read from the index? We want to check if it's empty. If so, then the
	// next record appended to the segment would be the first record (duh), therefore
	// the next offset is the segment's base offset.
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		// Why do we want to calculate the nextOffset as such? Well, if the index is not empty,
		// it means we have at least one record appended to the segment. Therefore, the nextOffset is
		// the offset at the end of the segment, which is baseOffset + relativeOffset + 1.
		// Recall it's:
		// +--------------------------+---------------------+
		// | relativeOffset (32 bits) | baseOffset (32 bits)|
		// +--------------------------+---------------------+
		s.nextOffset = baseOffset + uint64(off) + 1
	}

	return s, nil
}

func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	cur := s.nextOffset
	record.Offset = cur
	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}
	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}
	// index offsets are relative to base offset
	// +--------------------------+---------------------+
	// | relativeOffset (32 bits) | baseOffset (32 bits)|
	// +--------------------------+---------------------+
	if err = s.index.Write(
		// why are we subtracting? the index's offset is the relative to the base offset
		uint32(s.nextOffset-uint64(s.baseOffset)),
		pos,
	); err != nil {
		return 0, err
	}
	s.nextOffset++
	return cur, nil
}

func (s *segment) Read(off uint64) (*api.Record, error) {
	// Why are we reading from the index? We want to retrieve a record entry's position from the index
	// why are we taking the difference? The index offset parameter is the absolute index. We need
	// to translate the absolute index into a relative offset.
	_, pos, err := s.index.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}
	// Why are reading from the store? We are using the retrieved record-entry position from the index file
	// to retrieve the record stored in the store file
	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}

	record := &api.Record{}
	// why are we using proto.Unmarshal()? we want to use the data and extract it as api.Record type
	err = proto.Unmarshal(p, record)
	return record, err
}

func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes ||
		s.index.size >= s.config.Segment.MaxIndexBytes
}

func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}

func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return nil
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	return nil
}

func nearestMultiple(j, k uint64) uint64 {
	// every value of type uint64 is >= 0
	if j >= 0 {
		return (j / k) * k
	}
	// handles negative case
	return ((j - k + 1) / k) * k
}
