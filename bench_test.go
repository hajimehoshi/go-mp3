// Copyright 2017 The go-mp3 Authors
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
	"io"
	"io/ioutil"
	"testing"
)

func BenchmarkDecode(b *testing.B) {
	buf, err := ioutil.ReadFile("example/classic.mp3")
	if err != nil {
		b.Fatal(err)
	}
	src := &bytesReadCloser{bytes.NewReader(buf)}
	for i := 0; i < b.N; i++ {
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			b.Fatal(err)
		}
		d, err := NewDecoder(src)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := ioutil.ReadAll(d); err != nil {
			b.Fatal(err)
		}
	}
}
