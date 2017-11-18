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
