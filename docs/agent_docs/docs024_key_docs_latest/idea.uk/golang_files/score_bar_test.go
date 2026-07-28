package main

// score_bar_test.go — the report's score bars.
//
// These render inside an email, where there is no console to check and no way to
// push a fix to something already delivered. The failure modes below are all ones
// that look fine in a browser and wrong in a mail client.

import (
	"strings"
	"testing"
)

const (
	tFill  = "#15243d"
	tTrack = "#e2e6ec"
	tLab   = "#36424f"
	tVal   = "#707b88"
)

// A zero-width <td> is rendered inconsistently — several clients apply a minimum
// width, which would draw a visible bar for a score of zero. The empty side must
// be omitted entirely, not emitted as width:0%.
func TestScoreBarZeroDrawsNoFill(t *testing.T) {
	got := scoreBar("people will pay", 0, 5, tFill, tTrack, tLab, tVal)
	if strings.Contains(got, "width:0%") {
		t.Errorf("zero-width cell emitted; some clients give it a minimum width:\n%s", got)
	}
	if strings.Contains(got, "background:"+tFill) {
		t.Errorf("a score of 0 drew a filled bar:\n%s", got)
	}
	if !strings.Contains(got, "0/5") {
		t.Errorf("the number must survive even with no bar:\n%s", got)
	}
}

// The mirror: a full score must not emit a zero-width empty cell either.
func TestScoreBarFullDrawsNoRemainder(t *testing.T) {
	got := scoreBar("built to last", 5, 5, tFill, tTrack, tLab, tVal)
	if strings.Contains(got, "width:0%") {
		t.Errorf("zero-width remainder cell emitted:\n%s", got)
	}
	if !strings.Contains(got, "width:100%;background:"+tFill) {
		t.Errorf("a full score should fill the track:\n%s", got)
	}
}

// A model returning a score above the maximum would otherwise produce a bar wider
// than its track and blow the table layout out.
func TestScoreBarClampsOutOfRange(t *testing.T) {
	for _, c := range []struct {
		in      int
		wantTxt string
	}{
		{7, "5/5"},
		{-2, "0/5"},
	} {
		got := scoreBar("hard to copy", c.in, 5, tFill, tTrack, tLab, tVal)
		if !strings.Contains(got, c.wantTxt) {
			t.Errorf("scoreBar(%d) did not clamp to %s:\n%s", c.in, c.wantTxt, got)
		}
		if strings.Contains(got, "width:140%") || strings.Contains(got, "width:-") {
			t.Errorf("scoreBar(%d) produced an out-of-range width:\n%s", c.in, got)
		}
	}
}

// The bar is an enhancement; the number is the carrier. A client that flattens
// every style must still leave the score readable, so the digits are always in
// the markup rather than implied by the bar's length.
func TestScoreBarAlwaysCarriesTheNumber(t *testing.T) {
	for v := 0; v <= 5; v++ {
		got := scoreBar("reusable elsewhere", v, 5, tFill, tTrack, tLab, tVal)
		if !strings.Contains(got, "reusable elsewhere") {
			t.Errorf("label missing at v=%d", v)
		}
		if !strings.Contains(got, itoaTest(v)+"/5") {
			t.Errorf("value %d/5 missing from markup:\n%s", v, got)
		}
	}
}

// Labels are our own strings today, but the same helper renders anything passed
// to it — escaping is asserted so that stays true if a model-supplied label is
// ever fed through.
func TestScoreBarEscapesItsLabel(t *testing.T) {
	got := scoreBar(`<script>alert(1)</script>`, 3, 5, tFill, tTrack, tLab, tVal)
	if strings.Contains(got, "<script>") {
		t.Errorf("label was not escaped:\n%s", got)
	}
}

// An overall score uses a different maximum; the percentage must follow max, not
// assume 5.
func TestScoreBarHonoursMaximum(t *testing.T) {
	got := scoreBar("overall", 20, 25, tFill, tTrack, tLab, tVal)
	if !strings.Contains(got, "width:80%") {
		t.Errorf("20/25 should fill 80%% of the track:\n%s", got)
	}
	if !strings.Contains(got, "20/25") {
		t.Errorf("value should read 20/25:\n%s", got)
	}
}

func itoaTest(i int) string {
	if i == 0 {
		return "0"
	}
	var d []byte
	for i > 0 {
		d = append([]byte{byte('0' + i%10)}, d...)
		i /= 10
	}
	return string(d)
}
