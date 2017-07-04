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
)

func (s *source) readCRC() error {
	buf := make([]uint8, 2)
	n, err := s.getBytes(buf)
	if n < 2 {
		if err == io.EOF {
			return &unexpectedEOF{"readCRC"}
		}
		return fmt.Errorf("mp3: error at readCRC: %v", err)
	}
	return nil
}

func (s *source) readNextFrame(prev *frame) (f *frame, startPosition int64, err error) {
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
	if h.Layer() != mpeg1Layer3 {
		return nil, 0, fmt.Errorf("mp3: only layer3 (want %d; got %d) is supported!", mpeg1Layer3, h.Layer())
	}
	// Get side info
	si, err := s.readSideInfo(h)
	if err != nil {
		return nil, 0, err
	}
	// If there's not enough main data in the bit reservoir,
	// signal to calling function so that decoding isn't done!
	// Get main data(scalefactors and Huffman coded frequency data)
	var prevM *bits.Bits
	if prev != nil {
		prevM = prev.mainDataBytes
	}
	md, mdb, err := s.readMainL3(prevM, h, si)
	if err != nil {
		return nil, 0, err
	}
	nf := &frame{
		header:        h,
		sideInfo:      si,
		mainData:      md,
		mainDataBytes: mdb,
	}
	if prev != nil {
		nf.store = prev.store
		nf.v_vec = prev.v_vec
	}
	return nf, pos, nil
}

func isHeader(header uint32) bool {
	const C_SYNC = 0xffe00000
	if (header & C_SYNC) != C_SYNC {
		return false
	}
	// Bitrate must not be 15.
	if (header & (0xf << 12)) == 0xf<<12 {
		return false
	}
	// Sample Frequency must not be 3.
	if (header & (3 << 10)) == 3<<10 {
		return false
	}
	return true
}

func (s *source) readHeader() (h *mpeg1FrameHeader, startPosition int64, err error) {
	// Get the next four bytes from the bitstream
	pos := s.getFilepos()
	buf := make([]uint8, 4)
	n, err := s.getBytes(buf)
	if n < 4 {
		if err == io.EOF {
			if n == 0 {
				// Expected EOF
				return nil, 0, io.EOF
			}
			return nil, 0, &unexpectedEOF{"readHeader (1)"}
		}
		return nil, 0, err
	}
	b1 := uint32(buf[0])
	b2 := uint32(buf[1])
	b3 := uint32(buf[2])
	b4 := uint32(buf[3])
	header := (b1 << 24) | (b2 << 16) | (b3 << 8) | (b4 << 0)
	for !isHeader(uint32(header)) {
		// No,so scan the bitstream one byte at a time until we find it or EOF
		// Shift the values one byte to the left
		b1 = b2
		b2 = b3
		b3 = b4
		// Get one new byte from the bitstream
		buf := make([]uint8, 1)
		if _, err := s.getBytes(buf); err != nil {
			if err == io.EOF {
				return nil, 0, &unexpectedEOF{"readHeader (2)"}
			}
			return nil, 0, err
		}
		b4 = uint32(buf[0])
		header = (b1 << 24) | (b2 << 16) | (b3 << 8) | (b4 << 0)
		pos++
	}
	// If we get here we've found the sync word,and can decode the header
	// which is in the low 20 bits of the 32-bit sync+header word.
	// NewDecoder the header
	head := mpeg1FrameHeader(header)

	// Check for invalid values and impossible combinations
	if head.ID() != 3 {
		return nil, 0, fmt.Errorf("mp3: ID must be 3. Header word is 0x%08x at file pos %d",
			header, s.getFilepos())
	}
	if head.BitrateIndex() == 0 {
		return nil, 0, fmt.Errorf("mp3: Free bitrate format NIY! Header word is 0x%08x at file pos %d",
			header, s.getFilepos())
	}
	if head.BitrateIndex() == 15 {
		return nil, 0, fmt.Errorf("mp3: bitrate_index = 15 is invalid! Header word is 0x%08x at file pos %d",
			header, s.getFilepos())
	}
	if head.SamplingFrequency() == 3 {
		return nil, 0, fmt.Errorf("mp3: sampling_frequency = 3 is invalid! Header word is 0x%08x at file pos %d",
			header, s.getFilepos())
	}
	if head.Layer() == mpeg1LayerReserved {
		return nil, 0, fmt.Errorf("mp3: layer = %d is invalid! Header word is 0x%08x at file pos %d",
			mpeg1LayerReserved, header, s.getFilepos())
	}
	return &head, pos, nil
}

