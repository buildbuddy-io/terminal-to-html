package terminal

import (
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
