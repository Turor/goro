package ui

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestPersonalExpDropTextScalesThousandths(t *testing.T) {
	s := &session.Session{PersonalExp: 100000, PersonalDrop: 150500}
	if got := personalExpDropText(s); got != "100%/150.5%" {
		t.Fatalf("rate text = %q", got)
	}
	a := statsWindowSnapshot(s)
	b := statsWindowSnapshot(&session.Session{})
	if a == b {
		t.Fatal("open status window would not rebuild when personal rates arrive")
	}
}