func readHuffman(m *bits.Bits, header *mpeg1FrameHeader, sideInfo *mpeg1SideInfo, mainData *mpeg1MainData, part_2_start, gr, ch int) error {
	// Check that there is any data to decode. If not,zero the array.
	if sideInfo.part2_3_length[gr][ch] == 0 {
		for is_pos := 0; is_pos < samplesPerGr; is_pos++ {
			mainData.is[gr][ch][is_pos] = 0.0
		}
		return nil
	}
	// Calculate bit_pos_end which is the index of the last bit for this part.
	bit_pos_end := part_2_start + sideInfo.part2_3_length[gr][ch] - 1
	// Determine region boundaries
	region_1_start := 0
	region_2_start := 0
	if (sideInfo.win_switch_flag[gr][ch] == 1) && (sideInfo.block_type[gr][ch] == 2) {
		region_1_start = 36           // sfb[9/3]*3=36
		region_2_start = samplesPerGr // No Region2 for short block case.
	} else {
		sfreq := header.SamplingFrequency()
		l := sfBandIndicesSet[sfreq].l
		i := sideInfo.region0_count[gr][ch] + 1
		if i < 0 || len(l) <= i {
			// TODO: Better error messages (#3)
			return fmt.Errorf("mp3: readHuffman failed: invalid index i: %d", i)
		}
		region_1_start = l[i]
		j := sideInfo.region0_count[gr][ch] + sideInfo.region1_count[gr][ch] + 2
		if j < 0 || len(l) <= j {
			// TODO: Better error messages (#3)
			return fmt.Errorf("mp3: readHuffman failed: invalid index j: %d", j)
		}
		region_2_start = l[j]
	}
	// Read big_values using tables according to region_x_start
	for is_pos := 0; is_pos < sideInfo.big_values[gr][ch]*2; is_pos++ {
		table_num := 0
		if is_pos < region_1_start {
			table_num = sideInfo.table_select[gr][ch][0]
		} else if is_pos < region_2_start {
			table_num = sideInfo.table_select[gr][ch][1]
		} else {
			table_num = sideInfo.table_select[gr][ch][2]
		}
		// Get next Huffman coded words
		x, y, _, _, err := huffmanDecode(m, table_num)
		if err != nil {
			return err
		}
		// In the big_values area there are two freq lines per Huffman word
		mainData.is[gr][ch][is_pos] = float32(x)
		is_pos++
		mainData.is[gr][ch][is_pos] = float32(y)
	}
	// Read small values until is_pos = 576 or we run out of huffman data
	table_num := sideInfo.count1table_select[gr][ch] + 32
	is_pos := sideInfo.big_values[gr][ch] * 2
	for (is_pos <= 572) && (m.Pos() <= bit_pos_end) {
		// Get next Huffman coded words
		x, y, v, w, err := huffmanDecode(m, table_num)
		if err != nil {
			return err
		}
		mainData.is[gr][ch][is_pos] = float32(v)
		is_pos++
		if is_pos >= samplesPerGr {
			break
		}
		mainData.is[gr][ch][is_pos] = float32(w)
		is_pos++
		if is_pos >= samplesPerGr {
			break
		}
		mainData.is[gr][ch][is_pos] = float32(x)
		is_pos++
		if is_pos >= samplesPerGr {
			break
		}
		mainData.is[gr][ch][is_pos] = float32(y)
		is_pos++
	}
	// Check that we didn't read past the end of this section
	if m.Pos() > (bit_pos_end + 1) {
		// Remove last words read
		is_pos -= 4
	}
	// Setup count1 which is the index of the first sample in the rzero reg.
	sideInfo.count1[gr][ch] = is_pos
	// Zero out the last part if necessary
	for is_pos < samplesPerGr {
		mainData.is[gr][ch][is_pos] = 0.0
		is_pos++
	}
	// Set the bitpos to point to the next part to read
	m.SetPos(bit_pos_end + 1)
	return nil
}
