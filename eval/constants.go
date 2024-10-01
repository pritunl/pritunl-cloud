package eval

import (
	"github.com/dropbox/godropbox/container/set"
)

type Equal struct{}
type NotEqual struct{}
type Less struct{}
type LessEqual struct{}
type Greater struct{}
type GreaterEqual struct{}
type If struct{}
type And struct{}
type Or struct{}
type For struct{}
type Then struct{}

const (
	StatementMaxLength = 1024
	StatementMaxParts  = 30
)

var StatementSafeCharacters = set.NewSet(
	'a',
	'b',
	'c',
	'd',
	'e',
	'f',
	'g',
	'h',
	'i',
	'j',
	'k',
	'l',
	'm',
	'n',
	'o',
	'p',
	'q',
	'r',
	's',
	't',
	'u',
	'v',
	'w',
	'x',
	'y',
	'z',
	'A',
	'B',
	'C',
	'D',
	'E',
	'F',
	'G',
	'H',
	'I',
	'J',
	'K',
	'L',
	'M',
	'N',
	'O',
	'P',
	'Q',
	'R',
	'S',
	'T',
	'U',
	'V',
	'W',
	'X',
	'Y',
	'Z',
	'0',
	'1',
	'2',
	'3',
	'4',
	'5',
	'6',
	'7',
	'8',
	'9',
	' ',
	'.',
	'_',
	'-',
	'=',
	'>',
	'<',
	'!',
	'\'',
	'(',
	')',
)
