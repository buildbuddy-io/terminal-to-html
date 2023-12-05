package terminal

import (
	"bytes"
	"math"
	"strconv"
	"strings"
)

func NewScreen() *Screen {
	return &Screen{style: &emptyStyle}
}

// A terminal 'Screen'. Current cursor position, cursor style, and characters
type Screen struct {
	x      int
	y      int
	screen []screenLine
	style  *style
}

type screenLine struct {
	nodes []node

	// metadata is { namespace => { key => value, ... }, ... }
	// e.g. { "bk" => { "t" => "1234" } }
	metadata map[string]map[string]string
}

const screenEndOfLine = -1
const screenStartOfLine = 0
const extraLinesToBuffer = 100

// Clear part (or all) of a line on the screen
func (s *Screen) clear(y int, xStart int, xEnd int) {
	if len(s.screen) <= y {
		return
	}

	line := s.screen[y]
	if xStart == screenStartOfLine && xEnd == screenEndOfLine {
		// Blank the entire line, but keep any existing line metadata.
		s.screen[y].nodes = make([]node, 0, 80)
	} else {
		if xEnd == screenEndOfLine {
			xEnd = len(line.nodes) - 1
		}
		// TODO: optimise clear-to-end-of-line by truncating line.nodes?
		for i := xStart; i <= xEnd && i < len(line.nodes); i++ {
			line.nodes[i] = emptyNode
		}
	}
}

// "Safe" parseint for parsing ANSI instructions
func ansiInt(s string) int {
	if s == "" {
		return 1
	}
	i, _ := strconv.ParseInt(s, 10, 8)
	return int(i)
}

// Move the cursor up, if we can
func (s *Screen) up(i string) {
	s.y -= ansiInt(i)
	s.y = int(math.Max(0, float64(s.y)))
}

// Move the cursor down
func (s *Screen) down(i string) {
	s.y += ansiInt(i)
}

// Move the cursor forward on the line
func (s *Screen) forward(i string) {
	s.x += ansiInt(i)
}

// Move the cursor backward, if we can
func (s *Screen) backward(i string) {
	s.x -= ansiInt(i)
	s.x = int(math.Max(0, float64(s.x)))
}

func (s *Screen) getCurrentLineForWriting() *screenLine {
	// Add rows to our screen if necessary
	for i := len(s.screen); i <= s.y; i++ {
		s.screen = append(s.screen, screenLine{nodes: make([]node, 0, 80)})
	}

	line := &s.screen[s.y]

	// Add columns if currently shorter than the cursor's x position
	for i := len(line.nodes); i <= s.x; i++ {
		line.nodes = append(line.nodes, emptyNode)
	}
	return line
}

// Write a character to the screen's current X&Y, along with the current screen style
func (s *Screen) write(data rune) {
	line := s.getCurrentLineForWriting()
	line.nodes[s.x] = node{blob: data, style: s.style}
}

// Append a character to the screen
func (s *Screen) append(data rune) {
	s.write(data)
	s.x++
}

// Append multiple characters to the screen
func (s *Screen) appendMany(data []rune) {
	for _, char := range data {
		s.append(char)
	}
}

func (s *Screen) appendElement(i *element) {
	line := s.getCurrentLineForWriting()
	line.nodes[s.x] = node{style: s.style, elem: i}
	s.x++
}

// Set non-existing line metadata. Merges the provided data into any existing
// metadata for the current line, keeping existing data when keys collide.
func (s *Screen) setnxLineMetadata(namespace string, data map[string]string) {
	line := s.getCurrentLineForWriting()
	if line.metadata == nil {
		line.metadata = make(map[string]map[string]string)
	}
	if ns, nsExists := line.metadata[namespace]; nsExists {
		// set keys that don't already exist
		for k, v := range data {
			if _, kExists := ns[k]; !kExists {
				ns[k] = v
			}
		}
	} else {
		// namespace did not exist, set all data
		line.metadata[namespace] = data
	}
}

// Apply color instruction codes to the screen's current style
func (s *Screen) color(i []string) {
	s.style = s.style.color(i)
}

