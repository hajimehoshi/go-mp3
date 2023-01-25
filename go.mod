module github.com/hajimehoshi/go-mp3

go 1.14

require github.com/hajimehoshi/oto/v2 v2.3.1

// Fixed issues with internal module import paths for forked repositories.
replace github.com/hajimehoshi/go-mp3/internal/bits => ./internal/bits

replace github.com/hajimehoshi/go-mp3/internal/consts => ./internal/consts

replace github.com/hajimehoshi/go-mp3/internal/frame => ./internal/frame

replace github.com/hajimehoshi/go-mp3/internal/frameheader => ./internal/frameheader

replace github.com/hajimehoshi/go-mp3/internal/huffman => ./internal/huffman

replace github.com/hajimehoshi/go-mp3/internal/imdct => ./internal/imdct

replace github.com/hajimehoshi/go-mp3/internal/maindata => ./internal/maindata

replace github.com/hajimehoshi/go-mp3/internal/sideinfo => ./internal/sideinfo
