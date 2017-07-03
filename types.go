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

import "github.com/hajimehoshi/go-mp3/internal/bits"

type mpeg1Layer int

const (
	mpeg1LayerReserved mpeg1Layer = 0
	mpeg1Layer3        mpeg1Layer = 1
	mpeg1Layer2        mpeg1Layer = 2
	mpeg1Layer1        mpeg1Layer = 3
)

type mpeg1Mode int

const (
	mpeg1ModeStereo mpeg1Mode = iota
	mpeg1ModeJointStereo
	mpeg1ModeDualChannel
	mpeg1ModeSingleChannel
)

// A mepg1FrameHeader is MPEG1 Layer 1-3 frame header
type mpeg1FrameHeader int

// ID returns this header's ID stored in position 20,19
func (m mpeg1FrameHeader) ID() int {
	return int((m & 0x00180000) >> 19)
}

// Layer returns the mpeg layer of this frame stored in position 18,17
func (m mpeg1FrameHeader) Layer() mpeg1Layer {
	return mpeg1Layer((m & 0x00060000) >> 17)
}

// ProtectionBit returns the protection bit stored in position 16
func (m mpeg1FrameHeader) ProtectionBit() int {
	return int(m&0x00010000) >> 16
}

// BirateIndex returns the bitrate index stored in position 15,12
func (m mpeg1FrameHeader) BitrateIndex() int {
	return int(m&0x0000f000) >> 12
}

// SamplingFrequency returns the SamplingFrequency in Hz stored in position 11,10
func (m mpeg1FrameHeader) SamplingFrequency() int {
	return int(m&0x00000c00) >> 10
}

// PaddingBit returns the padding bit stored in position 9
func (m mpeg1FrameHeader) PaddingBit() int {
	return int(m&0x00000200) >> 9
}

// PrivateBit returns the private bit stored in position 8 - this bit may be used to store arbitrary data to be used
// by an application
func (m mpeg1FrameHeader) PrivateBit() int {
	return int(m&0x00000100) >> 8
}

// Mode returns the channel mode, stored in position 7,6
func (m mpeg1FrameHeader) Mode() mpeg1Mode {
	return mpeg1Mode((m & 0x000000c0) >> 6)
}

// ModeExtension returns the mode_extension - for use with Joint Stereo - stored in position 4,5
func (m mpeg1FrameHeader) ModeExtension() int {
	return int(m&0x00000030) >> 4
}

// Copyright returns whether or not this recording is copywritten - stored in position 3
func (m mpeg1FrameHeader) Copyright() int {
	return int(m&0x00000008) >> 3
}

// OriginalOrCopy returns whether or not this is an Original recording or a copy of one - stored in position 2
func (m mpeg1FrameHeader) OriginalOrCopy() int {
	return int(m&0x00000004) >> 2
}

// Emphasis returns emphasis - the emphasis indication is here to tell the decoder that the file must be de-emphasized - stored in position 0,1
func (m mpeg1FrameHeader) Emphasis() int {
	return int(m&0x00000003) >> 0
}

// A mpeg1SideInfo is  MPEG1 Layer 3 Side Information.
// [2][2] means [gr][ch].
type mpeg1SideInfo struct {
	main_data_begin   int       // 9 bits
	private_bits      int       // 3 bits in mono, 5 in stereo
	scfsi             [2][4]int // 1 bit
	part2_3_length    [2][2]int // 12 bits
	big_values        [2][2]int // 9 bits
	global_gain       [2][2]int // 8 bits
	scalefac_compress [2][2]int // 4 bits
	win_switch_flag   [2][2]int // 1 bit

	block_type       [2][2]int    // 2 bits
	mixed_block_flag [2][2]int    // 1 bit
	table_select     [2][2][3]int // 5 bits
	subblock_gain    [2][2][3]int // 3 bits

	region0_count [2][2]int // 4 bits
	region1_count [2][2]int // 3 bits

	preflag            [2][2]int // 1 bit
	scalefac_scale     [2][2]int // 1 bit
	count1table_select [2][2]int // 1 bit
	count1             [2][2]int // Not in file,calc. by huff.dec.!
}

// A mpeg1MainData is MPEG1 Layer 3 Main Data.
type mpeg1MainData struct {
	scalefac_l [2][2][21]int      // 0-4 bits
	scalefac_s [2][2][12][3]int   // 0-4 bits
	is         [2][2][576]float32 // Huffman coded freq. lines
}

type frame struct {
	header   *mpeg1FrameHeader
	sideInfo *mpeg1SideInfo
	mainData *mpeg1MainData

	mainDataBytes *bits.Bits
	store         [2][32][18]float32
	v_vec         [2][1024]float32
}

var mpeg1Bitrates = map[mpeg1Layer][15]int{
	mpeg1Layer1: {
		0, 32000, 64000, 96000, 128000, 160000, 192000, 224000,
		256000, 288000, 320000, 352000, 384000, 416000, 448000,
	},
	mpeg1Layer2: {
		0, 32000, 48000, 56000, 64000, 80000, 96000, 112000,
		128000, 160000, 192000, 224000, 256000, 320000, 384000,
	},
	mpeg1Layer3: {
		0, 32000, 40000, 48000, 56000, 64000, 80000, 96000,
		112000, 128000, 160000, 192000, 224000, 256000, 320000,
	},
}

var samplingFrequency = []int{44100, 48000, 32000}

func (h *mpeg1FrameHeader) frameSize() int {
	return (144*mpeg1Bitrates[h.Layer()][h.BitrateIndex()])/
		samplingFrequency[h.SamplingFrequency()] +
		int(h.PaddingBit())
}

func (h *mpeg1FrameHeader) numberOfChannels() int {
	if h.Mode() == mpeg1ModeSingleChannel {
		return 1
	}
	return 2
}

const (
	samplesPerGr    = 576
	samplesPerFrame = samplesPerGr * 2
	bytesPerFrame   = samplesPerFrame * 4
)
