package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {

	idx := &index{
		file: f,
	}
	// why is fi named fi? it's short for "file info". don't confuse for f, short for "file"
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	); err != nil {
		return nil, err
	}

	return idx, nil
}

func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	// https://stackoverflow.com/questions/10862375/when-to-flush-a-file-in-go
	// the link above explains why we need to call sync. The data is in memory and not in a buffer
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

// Read(in int64) Why did we name our parameter? It's short for input, and that input is the offset.
// Why are we passing in an offset? From this offset, we can get the associated record's position in the store.
// Why do we truncate? Because the offset is relative to the segement's base offset(it means this input value stores
// both the segment offset and index offset, and both are 4 bytes or 32 bits long). Why do we want the offset for the
// index file? With the offset of the index file, we can caculate the position of our entry in the index file. Why
// do we want the entry in the index file?
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	// why are we performing this calculating? we multiply the relative offset by the entWidth. This gives our entry location in the index file.
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}

	// The offset has a lenght of 4 bytes. We want to read a subslice of that size. Todo: work with slices more in go
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	// next we start from the last place we read the offset (which is 4 bytes). Why are we stoping the read
	// at pos+entWidth? Because there's overlap. entWidth - offWidth = 8 bytes. Therefore, we are extracting 8 bytes.
	// Reminder slices are read as slice[start:end]
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	// Does this don't write if we don't have any room in our index file?
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}

	// I don't get how to read this memormy map
	// usually : is used to give us a sub-slice in a slice
	// this means, from i.size to the next 4 bytes, store the offset
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	// Becuase i.size to i.size+offWidth (the next 4 bytes) stores the offset value,
	// this is where we want to start storing the position for our record
	// why do store the position in this subslice i.size+offWidth:i.size+entWidth
	// our starting point is the position after we stored our index. entWidth is 4 + 8 bytes long
	// there's overlap with i.size+offWidth:i.size+entWidth, meaning i.size + (entWidth - offWidth) = 8 bytes,
	// which is the length of our position
	// Why are memory maps like encoding something to sub-slice? Is that really how it works?
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
