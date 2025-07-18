package terminal

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var currentLineForWritingTestCases = []struct {
	name     string
	input    string
	want     []string
	maxlines int
}{
	{
		name: "Test no index out of range panic",
		input: "\n",
		want: []string{"&nbsp;\n"},
		maxlines: 1,
	},
	{
		name: "Test scroll out first line",
		input: "a\n",
		want: []string{"a\n"},
		maxlines: 1,
	},
	{
		name: "Test scroll out several lines",
		input: "a\nb\nc\nd",
		want: []string{"a\n", "b\n"},
		maxlines: 2,
	},
}

func TestCurrentLineForWriting(t *testing.T) {
	for _, test := range currentLineForWritingTestCases {
		s, err := NewScreen(WithMaxSize(0, test.maxlines))
		got := []string{}
		s.ScrollOutFunc = func(line string) { got = append(got, line) }
		_ = s.currentLineForWriting()
		s.Write([]byte(test.input))
		if err != nil {
			t.Errorf("Failure for '%s':\nNewScreen returned an error: %s", test.name, err)
		}
		_ = s.currentLineForWriting()

		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Errorf(
				"Failure for '%s':\nscrolledOutFunc sequence of parameters diff (-got +want):\n%s",
				test.name,
				diff,
			)
		}
	}
}

var screenWriteTestCases = []struct{
	name string
	input []string
	want []screenLine
	opts []ScreenOption
} {
	{
		name: "Test leading newline",
		input: []string{
			"\n1234",
		},
		want: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '1',
					},
					{
						blob: '2',
					},
					{
						blob: '3',
					},
					{
						blob: '4',
					},
				},
			},
		},
	},
	{
		name: "Test double newline",
		input: []string{
			"\n\n12\n\n34\n",
		},
		want: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: '1',
					},
					{
						blob: '2',
					},
				},
			},
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: '3',
					},
					{
						blob: '4',
					},
				},
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
}

func TestScreenWrite(t *testing.T) {
	diffOpt := cmp.AllowUnexported(*new(screenLine), *new(node))
	for _, tc := range screenWriteTestCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewScreen()
			if err != nil {
				t.Errorf("NewScreen returned an error: %s", err)
			}
			for _, b := range tc.input {
				_, err := s.Write([]byte(b))
				if err != nil {
					t.Errorf("screen.Write returned an error: %s", err)
				}
			}
			if diff := cmp.Diff(s.screen, tc.want, diffOpt); diff != "" {
				t.Errorf(
					"(-got +want):\n%s",
					diff,
				)
			}
		})
	}
}

var scrollOutTestCases = []struct{
	name string
	input []string
	wantScrollOut string
	wantWindow []screenLine
	opts []ScreenOption
} {
	{
		name: "Test single blank line",
		input: []string{
			"\n",
		},
		wantScrollOut: "",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test two blank lines",
		input: []string{
			"\n\n",
		},
		wantScrollOut: "",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test leading newline",
		input: []string{
			"\n123456789",
		},
		wantScrollOut: "",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '1',
					},
					{
						blob: '2',
					},
					{
						blob: '3',
					},
					{
						blob: '4',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '5',
					},
					{
						blob: '6',
					},
					{
						blob: '7',
					},
					{
						blob: '8',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '9',
					},
				},
			},
		},
	},
	{
		name: "Test leading newline scrolls out",
		input: []string{
			"\n123456789\n",
		},
		wantScrollOut: "\n",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: []node{
					{
						blob: '1',
					},
					{
						blob: '2',
					},
					{
						blob: '3',
					},
					{
						blob: '4',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '5',
					},
					{
						blob: '6',
					},
					{
						blob: '7',
					},
					{
						blob: '8',
					},
				},
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: '9',
					},
				},
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test content on final line doesn't scroll out content",
		input: []string{
			"\n123456789\nabc",
		},
		wantScrollOut: "\n",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: []node{
					{
						blob: '1',
					},
					{
						blob: '2',
					},
					{
						blob: '3',
					},
					{
						blob: '4',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '5',
					},
					{
						blob: '6',
					},
					{
						blob: '7',
					},
					{
						blob: '8',
					},
				},
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: '9',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: 'a',
					},
					{
						blob: 'b',
					},
					{
						blob: 'c',
					},
				},
			},
		},
	},
	{
		name: "Test scroll out long line",
		input: []string{
			"\n123456789\nabc\ndef",
		},
		wantScrollOut: "\n1234",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: []node{
					{
						blob: '5',
					},
					{
						blob: '6',
					},
					{
						blob: '7',
					},
					{
						blob: '8',
					},
				},
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: '9',
					},
				},
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: 'a',
					},
					{
						blob: 'b',
					},
					{
						blob: 'c',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: 'd',
					},
					{
						blob: 'e',
					},
					{
						blob: 'f',
					},
				},
			},
		},
	},
	{
		name: "Test scroll out too long line",
		input: []string{
			"abcdefghijklmnopq\n0123",
		},
		wantScrollOut: "abcdefgh",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: []node{
					{
						blob: 'i',
					},
					{
						blob: 'j',
					},
					{
						blob: 'k',
					},
					{
						blob: 'l',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: 'm',
					},
					{
						blob: 'n',
					},
					{
						blob: 'o',
					},
					{
						blob: 'p',
					},
				},
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: 'q',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '0',
					},
					{
						blob: '1',
					},
					{
						blob: '2',
					},
					{
						blob: '3',
					},
				},
			},
		},
	},
	{
		name: "Test scroll out much too long line",
		input: []string{
			"abcdefghijklmnopqrstuvwxyz\n0123",
		},
		wantScrollOut: "abcdefghijklmnop",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: []node{
					{
						blob: 'q',
					},
					{
						blob: 'r',
					},
					{
						blob: 's',
					},
					{
						blob: 't',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: 'u',
					},
					{
						blob: 'v',
					},
					{
						blob: 'w',
					},
					{
						blob: 'x',
					},
				},
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: 'y',
					},
					{
						blob: 'z',
					},
				},
			},
			{
				newline: false,
				nodes: []node{
					{
						blob: '0',
					},
					{
						blob: '1',
					},
					{
						blob: '2',
					},
					{
						blob: '3',
					},
				},
			},
		},
	},
	{
		name: "Test clear full window and write",
		input: []string{
			"abcdefghijklmnop\x1b[2J\x1b[;H012",
		},
		wantScrollOut: "",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: []node{
					{
						blob: '0',
					},
					{
						blob: '1',
					},
					{
						blob: '2',
					},
				},
			},
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test clear partial window and write",
		input: []string{
			"abcdef\x1b[2J\x1b[;H012",
		},
		wantScrollOut: "",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: []node{
					{
						blob: '0',
					},
					{
						blob: '1',
					},
					{
						blob: '2',
					},
				},
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
}

