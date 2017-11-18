// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mp3

import (
	"fmt"
	"io"

	"github.com/hajimehoshi/go-mp3/internal/bits"
	"github.com/hajimehoshi/go-mp3/internal/consts"
	"github.com/hajimehoshi/go-mp3/internal/frame"
	"github.com/hajimehoshi/go-mp3/internal/frameheader"
	"github.com/hajimehoshi/go-mp3/internal/maindata"
	"github.com/hajimehoshi/go-mp3/internal/sideinfo"
)

type source struct {
	reader io.ReadCloser
	buf    []byte
	pos    int64
}

func (s *source) Seek(position int64, whence int) (int64, error) {
	seeker, ok := s.reader.(io.Seeker)
	if !ok {
		panic("mp3: source must be io.Seeker")
	}
	s.buf = nil
	n, err := seeker.Seek(position, whence)
	if err != nil {
		return 0, err
	}
	s.pos = n
	return n, nil
}

func (s *source) Close() error {
	s.buf = nil
	return s.reader.Close()
}

func (s *source) skipTags() error {
	buf := make([]byte, 3)
	if _, err := s.ReadFull(buf); err != nil {
		return err
	}
	switch string(buf) {
	case "TAG":
		buf := make([]byte, 125)
		if _, err := s.ReadFull(buf); err != nil {
			return err
		}

	case "ID3":
		// Skip version (2 bytes) and flag (1 byte)
		buf := make([]byte, 3)
		if _, err := s.ReadFull(buf); err != nil {
			return err
		}

		buf = make([]byte, 4)
		n, err := s.ReadFull(buf)
		if err != nil {
			return err
		}
		if n != 4 {
			return nil
		}
		size := (uint32(buf[0]) << 21) | (uint32(buf[1]) << 14) |
			(uint32(buf[2]) << 7) | uint32(buf[3])
		buf = make([]byte, size)
		if _, err := s.ReadFull(buf); err != nil {
			return err
		}

	default:
		s.Unread(buf)
	}

	return nil
}

func (s *source) rewind() error {
	if _, err := s.Seek(0, io.SeekStart); err != nil {
		return err
	}
	s.pos = 0
	s.buf = nil
	return nil
}

func (s *source) Unread(buf []byte) {
	s.buf = append(s.buf, buf...)
	s.pos -= int64(len(buf))
}

func (s *source) ReadFull(buf []byte) (int, error) {
	read := 0
	if s.buf != nil {
		read = copy(buf, s.buf)
		if len(s.buf) > read {
			s.buf = s.buf[read:]
		} else {
			s.buf = nil
		}
		if len(buf) == read {
			return read, nil
		}
	}

	n, err := io.ReadFull(s.reader, buf[read:])
	if err != nil {
		// Allow if all data can't be read. This is common.
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
	}
	s.pos += int64(n)
	return n + read, err
}

func (s *source) readCRC() error {
	buf := make([]byte, 2)
	n, err := s.ReadFull(buf)
	if n < 2 {
		if err == io.EOF {
			return &consts.UnexpectedEOF{"readCRC"}
		}
		return fmt.Errorf("mp3: error at readCRC: %v", err)
	}
	return nil
}

func (s *source) readNextFrame(prev *frame.Frame) (f *frame.Frame, startPosition int64, err error) {
	h, pos, err := s.readHeader()
	if err != nil {
		return nil, 0, err
	}
	// Get CRC word if present
	if h.ProtectionBit() == 0 {
		if err := s.readCRC(); err != nil {
			return nil, 0, err
		}
	}
	if h.ID() != consts.Version1 {
		return nil, 0, fmt.Errorf("mp3: only MPEG version 1 (want %d; got %d) is supported", consts.Version1, h.ID())
	}
	if h.Layer() != consts.Layer3 {
		return nil, 0, fmt.Errorf("mp3: only layer3 (want %d; got %d) is supported", consts.Layer3, h.Layer())
	}
	// Get side info
	si, err := sideinfo.Read(s, h)
	if err != nil {
		return nil, 0, err
	}
	// If there's not enough main data in the bit reservoir,
	// signal to calling function so that decoding isn't done!
	// Get main data(scalefactors and Huffman coded frequency data)
	var prevM *bits.Bits
	if prev != nil {
		prevM = prev.MainDataBits()
	}
	md, mdb, err := maindata.Read(s, prevM, h, si)
	if err != nil {
		return nil, 0, err
	}
	nf := frame.New(h, si, md, mdb, prev)
	return nf, pos, nil
}

func (s *source) readHeader() (h frameheader.FrameHeader, startPosition int64, err error) {
	pos := s.pos
	buf := make([]byte, 4)
	if n, err := s.ReadFull(buf); n < 4 {
		if err == io.EOF {
			if n == 0 {
				// Expected EOF
				return 0, 0, io.EOF
			}
			return 0, 0, &consts.UnexpectedEOF{"readHeader (1)"}
		}
		return 0, 0, err
	}

	b1 := uint32(buf[0])
	b2 := uint32(buf[1])
	b3 := uint32(buf[2])
	b4 := uint32(buf[3])
	header := frameheader.FrameHeader((b1 << 24) | (b2 << 16) | (b3 << 8) | (b4 << 0))
	for !header.IsValid() {
		b1 = b2
		b2 = b3
		b3 = b4

		buf := make([]byte, 1)
		if _, err := s.ReadFull(buf); err != nil {
			if err == io.EOF {
				return 0, 0, &consts.UnexpectedEOF{"readHeader (2)"}
			}
			return 0, 0, err
		}
		b4 = uint32(buf[0])
		header = frameheader.FrameHeader((b1 << 24) | (b2 << 16) | (b3 << 8) | (b4 << 0))
		pos++
	}

	// If we get here we've found the sync word, and can decode the header
	// which is in the low 20 bits of the 32-bit sync+header word.

	if header.BitrateIndex() == 0 {
		return 0, 0, fmt.Errorf("mp3: free bitrate format is not supported. Header word is 0x%08x at position %d",
			header, pos)
	}
	return header, pos, nil
}
