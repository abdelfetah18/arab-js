package lsp

import (
	"errors"
	"unicode/utf8"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type Snapshot struct {
	uri     defines.DocumentUri
	content string
	version int
}

func NewSnapshot(uri defines.DocumentUri, content string, version int) *Snapshot {
	return &Snapshot{
		uri:     uri,
		content: content,
		version: version,
	}
}

func (s *Snapshot) PositionToIndex(pos defines.Position) (uint, error) {
	data := []byte(s.content)
	line := 0
	character := 0
	for i := 0; i < len(data); {
		ch, size := utf8.DecodeRune(data[i:])
		i += size
		character += size
		if ch == '\n' {
			line += 1
			character = 0
		}

		if line == int(pos.Line) && character == int(pos.Character) {
			return uint(i), nil
		}
	}
	return 0, errors.New("position out of range")
}
