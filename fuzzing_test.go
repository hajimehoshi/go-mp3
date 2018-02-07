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
	"bytes"
	"testing"
)

type bytesReadCloser struct {
	*bytes.Reader
}

func (b *bytesReadCloser) Close() error {
	return nil
}

func TestFuzzing(t *testing.T) {
	inputs := []string{
		// #3
		"\xff\xfa500000000000\xff\xff0000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"0000",
		"\xff\xfb\x100004000094\xff000000" +
			"00000000000000000000" +
			"00\u007f0\xff\xee\u007f\xff\xee\u007f\xff\xff\u007f\xff\xff\xee\u007f\xff\xff0" +
			"\xff\xff00\xff\xee\u007f\xff0000\u007f00\xff00\xee0" +
			"000\xff000\xff\xff\xee\u007f0\xff0000\u007f\xff0" +
			"00\xff0",
		"\xff\xfb\x100004000094\xff000000" +
			"00000000000000000000" +
			"00\u007f0\xff\xee\u007f\xff\xee\u007f\xff\xff\u007f\xff\xff\xee\u007f\xff\xff\u007f" +
			"\xff\xff\u007f0\xff\xee\u007f\xff0000\u007f00\xff\xff\xee\xee0" +
			"0\xee\u007f\xff000\xff\xff\xee\u007f0\xff0000\u007f\xff0" +
			"0\xff\xff0",
		"\xff\xfa\x1000000000000000000" +
			"00000000000000000000" +
			"000000000000000000\xff\xff" +
			"0\u007f\xff\xff\u007f\xff\xff\u007f\xff\xff\xfc0\xff\xef\xbf0\xef\xbf00" +
			"0\xff\xee\u007f\xff\xff\u007f\xff\xff\xee\u007f\xff\xff\u007f\xff\xff\u007f\xff00" +
			"\xff\xff00",
		"\xff\xfa00000031000000000n" +
			"s0f00000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000\u007f\xff\xff000\xff\xee",
		"\xff\xfa\x1000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000\xbf0\xef\xbf00" +
			"0\xff\xee0\xff\xff\u007f\xff\xff\xee\u007f\xff\xff\u007f\xff\xff\u007f\xff00" +
			"\xff0\xee0",
		"\xff\xfa\x100000050000000000\u007f" +
			"00000000000000000000" +
			"0000000000\xee\u007f0\xff\xff\xff\xff\u007f\xff\xff" +
			"\xee\u007f\xff\xff\u007f\xff\xff\u007f\xff\xff\xfc\xee\xff\xef\xbf0\xef\xbf00" +
			"0\xff\xee\u007f\xff\xff\u007f\xff\xff\xee\u007f\xff\xff\u007f\xff\xff\u007f\xff0\t" +
			"\xff\xff\xee\xee",
		// #22
		"\xff\xfa%00000000000000000" +
			"000000000000s0000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000",
		// #23
		"\xff\xfb%S000000v000\x00\x010000" +
			"00000000000000000000" +
			"0000\xf4000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000",
		// #24
		"\xff\xfb0x000000\xf9000\x00\x030000" +
			"000000000000\xf70000000" +
			"\x900000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"0000000000000",
	}
	for _, input := range inputs {
		b := &bytesReadCloser{bytes.NewReader([]byte(input))}
		_, _ = NewDecoder(b)
	}
}
