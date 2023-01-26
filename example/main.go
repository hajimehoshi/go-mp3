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

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/oto/v2"

	"github.com/hajimehoshi/go-mp3"
)

func run() error {
	fp := "classic.mp3"
	if len(os.Args) > 1 {
		if nfp := os.Args[1]; nfp != "" {
			fp = nfp
		}
	}
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}

	c, ready, err := oto.NewContext(d.SampleRate(), 2, 2)
	if err != nil {
		return err
	}
	<-ready

	p := c.NewPlayer(d)
	defer p.Close()
	p.Play()

	fmt.Printf("Length: %d[bytes]\n", d.Length())
	fmt.Printf("Duration: %v\n", d.Duration().Round(time.Second))
	for {
		time.Sleep(time.Second)
		if !p.IsPlaying() {
			break
		}
		fmt.Printf("\rElapsed time: %v", d.ElapsedTime().Round(time.Second))
	}

	fmt.Println()
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
