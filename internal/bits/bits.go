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

package bits

type Bits struct {
	Vec []uint8
	idx int
	pos int
}

func (b *Bits) Bit() int {
	if len(b.Vec) <= b.pos {
		// TODO: Should this return error?
		return 0
	}
	tmp := uint(b.Vec[b.pos]) >> (7 - uint(b.idx))
	tmp &= 0x01
	b.pos += (b.idx + 1) >> 3
	b.idx = (b.idx + 1) & 0x07
	return int(tmp)
}

func (b *Bits) Bits(num int) int {
	if num == 0 {
		return 0
	}
	if len(b.Vec) <= b.pos {
		// TODO: Should this return error?
		return 0
	}
	bb := make([]uint8, 4)
	copy(bb, b.Vec[b.pos:])
	tmp := (uint32(bb[0]) << 24) | (uint32(bb[1]) << 16) | (uint32(bb[2]) << 8) | (uint32(bb[3]) << 0)
	tmp = tmp << uint(b.idx)
	tmp = tmp >> (32 - uint(num))
	b.pos += (b.idx + num) >> 3
	b.idx = (b.idx + num) & 0x07
	return int(tmp)
}

func (b *Bits) Pos() int {
	pos := b.pos
	pos *= 8 // Multiply by 8 to get number of bits
	pos += b.idx
	return pos
}

func (b *Bits) SetPos(bit_pos int) {
	b.pos = bit_pos >> 3
	b.idx = bit_pos & 0x7
}
