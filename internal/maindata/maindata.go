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

package maindata

import (
	"fmt"
	"io"

	"github.com/hajimehoshi/go-mp3/internal/bits"
	"github.com/hajimehoshi/go-mp3/internal/consts"
)

type FullReader interface {
	ReadFull([]byte) (int, error)
}

// A MainData is MPEG1 Layer 3 Main Data.
type MainData struct {
	ScalefacL [2][2][21]int      // 0-4 bits
	ScalefacS [2][2][12][3]int   // 0-4 bits
	Is        [2][2][576]float32 // Huffman coded freq. lines
}

func Read(source FullReader, prev *bits.Bits, size int, offset int) (*bits.Bits, error) {
	if size > 1500 {
		return nil, fmt.Errorf("mp3: size = %d", size)
	}
	// Check that there's data available from previous frames if needed
	if prev != nil && offset > prev.LenInBytes() {
		// No, there is not, so we skip decoding this frame, but we have to
		// read the main_data bits from the bitstream in case they are needed
		// for decoding the next frame.
		buf := make([]byte, size)
		if n, err := source.ReadFull(buf); n < size {
			if err == io.EOF {
				return nil, &consts.UnexpectedEOF{"maindata.Read (1)"}
			}
			return nil, err
		}
		// TODO: Define a special error and enable to continue the next frame.
		return bits.Append(prev, buf), nil
	}
	// Copy data from previous frames
	vec := []byte{}
	if prev != nil {
		vec = prev.Tail(offset)
	}
	// Read the main_data from file
	buf := make([]byte, size)
	if n, err := source.ReadFull(buf); n < size {
		if err == io.EOF {
			return nil, &consts.UnexpectedEOF{"maindata.Read (2)"}
		}
		return nil, err
	}
	return bits.New(append(vec, buf...)), nil
}
