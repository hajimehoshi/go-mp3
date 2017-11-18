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

package bits_test

import (
	"testing"

	. "github.com/hajimehoshi/go-mp3/internal/bits"
)

func TestBits(t *testing.T) {
	b1 := byte(85)  // 01010101
	b2 := byte(170) // 10101010
	b3 := byte(204) // 11001100
	b4 := byte(51)  // 00110011
	b := New([]byte{b1, b2, b3, b4})
	if b.Bits(1) != 0 {
		t.Fail()
	}
	if b.Bits(1) != 1 {
		t.Fail()
	}
	if b.Bits(1) != 0 {
		t.Fail()
	}
	if b.Bits(1) != 1 {
		t.Fail()
	}
	if b.Bits(8) != 90 /* 01011010 */ {
		t.Fail()
	}
	if b.Bits(12) != 2764 /* 101011001100 */ {
		t.Fail()
	}
}
