package teesectionreader

import (
	"io"

	"github.com/pkg/errors"
)

// NewTeeSectionReader returns a TeeSectionReader that reads from r
// starting at offset off and stops with EOF after n bytes.
func NewTeeSectionReader(r io.ReaderAt, w io.Writer, off int64, n int64) *TeeSectionReader {
	return &TeeSectionReader{r, w, off, off, off + n, off}
}

// TeeSectionReader implements Read, Seek, and ReadAt on a section
// of an underlying ReaderAt.
type TeeSectionReader struct {
	r             io.ReaderAt
	w             io.Writer
	base          int64
	off           int64
	limit         int64
	writtenOffset int64
}

/*

TeeSectionReader
================

TeeSectionReader is like an amalgamation of the StdLib SectionReader, and the TeeReader.
Things to note about it are:

1. It takes a writer as an argument to the create function
2. It writes all bytes read to the writer provided

The goal of this is to be able to checksum the data read using the Read method by
passing in an object that is a writer and passing all data read to this writer. This is
complicated by the fact that the AWS SDK which this is designed to be used with reads
the source data twice, once to checksum a request, and once to transfer the data. There
was no easy way of getting round this so I've had to implement a sort of "ratchet"
writer, keeping track of the offset of data written so that re-reads aren't then sent
to the writer, messing up the checksums.

If this code seems completely weird to you just find me and talk to me about it :)

*/

func (s *TeeSectionReader) Read(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, io.EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.ReadAt(p, s.off)
	s.off += int64(n)

	if n > 0 {
		if s.off > s.writtenOffset {
			n1, err := s.w.Write(p[s.off-s.writtenOffset-int64(n) : s.off-s.writtenOffset])
			if err != nil {
				return n1, err
			}

			s.writtenOffset += int64(n1)
		}
	}

	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

// Seek to a point in a file
func (s *TeeSectionReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case io.SeekStart:
		offset += s.base
	case io.SeekCurrent:
		offset += s.off
	case io.SeekEnd:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

// ReadAt an offset in a file into p
func (s *TeeSectionReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit-s.base {
		return 0, io.EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err = s.r.ReadAt(p, off)
		if err == nil {
			err = io.EOF
		}

		return n, err
	}
	return s.r.ReadAt(p, off)
}

// Size returns the size of the section in bytes.
func (s *TeeSectionReader) Size() int64 { return s.limit - s.base }
