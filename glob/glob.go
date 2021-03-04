// package glob implements translating simple glob patterns into go regexps.
package glob

import (
	"regexp"
	"strings"
)

type compiler struct {
	text         []rune
	ch           rune
	pos, readpos int
}

// Compile takes in a glob pattern as argument and compiles it into a go regexp.
func Compile(s string) (*regexp.Regexp, error) {
	c := &compiler{
		text: []rune(s),
	}
	c.read()
	return regexp.Compile(c.compile())
}

// MustCompile is same as Compile but instead of returning any compilation errors, it panics.
func MustCompile(s string) *regexp.Regexp {
	r, err := Compile(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (c *compiler) read() {
	if c.readpos >= len(c.text) {
		c.ch = 0
	} else {
		c.ch = c.text[c.readpos]
	}
	c.pos = c.readpos
	c.readpos += 1
}

func (c *compiler) peek() rune {
	if c.readpos >= len(c.text) {
		return 0
	}
	return c.text[c.readpos]
}

func (c *compiler) compile() string {
	var buff strings.Builder
	buff.WriteString("(?i)^")
LOOP:
	for {
		switch c.ch {
		case '*':
			buff.WriteString("[a-zA-Z0-9_]*")
		case '?':
			buff.WriteString(".*")
		case '\\':
			buff.WriteRune(c.ch)
			c.read()
			buff.WriteRune(c.ch)
		case '.':
			buff.WriteString("\\.")
		case 0:
			break LOOP
		default:
			buff.WriteRune(c.ch)
		}
		c.read()
	}
	buff.WriteRune('$')
	return buff.String()
}
