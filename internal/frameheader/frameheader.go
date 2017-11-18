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

package frameheader

import (
	"github.com/hajimehoshi/go-mp3/internal/consts"
)

// A mepg1FrameHeader is MPEG1 Layer 1-3 frame header
type FrameHeader uint32

// ID returns this header's ID stored in position 20,19
func (m FrameHeader) ID() consts.Version {
	return consts.Version((m & 0x00180000) >> 19)
}

// Layer returns the mpeg layer of this frame stored in position 18,17
func (m FrameHeader) Layer() consts.Layer {
	return consts.Layer((m & 0x00060000) >> 17)
}

// ProtectionBit returns the protection bit stored in position 16
func (m FrameHeader) ProtectionBit() int {
	return int(m&0x00010000) >> 16
}

// BirateIndex returns the bitrate index stored in position 15,12
func (m FrameHeader) BitrateIndex() int {
	return int(m&0x0000f000) >> 12
}

// SamplingFrequency returns the SamplingFrequency in Hz stored in position 11,10
func (m FrameHeader) SamplingFrequency() consts.SamplingFrequency {
	return consts.SamplingFrequency(int(m&0x00000c00) >> 10)
}

// PaddingBit returns the padding bit stored in position 9
func (m FrameHeader) PaddingBit() int {
	return int(m&0x00000200) >> 9
}

// PrivateBit returns the private bit stored in position 8 - this bit may be used to store arbitrary data to be used
// by an application
func (m FrameHeader) PrivateBit() int {
	return int(m&0x00000100) >> 8
}

// Mode returns the channel mode, stored in position 7,6
func (m FrameHeader) Mode() consts.Mode {
	return consts.Mode((m & 0x000000c0) >> 6)
}

// ModeExtension returns the mode_extension - for use with Joint Stereo - stored in position 4,5
func (m FrameHeader) ModeExtension() int {
	return int(m&0x00000030) >> 4
}

// Copyright returns whether or not this recording is copywritten - stored in position 3
func (m FrameHeader) Copyright() int {
	return int(m&0x00000008) >> 3
}

// OriginalOrCopy returns whether or not this is an Original recording or a copy of one - stored in position 2
func (m FrameHeader) OriginalOrCopy() int {
	return int(m&0x00000004) >> 2
}

// Emphasis returns emphasis - the emphasis indication is here to tell the decoder that the file must be de-emphasized - stored in position 0,1
func (m FrameHeader) Emphasis() int {
	return int(m&0x00000003) >> 0
}

// IsValid returns a boolean value indicating whether the header is valid or not.
func (m FrameHeader) IsValid() bool {
	const sync = 0xffe00000
	if (m & sync) != sync {
		return false
	}
	if m.ID() == consts.VersionReserved {
		return false
	}
	if m.BitrateIndex() == 15 {
		return false
	}
	if m.SamplingFrequency() == 3 {
		return false
	}
	if m.Layer() == consts.LayerReserved {
		return false
	}
	if m.Emphasis() == 2 {
		return false
	}
	return true
}

func bitrate(layer consts.Layer, index int) int {
	switch layer {
	case consts.Layer1:
		return []int{
			0, 32000, 64000, 96000, 128000, 160000, 192000, 224000,
			256000, 288000, 320000, 352000, 384000, 416000, 448000}[index]
	case consts.Layer2:
		return []int{
			0, 32000, 48000, 56000, 64000, 80000, 96000, 112000,
			128000, 160000, 192000, 224000, 256000, 320000, 384000}[index]
	case consts.Layer3:
		return []int{
			0, 32000, 40000, 48000, 56000, 64000, 80000, 96000,
			112000, 128000, 160000, 192000, 224000, 256000, 320000}[index]
	}
	panic("not reached")
}

func (h FrameHeader) FrameSize() int {
	return (144*bitrate(h.Layer(), h.BitrateIndex()))/
		h.SamplingFrequency().Int() +
		int(h.PaddingBit())
}

func (h FrameHeader) NumberOfChannels() int {
	if h.Mode() == consts.ModeSingleChannel {
		return 1
	}
	return 2
}