func TestScrollout(t *testing.T) {
	diffOpt := cmp.AllowUnexported(*new(screenLine), *new(node))
	sizeOpt := WithMaxSize(4, 4)

	for _, tc := range scrollOutTestCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewScreen(sizeOpt, WithANSIRenderer(), WithRealWindow())
			sb := strings.Builder{}
			s.ScrollOutFunc = func(line string) {
				sb.Write([]byte(line))
				t.Logf("Scrollout called with line:\n---------\n%s\n----------\n", line)
			} 
			if err != nil {
				t.Errorf("NewScreen returned an error: %s", err)
			}
			for _, b := range tc.input {
				_, err := s.Write([]byte(b))
				if err != nil {
					t.Errorf("screen.Write returned an error: %s", err)
				}
			}
			if diff := cmp.Diff(s.screen, tc.wantWindow, diffOpt); diff != "" {
				renderedActual, _ := s.AsANSI()
				e, err := NewScreen(sizeOpt, WithANSIRenderer())
				if err != nil {
					t.Errorf("NewScreen returned an error: %s", err)
				}
				e.screen = tc.wantWindow
				renderedExpected, _ := e.AsANSI()
				t.Errorf(
					"window (-got +want):\n%s\nrendered actual:\n----------\n%s\n----------\nrendered expected:\n----------\n%s\n----------\n",
					diff,
					renderedActual,
					renderedExpected,
				)
			}
			if diff := cmp.Diff(sb.String(), tc.wantScrollOut, diffOpt); diff != "" {
				t.Errorf(
					"scrollout (-got +want):\n%s",
					diff,
				)
			}
		})
	}
}

var UnlimitedWindowTestCases = []struct{
	name string
	input []string
	wantWindow []screenLine
	opts []ScreenOption
} {
	{
		name: "Test single blank line",
		input: []string{
			"\n",
		},
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test two blank lines",
		input: []string{
			"\n\n",
		},
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test two blank lines surrounding text",
		input: []string{
			"\nabc\n",
		},
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: true,
				nodes: []node{
					{
						blob: 'a',
					},
					{
						blob: 'b',
					},
					{
						blob: 'c',
					},
				},
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
}

func TestUnlimitedWindowSize(t *testing.T) {
	diffOpt := cmp.AllowUnexported(*new(screenLine), *new(node))
	sizeOpt := WithMaxSize(100, 0)

	for _, tc := range UnlimitedWindowTestCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewScreen(sizeOpt, WithANSIRenderer(), WithRealWindow())
			if err != nil {
				t.Errorf("NewScreen returned an error: %s", err)
			}
			for _, b := range tc.input {
				_, err := s.Write([]byte(b))
				if err != nil {
					t.Errorf("screen.Write returned an error: %s", err)
				}
			}
			if diff := cmp.Diff(s.screen, tc.wantWindow, diffOpt); diff != "" {
				renderedActual, _ := s.AsANSI()
				e, err := NewScreen(sizeOpt, WithANSIRenderer())
				if err != nil {
					t.Errorf("NewScreen returned an error: %s", err)
				}
				e.screen = tc.wantWindow
				renderedExpected, _ := e.AsANSI()
				t.Errorf(
					"window (-got +want):\n%s\nrendered actual:\n----------\n%s\n----------\nrendered expected:\n----------\n%s\n----------\n",
					diff,
					renderedActual,
					renderedExpected,
				)
			}
		})
	}
}

