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

	"github.com/hajimehoshi/go-mp3/internal/consts"
	"github.com/hajimehoshi/go-mp3/internal/frame"
)

// A Decoder is a MP3-decoded stream.
//
// Decoder decodes its underlying source on the fly.
type Decoder struct {
	source      *source
	sampleRate  int
	length      int64
	frameStarts []int64
	buf         []byte
	frame       *frame.Frame
	pos         int64
}

func (d *Decoder) readFrame() error {
	var err error
	d.frame, _, err = d.source.readNextFrame(d.frame)
	if err != nil {
		if err == io.EOF {
			return io.EOF
		}
		if _, ok := err.(*consts.UnexpectedEOF); ok {
			// TODO: Log here?
			return io.EOF
		}
		return err
	}
	d.buf = append(d.buf, d.frame.Decode()...)
	return nil
}

// Read is io.Reader's Read.
func (d *Decoder) Read(buf []byte) (int, error) {
	for len(d.buf) == 0 {
		if err := d.readFrame(); err != nil {
			return 0, err
		}
	}
	n := copy(buf, d.buf)
	d.buf = d.buf[n:]
	d.pos += int64(n)
	return n, nil
}

// Seek is io.Seeker's Seek.
//
// Seek panics when the underlying source is not io.Seeker.
func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	npos := int64(0)
	switch whence {
	case io.SeekStart:
		npos = offset
	case io.SeekCurrent:
		npos = d.pos + offset
	case io.SeekEnd:
		npos = d.length + offset
	default:
		panic(fmt.Sprintf("mp3: invalid whence: %v", whence))
	}
	d.pos = npos
	d.buf = nil
	d.frame = nil
	f := d.pos / consts.BytesPerFrame
	// If the frame is not first, read the previous ahead of reading that
	// because the previous frame can affect the targeted frame.
	if f > 0 {
		f--
		if _, err := d.source.Seek(d.frameStarts[f], 0); err != nil {
			return 0, err
		}
		if err := d.readFrame(); err != nil {
			return 0, err
		}
		if err := d.readFrame(); err != nil {
			return 0, err
		}
		d.buf = d.buf[consts.BytesPerFrame+(d.pos%consts.BytesPerFrame):]
	} else {
		if _, err := d.source.Seek(d.frameStarts[f], 0); err != nil {
			return 0, err
		}
		if err := d.readFrame(); err != nil {
			return 0, err
		}
		d.buf = d.buf[d.pos:]
	}
	return npos, nil
}

// Close is io.Closer's Close.
func (d *Decoder) Close() error {
	return d.source.Close()
}

// SampleRate returns the sample rate like 44100.
//
// Note that the sample rate is retrieved from the first frame.
func (d *Decoder) SampleRate() int {
	return d.sampleRate
}

// Length returns the total size in bytes.
//
// Length returns -1 when the total size is not available
// e.g. when the given source is not io.Seeker.
func (d *Decoder) Length() int64 {
	return d.length
}

// NewDecoder decodes the given io.ReadCloser and returns a decoded stream.
//
// The stream is always formatted as 16bit (little endian) 2 channels
// even if the source is single channel MP3.
// Thus, a sample always consists of 4 bytes.
//
// If r is io.Seeker, a decoded stream checks its length and Length returns a valid value.
func NewDecoder(r io.ReadCloser) (*Decoder, error) {
	s := &source{
		reader: r,
	}
	d := &Decoder{
		source: s,
		length: -1,
	}
	if _, ok := r.(io.Seeker); ok {
		if err := s.skipTags(); err != nil {
			return nil, err
		}
		l := int64(0)
		var f *frame.Frame
		for {
			var err error
			pos := int64(0)
			f, pos, err = s.readNextFrame(f)
			if err != nil {
				if err == io.EOF {
					break
				}
				if _, ok := err.(*consts.UnexpectedEOF); ok {
					// TODO: Log here?
					break
				}
				return nil, err
			}
			d.frameStarts = append(d.frameStarts, pos)
			l += consts.BytesPerFrame
		}
		if err := s.rewind(); err != nil {
			return nil, err
		}
		d.length = l
	}
	if err := s.skipTags(); err != nil {
		return nil, err
	}
	// TODO: Is readFrame here really needed?
	if err := d.readFrame(); err != nil {
		return nil, err
	}
	d.sampleRate = d.frame.SamplingFrequency()
	return d, nil
}
