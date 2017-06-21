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
	}
	for i, input := range inputs {
		b := &bytesReadCloser{bytes.NewReader([]uint8(input))}
		if _, err := Decode(b); err == nil {
			t.Errorf("test %d must fail but not", i)
		}
	}
}
