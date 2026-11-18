package res

import (
	"testing"
	"time"

	glua "github.com/yuin/gopher-lua"
)

func TestGopherLuaTableCycleDoesNotOverflow(t *testing.T) {
	state := glua.NewState(glua.Options{SkipOpenLibs: true})
	defer state.Close()
	tab := state.NewTable()
	tab.RawSetString("t", tab)
	tab.RawSetString("n", glua.LNumber(1))
	converted, ok := gopherLuaToLuaValue(tab)
	if !ok || converted.kind != luaTable {
		t.Fatalf("converted = %+v ok=%v", converted, ok)
	}
	if converted.table["n"].num != 1 {
		t.Fatalf("n = %+v", converted.table["n"])
	}
	cycle, ok := converted.table["t"]
	if !ok || cycle.kind != luaTable {
		t.Fatalf("cycle = %+v ok=%v", cycle, ok)
	}
	// The cycle must share the same map, not recurse into a copy.
	if cycle.table["n"].num != 1 {
		t.Fatalf("cycle.n = %+v", cycle.table["n"])
	}
}

func TestLuaValueTableCycleDoesNotOverflow(t *testing.T) {
	m := map[interface{}]luaValue{}
	v := luaValue{kind: luaTable, table: m}
	m["t"] = v
	m["n"] = luaValue{kind: luaNumber, num: 7}
	state := glua.NewState(glua.Options{SkipOpenLibs: true})
	defer state.Close()
	got := luaValueToGopherLua(state, v)
	tab, ok := got.(*glua.LTable)
	if !ok {
		t.Fatalf("got %T", got)
	}
	if num, ok := tab.RawGetString("n").(glua.LNumber); !ok || num != 7 {
		t.Fatalf("n = %v", tab.RawGetString("n"))
	}
	if child, ok := tab.RawGetString("t").(*glua.LTable); !ok || child != tab {
		t.Fatalf("cycle not preserved: %v", tab.RawGetString("t"))
	}
}

func TestExecuteLuaSourceBudgetOnInfiniteLoop(t *testing.T) {
	old := luaVMBudget
	luaVMBudget = 50 * time.Millisecond
	defer func() { luaVMBudget = old }()
	start := time.Now()
	err := executeLuaSource([]byte("while true do end"), map[string]luaValue{}, nil)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected a VM budget error")
	}
	if !luaBudgetExceeded(err) {
		t.Fatalf("err = %v", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("budget took too long: %s", elapsed)
	}
}

func TestExecuteLuaSourceCycleViaVM(t *testing.T) {
	globals := map[string]luaValue{}
	src := []byte("t = {}\nt.t = t\nfunction _() end\n")
	if err := executeLuaSource(src, globals, nil); err != nil {
		t.Fatal(err)
	}
	v, ok := globals["t"]
	if !ok || v.kind != luaTable {
		t.Fatalf("t = %+v ok=%v", v, ok)
	}
}