// Apply an escape sequence to the screen
func (s *Screen) applyEscape(code rune, instructions []string) {
	if len(instructions) == 0 {
		// Ensure we always have a first instruction
		instructions = []string{""}
	}

	switch code {
	case 'M':
		s.color(instructions)
	case 'G':
		s.x = 0
	// "Erase in Display"
	case 'J':
		switch instructions[0] {
		// "erase from current position to end (inclusive)"
		case "0", "":
			// This line should be equivalent to K0
			s.clear(s.y, s.x, screenEndOfLine)
			// Truncate the screen below the current line
			if len(s.screen) > s.y {
				s.screen = s.screen[:s.y+1]
			}
		// "erase from beginning to current position (inclusive)"
		case "1":
			// This line should be equivalent to K1
			s.clear(s.y, screenStartOfLine, s.x)
			// Truncate the screen above the current line
			if len(s.screen) > s.y {
				s.screen = s.screen[s.y+1:]
			}
			// Adjust the cursor position to compensate
			s.y = 0
		// 2: "erase entire display", 3: "erase whole display including scroll-back buffer"
		// Given we don't have a scrollback of our own, we treat these as equivalent
		case "2", "3":
			s.screen = nil
			s.x = 0
			s.y = 0
		}
	// "Erase in Line"
	case 'K':
		switch instructions[0] {
		case "0", "":
			s.clear(s.y, s.x, screenEndOfLine)
		case "1":
			s.clear(s.y, screenStartOfLine, s.x)
		case "2":
			s.clear(s.y, screenStartOfLine, screenEndOfLine)
		}
	case 'A':
		s.up(instructions[0])
	case 'B':
		s.down(instructions[0])
	case 'C':
		s.forward(instructions[0])
	case 'D':
		s.backward(instructions[0])
	}
}

// Parse ANSI input, populate our screen buffer with nodes
func (s *Screen) parse(ansi []byte) {
	s.style = &emptyStyle

	ParseANSIToScreen(s, ansi)
}

func (s *Screen) AsHTML() []byte {
	var lines []string

	for _, line := range s.screen {
		lines = append(lines, outputLineAsHTML(line))
	}

	return []byte(strings.Join(lines, "\n"))
}

// asPlainText renders the screen without any ANSI style etc.
func (s *Screen) AsPlainText() string {
	var buf bytes.Buffer
	for i, line := range s.screen {
		for _, node := range line.nodes {
			if node.elem == nil {
				buf.WriteRune(node.blob)
			}
		}
		if i < len(s.screen)-1 {
			buf.WriteRune('\n')
		}
	}
	return strings.TrimRight(buf.String(), " \t")
}

func (s *Screen) AsANSI() []byte {
	var lines []string

	for _, line := range s.screen {
		lines = append(lines, outputLineAsANSI(line))
	}

	return []byte(strings.Join(lines, "\n"))
}

func (s *Screen) newLine() {
	s.x = 0
	s.y++
}

func (s *Screen) revNewLine() {
	if s.y > 0 {
		s.y--
	}
}

func (s *Screen) carriageReturn() {
	s.x = 0
}

func (s *Screen) backspace() {
	if s.x > 0 {
		s.x--
	}
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (s *Screen) FlushLinesFromTop(numLinesToRetain int) []byte {
  numLinesToFlush := len(s.screen) - numLinesToRetain
  if numLinesToFlush > s.y {
    // log.Warningf("Screen attempted to pop line containing the current cursor position. Attempted to retain too few lines by %d line(s).", extraLines-s.y)
    numLinesToFlush = s.y
  }
  if numLinesToFlush < 1 {
    return []byte{}
  }
  flushedLines := (&Screen{screen: s.screen[:numLinesToFlush]}).AsANSI()
  s.screen = s.screen[numLinesToFlush:]

  // Hang onto the lines we just flushed instead of allocating them again.
  for i := 0; i + numLinesToFlush < len(s.screen); i++ {
	s.screen[i] = s.screen[i + numLinesToFlush]
  }
  s.screen = s.screen[:len(s.screen) - numLinesToFlush:min(len(s.screen), numLinesToRetain + extraLinesToBuffer)]
  s.y -= numLinesToFlush
  return flushedLines
}