var WindowHeight1TestCases = []struct{
	name string
	input []string
	wantWindow []screenLine
	wantScrollOut string
	opts []ScreenOption
} {
	{
		name: "Test single blank line",
		input: []string{
			"\n",
		},
		wantScrollOut: "\n",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test two blank lines",
		input: []string{
			"\n\n",
		},
		wantScrollOut: "\n\n",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test two blank lines surrounding text",
		input: []string{
			"\nabc\n",
		},
		wantScrollOut: "\nabc\n",
		wantWindow: []screenLine {
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
}

func TestWindowHeight1(t *testing.T) {
	diffOpt := cmp.AllowUnexported(*new(screenLine), *new(node))
	sizeOpt := WithMaxSize(100, 1)

	for _, tc := range WindowHeight1TestCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewScreen(sizeOpt, WithANSIRenderer(), WithRealWindow())
			if err != nil {
				t.Errorf("NewScreen returned an error: %s", err)
			}
			sb := strings.Builder{}
			s.ScrollOutFunc = func(line string) {
				sb.Write([]byte(line))
				t.Logf("Scrollout called with line:\n---------\n%s\n----------\n", line)
			} 
			for _, b := range tc.input {
				_, err := s.Write([]byte(b))
				if err != nil {
					t.Errorf("screen.Write returned an error: %s", err)
				}
			}
			if diff := cmp.Diff(s.screen, tc.wantWindow, diffOpt); diff != "" {
				renderedActual, _ := s.AsANSI()
				e, err := NewScreen(sizeOpt, WithANSIRenderer())
				if err != nil {
					t.Errorf("NewScreen returned an error: %s", err)
				}
				e.screen = tc.wantWindow
				renderedExpected, _ := e.AsANSI()
				t.Errorf(
					"window (-got +want):\n%s\nrendered actual:\n----------\n%s\n----------\nrendered expected:\n----------\n%s\n----------\n",
					diff,
					renderedActual,
					renderedExpected,
				)
			}
			if diff := cmp.Diff(sb.String(), tc.wantScrollOut, diffOpt); diff != "" {
				t.Errorf(
					"scrollout (-got +want):\n%s",
					diff,
				)
			}
		})
	}
}

var WindowHeight2TestCases = []struct{
	name string
	input []string
	wantWindow []screenLine
	wantScrollOut string
	opts []ScreenOption
} {
	{
		name: "Test single blank line",
		input: []string{
			"\n",
		},
		wantScrollOut: "",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test two blank lines",
		input: []string{
			"\n\n",
		},
		wantScrollOut: "\n",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: make([]node, 0),
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
	{
		name: "Test two blank lines surrounding text",
		input: []string{
			"\nabc\n",
		},
		wantScrollOut: "\n",
		wantWindow: []screenLine {
			{
				newline: true,
				nodes: []node{
					{
						blob: 'a',
					},
					{
						blob: 'b',
					},
					{
						blob: 'c',
					},
				},
			},
			{
				newline: false,
				nodes: make([]node, 0),
			},
		},
	},
}

func TestWindowHeight2(t *testing.T) {
	diffOpt := cmp.AllowUnexported(*new(screenLine), *new(node))
	sizeOpt := WithMaxSize(100, 2)

	for _, tc := range WindowHeight2TestCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewScreen(sizeOpt, WithANSIRenderer(), WithRealWindow())
			if err != nil {
				t.Errorf("NewScreen returned an error: %s", err)
			}
			sb := strings.Builder{}
			s.ScrollOutFunc = func(line string) {
				sb.Write([]byte(line))
				t.Logf("Scrollout called with line:\n---------\n%s\n----------\n", line)
			} 
			for _, b := range tc.input {
				_, err := s.Write([]byte(b))
				if err != nil {
					t.Errorf("screen.Write returned an error: %s", err)
				}
			}
			if diff := cmp.Diff(s.screen, tc.wantWindow, diffOpt); diff != "" {
				renderedActual, _ := s.AsANSI()
				e, err := NewScreen(sizeOpt, WithANSIRenderer())
				if err != nil {
					t.Errorf("NewScreen returned an error: %s", err)
				}
				e.screen = tc.wantWindow
				renderedExpected, _ := e.AsANSI()
				t.Errorf(
					"window (-got +want):\n%s\nrendered actual:\n----------\n%s\n----------\nrendered expected:\n----------\n%s\n----------\n",
					diff,
					renderedActual,
					renderedExpected,
				)
			}
			if diff := cmp.Diff(sb.String(), tc.wantScrollOut, diffOpt); diff != "" {
				t.Errorf(
					"scrollout (-got +want):\n%s",
					diff,
				)
			}
		})
	}
}
