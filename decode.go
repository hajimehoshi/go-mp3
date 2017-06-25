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
)

type unexpectedEOF struct {
	At string
}

func (u *unexpectedEOF) Error() string {
	return fmt.Sprintf("mp3: unexpected EOF at %s", u.At)
}

func (f *frame) decodeL3() []uint8 {
	out := make([]uint8, bytesPerFrame)
	nch := f.header.numberOfChannels()
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < nch; ch++ {
			f.l3Requantize(gr, ch)
			// Reorder short blocks
			f.l3Reorder(gr, ch)
		}
		f.l3Stereo(gr)
		for ch := 0; ch < nch; ch++ {
			f.l3Antialias(gr, ch)
			// (IMDCT,windowing,overlapp add)
			f.l3HybridSynthesis(gr, ch)
			f.l3FrequencyInversion(gr, ch)
			// Polyphase subband synthesis
			f.l3SubbandSynthesis(gr, ch, out[samplesPerFrame*4*gr:])
		}
	}
	return out
}

type source struct {
	reader io.ReadCloser
	pos    int
}

func (s *source) Close() error {
	return s.reader.Close()
}

func (s *source) rewind() error {
	seeker := s.reader.(io.Seeker)
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return err
	}
	s.pos = 0
	return nil
}

func (s *source) getByte() (uint8, error) {
	b := make([]uint8, 1)
	n, err := s.getBytes(b)
	s.pos += n
	return b[0], err
}

func (s *source) getBytes(buf []uint8) (int, error) {
	n, err := io.ReadFull(s.reader, buf)
	s.pos += n
	return n, err
}

func (s *source) getFilepos() int {
	return s.pos
}

type Decoded struct {
	source     *source
	sampleRate int
	length     int64
	buf        []uint8
	frame      *frame
	eof        bool
}

func (d *Decoded) read() error {
	var err error
	d.frame, err = d.source.readNextFrame(d.frame)
	if err != nil {
		if err == io.EOF {
			d.eof = true
		}
		if _, ok := err.(*unexpectedEOF); ok {
			// TODO: Log here?
			d.eof = true
		}
		return io.EOF
	}
	d.buf = append(d.buf, d.frame.decodeL3()...)
	return nil
}

// Read is io.Reader's Read.
func (d *Decoded) Read(buf []uint8) (int, error) {
	for len(d.buf) == 0 && !d.eof {
		if err := d.read(); err != nil {
			return 0, err
		}
	}
	if d.eof {
		return 0, io.EOF
	}
	n := copy(buf, d.buf)
	d.buf = d.buf[n:]
	return n, nil
}

// Close is io.Closer's Close.
func (d *Decoded) Close() error {
	return d.source.Close()
}

// SampleRate returns the sample rate like 44100.
//
// Note that the sample rate is retrieved from the first frame.
func (d *Decoded) SampleRate() int {
	return d.sampleRate
}

// Length returns the total size in bytes.
//
// Length returns -1 when the total size is not available
// e.g. when the given source is not io.Seeker.
func (d *Decoded) Length() int64 {
	return d.length
}

// Decode decodes the given io.ReadCLoser and returns a decoded stream.
//
// The stream is always formatted as 16bit (little endian) 2 channels
// even if the source is single channel MP3.
// Thus, a sample always consists of 4 bytes.
//
// If r is io.Seeker, a decoded stream checks its length and Length returns a valid value.
func Decode(r io.ReadCloser) (*Decoded, error) {
	s := &source{
		reader: r,
	}
	d := &Decoded{
		source: s,
		length: -1,
	}
	if _, ok := r.(io.Seeker); ok {
		l := int64(0)
		var f *frame
		for {
			var err error
			f, err = s.readNextFrame(f)
			if err != nil {
				if err == io.EOF {
					break
				}
				if _, ok := err.(*unexpectedEOF); ok {
					// TODO: Log here?
					break
				}
				return nil, err
			}
			l += bytesPerFrame
		}
		if err := s.rewind(); err != nil {
			return nil, err
		}
		d.length = l
	}
	if err := d.read(); err != nil {
		return nil, err
	}
	d.sampleRate = samplingFrequency[d.frame.header.sampling_frequency]
	return d, nil
}
