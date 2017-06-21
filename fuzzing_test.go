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

func TestFuzzingIssue3(t *testing.T) {
	inputs := []string{
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
	}
	for _, input := range inputs {
		b := &bytesReadCloser{bytes.NewReader([]uint8(input))}
		_, _ = Decode(b)
	}
}
