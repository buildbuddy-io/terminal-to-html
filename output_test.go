package terminal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestScreenLineAsHTML_Interleaving(t *testing.T) {
	// ANSI escapes can come in any order, but it is invalid to interleave HTML
	// tags.

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "a span /a /span",
			input: "five \x1b]8;;http://example.com\x1b\\six \x1b[35mseven \x1b]8;;\x1b\\eight\x1b[0m",
			want:  `five <a href="http://example.com">six <span class="term-fg35">seven </span></a><span class="term-fg35">eight</span>` + "\n",
		},
		{
			name:  "span a /span /a",
			input: "five \x1b[35msix \x1b]8;;http://example.com\x1b\\seven \x1b[0meight\x1b]8;;\x1b\\",
			want:  `five <span class="term-fg35">six <a href="http://example.com">seven </a></span><a href="http://example.com">eight</a>` + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := NewScreen()
			if err != nil {
				t.Fatalf("NewScreen() = %v", err)
			}
			s.Write([]byte(test.input))
			if len(s.screen) != 1 {
				t.Fatalf("len(s.screen) = %d, want 1", len(s.screen))
			}

			got := lineToHTML(s.screen[:1], true)
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("lineToHTML(s.screen[:1], true) diff (-got +want):\n%s", diff)
			}
		})
	}
}

// The <time> and <span class> tags are written directly rather than through
// html/template, on the proof that their interpolated values never contain
// characters the template escaper would touch. These tests pin the exact
// bytes of those tags, across every shape the values can take.

func TestLineToHTML_TimeTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "millisecond fraction",
			input: "\x1b_bk;t=123\x07x",
			want:  `<time datetime="1970-01-01T00:00:00.123Z">1970-01-01T00:00:00.123Z</time>x` + "\n",
		},
		{
			// The .999 format drops trailing zeros from the fraction.
			name:  "fraction with trailing zero",
			input: "\x1b_bk;t=120\x07x",
			want:  `<time datetime="1970-01-01T00:00:00.12Z">1970-01-01T00:00:00.12Z</time>x` + "\n",
		},
		{
			name:  "single-digit fraction",
			input: "\x1b_bk;t=100\x07x",
			want:  `<time datetime="1970-01-01T00:00:00.1Z">1970-01-01T00:00:00.1Z</time>x` + "\n",
		},
		{
			// …and the dot itself when the fraction is zero.
			name:  "whole second",
			input: "\x1b_bk;t=1000\x07x",
			want:  `<time datetime="1970-01-01T00:00:01Z">1970-01-01T00:00:01Z</time>x` + "\n",
		},
		{
			name:  "modern date",
			input: "\x1b_bk;t=1500000000123\x07x",
			want:  `<time datetime="2017-07-14T02:40:00.123Z">2017-07-14T02:40:00.123Z</time>x` + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := NewScreen()
			if err != nil {
				t.Fatalf("NewScreen() = %v", err)
			}
			s.Write([]byte(test.input))
			if len(s.screen) != 1 {
				t.Fatalf("len(s.screen) = %d, want 1", len(s.screen))
			}

			got := lineToHTML(s.screen[:1], true)
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("lineToHTML(s.screen[:1], true) diff (-got +want):\n%s", diff)
			}
		})
	}
}

func TestLineToHTML_SpanClassJoin(t *testing.T) {
	// Several styles at once, so the class attribute exercises the joined,
	// multi-class form: fg and bg colors first, then attributes.
	input := "\x1b[1;3;35;41mx"
	want := `<span class="term-fg35 term-bg41 term-fg1 term-fg3">x</span>` + "\n"

	s, err := NewScreen()
	if err != nil {
		t.Fatalf("NewScreen() = %v", err)
	}
	s.Write([]byte(input))
	if len(s.screen) != 1 {
		t.Fatalf("len(s.screen) = %d, want 1", len(s.screen))
	}

	got := lineToHTML(s.screen[:1], true)
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("lineToHTML(s.screen[:1], true) diff (-got +want):\n%s", diff)
	}
}
