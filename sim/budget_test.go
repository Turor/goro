package sim

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestBudgetConstantsMatchScript(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	script, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "scripts", "release", "budget-gate.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	for _, want := range []struct {
		needle string
		d      time.Duration
	}{
		{"CLEAN_BUILD_BUDGET_S:-15", CleanBuildBudget},
		{"TEST_SUITE_BUDGET_S:-15", TestSuiteBudget},
		{"HEADLESS_LOGIN_BUDGET_S:-3", HeadlessLoginBudget},
	} {
		if !strings.Contains(text, want.needle) {
			t.Errorf("budget-gate.sh missing %q for %s", want.needle, want.d)
		}
		if !strings.HasSuffix(want.needle, ":-"+itoa(int(want.d.Seconds()))) {
			t.Errorf("constant %s does not match needle %q", want.d, want.needle)
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
