package fetch_test

import (
	"bytes"
	"testing"

	"github.com/LYH2263/go-gitmirror/internal/fetch"
)

func TestPktLineRoundTrip(t *testing.T) {
	raw := fetch.EncodePktLine([]byte("refs/heads/main\n"))
	p, rest, flush, err := fetch.DecodePktLine(raw)
	if err != nil || flush || len(rest) != 0 {
		t.Fatalf("decode err=%v flush=%v rest=%d", err, flush, len(rest))
	}
	if !bytes.Equal(bytes.TrimSuffix(p, []byte{'\n'}), []byte("refs/heads/main")) {
		t.Fatalf("payload %q", p)
	}
	flushPkt := fetch.EncodePktFlush()
	_, _, flush, err = fetch.DecodePktLine(flushPkt)
	if err != nil || !flush {
		t.Fatalf("flush err=%v flush=%v", err, flush)
	}
}

func TestBuildWantHave(t *testing.T) {
	b := fetch.BuildWantHave([]string{"aaa"}, []string{"bbb"})
	lines, err := fetch.SplitPktLines(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 2 {
		t.Fatalf("lines=%d", len(lines))
	}
}
