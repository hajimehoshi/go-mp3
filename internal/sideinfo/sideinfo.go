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

package sideinfo

// A SideInfo is MPEG1 Layer 3 Side Information.
// [2][2] means [gr][ch].
type SideInfo struct {
	MainDataBegin    int       // 9 bits
	PrivateBits      int       // 3 bits in mono, 5 in stereo
	Scfsi            [2][4]int // 1 bit
	Part2_3Length    [2][2]int // 12 bits
	BigValues        [2][2]int // 9 bits
	GlobalGain       [2][2]int // 8 bits
	ScalefacCompress [2][2]int // 4 bits
	WinSwitchFlag    [2][2]int // 1 bit

	BlockType      [2][2]int    // 2 bits
	MixedBlockFlag [2][2]int    // 1 bit
	TableSelect    [2][2][3]int // 5 bits
	SubblockGain   [2][2][3]int // 3 bits

	Region0Count [2][2]int // 4 bits
	Region1Count [2][2]int // 3 bits

	Preflag           [2][2]int // 1 bit
	ScalefacScale     [2][2]int // 1 bit
	Count1TableSelect [2][2]int // 1 bit
	Count1            [2][2]int // Not in file,calc. by huff.dec.!
}
