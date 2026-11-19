package res

import (
	"testing"
	"testing/fstest"
)

func TestRandomOptionNamesFromLuaTable(t *testing.T) {
	files := fstest.MapFS{
		"data/luafiles514/lua files/datainfo/enumvar.lub": &fstest.MapFile{Data: []byte("EnumVAR = { VAR_MAXHPAMOUNT = 1 }\n")},
		"data/luafiles514/lua files/datainfo/addrandomoptionnametable.lub": &fstest.MapFile{Data: []byte(`
NameTable_VAR = {}
NameTable_VAR[EnumVAR.VAR_MAXHPAMOUNT] = "MaxHP + %d"
NameTable_VAR[2] = "ATK + %d%%"
`)},
	}
	manager, err := NewManagerFS(files)
	if err != nil {
		t.Fatal(err)
	}
	name, ok := manager.RandomOptionName(1)
	if !ok || name != "MaxHP + %d" {
		t.Fatalf("option 1 = %q ok=%t", name, ok)
	}
	name, ok = manager.RandomOptionName(2)
	if !ok || name != "ATK + %d%%" {
		t.Fatalf("option 2 = %q ok=%t", name, ok)
	}
	if _, ok := manager.RandomOptionName(0x7FFE); ok {
		t.Fatal("forge metadata was treated as a random option")
	}
}
