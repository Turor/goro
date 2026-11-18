package res

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kivutar/goro/db"
	"github.com/yuin/gopher-lua/parse"
)

func TestEvalLuaDataChunkResolvesReferencesAndSkipsNilKeys(t *testing.T) {
	src := `
-- comment
SKILL_INFO_LIST = {
	[SKID.AL_HEAL] = { "AL_HEAL", SkillName = "Heal", MaxLv = 10, SpAmount = { 13, 16 }, Neg = -2, Hex = 0x10, Flag = true },
	[SKID.NOT_IN_THIS_CLIENT] = { "GONE", SkillName = "Gone" },
}
OTHER = SKILL_INFO_LIST
`
	chunk, err := parse.Parse(strings.NewReader(src), "t")
	if err != nil {
		t.Fatal(err)
	}
	globals := map[string]luaValue{"SKID": {kind: luaTable, table: map[interface{}]luaValue{
		"AL_HEAL": {kind: luaNumber, num: 28},
	}}}
	if err := evalLuaDataChunk(chunk, globals); err != nil {
		t.Fatal(err)
	}
	list := globals["SKILL_INFO_LIST"]
	if len(list.table) != 1 {
		t.Fatalf("entries = %d, want 1 (nil key skipped)", len(list.table))
	}
	heal := list.table[28]
	if luaField(heal, "SkillName").str != "Heal" || luaField(heal, "MaxLv").num != 10 || heal.table[1].str != "AL_HEAL" {
		t.Fatalf("heal = %+v", heal)
	}
	if luaField(heal, "Neg").num != -2 || luaField(heal, "Hex").num != 16 || luaField(heal, "Flag").kind != luaBool {
		t.Fatalf("scalar fields = %+v", heal.table)
	}
	if got := luaStringList(luaField(heal, "SpAmount")); len(got) != 0 {
		t.Fatalf("numeric list read as strings: %v", got)
	}
	if globals["OTHER"].kind != luaTable {
		t.Fatal("global reference not resolved")
	}
}

func TestEvalLuaDataChunkRejectsCode(t *testing.T) {
	for _, src := range []string{"f()", "local x = 1", "function f() end", "X = 1 + 2", "X = f()"} {
		chunk, err := parse.Parse(strings.NewReader(src), "t")
		if err != nil {
			t.Fatal(err)
		}
		if err := evalLuaDataChunk(chunk, map[string]luaValue{}); err != errLuaNotData {
			t.Fatalf("%q: err = %v, want errLuaNotData", src, err)
		}
	}
}

func TestExecuteLuaChunkRunsSourceThatNeedsTheVM(t *testing.T) {
	globals := map[string]luaValue{}
	src := []byte("\xef\xbb\xbfT = {}\nfor i = 1, 3 do T[i] = i * 2 end\nNAME = \"\xbb\xe7\xb0\xfa\"")
	if err := executeLuaChunk(src, globals); err != nil {
		t.Fatal(err)
	}
	if globals["T"].table[3].num != 6 {
		t.Fatalf("T = %+v", globals["T"])
	}
	if globals["NAME"].str != "\xbb\xe7\xb0\xfa" {
		t.Fatalf("CP949 bytes changed: %q", globals["NAME"].str)
	}
}

func TestItemInfoTextTblWithoutMain(t *testing.T) {
	manager, err := NewManagerFS(fstest.MapFS{
		"System/itemInfo.lua": {Data: []byte(`
tbl = {
	[512] = {
		unidentifiedDisplayName = "Apple",
		unidentifiedResourceName = "` + "\xbb\xe7\xb0\xfa" + `",
		unidentifiedDescriptionName = { "" },
		identifiedDisplayName = "Apple",
		identifiedResourceName = "` + "\xbb\xe7\xb0\xfa" + `",
		identifiedDescriptionName = { "A round, edible fruit.", "Weight: ^7777771^000000" },
		slotCount = 0,
		ClassNum = 0
	},
	[490000] = {
		identifiedDisplayName = "Renewal Ring",
		identifiedResourceName = "Ring",
		identifiedDescriptionName = { "A 32-bit item ID." },
		slotCount = 1,
		ClassNum = 0
	},
}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if name, ok := manager.ItemDisplayName(512, true); !ok || name != "Apple" {
		t.Fatalf("512 name = %q %v", name, ok)
	}
	if res, _ := manager.ItemResourceName(512, true); res != "사과" {
		t.Fatalf("512 resource = %q, want CP949 decoded to 사과", res)
	}
	if desc, _ := manager.ItemDescription(512, true); len(desc) != 2 {
		t.Fatalf("512 desc = %q", desc)
	}
	if desc := manager.itemMetadata[512].UnidentifiedDescription; desc != nil {
		t.Fatalf("an empty { \"\" } description should be absent, got %q", desc)
	}
	if slots, _ := manager.ItemSlotCount(490000); slots != 1 {
		t.Fatalf("490000 slots = %d", slots)
	}
}

func TestRenewalSkillTablesFromText(t *testing.T) {
	lua := "data/luafiles514/lua files/"
	manager, err := NewManagerFS(fstest.MapFS{
		lua + "datainfo/jobidentity.lub": {Data: []byte(`JTtbl = { JT_NOVICE = 0, JT_ALCHEMIST = 18 }`)},
		lua + "skillinfoz/skillid.lub":   {Data: []byte(`SKID = { AL_HEAL = 28, AM_PHARMACY = 228 }`)},
		lua + "skillinfoz/skillinfolist.lub": {Data: []byte(`SKILL_INFO_LIST = {
	[SKID.AL_HEAL] = { "AL_HEAL", SkillName = "Heal", MaxLv = 10,
		NeedSkillList = { [JOBID.JT_ALCHEMIST] = { { SKID.AM_PHARMACY, 1 } } } },
}`)},
		lua + "skillinfoz/skilldescript.lub": {Data: []byte(`SKILL_DESCRIPT = {
	[SKID.AM_PHARMACY] = { "Prepare Potion", "Max Level: 10" },
}`)},
		lua + "skillinfoz/skilltreeview.lub": {Data: []byte(`SKILL_TREEVIEW_FOR_JOB = {
	[JOBID.JT_ALCHEMIST] = { [7] = SKID.AM_PHARMACY },
}
JobSkillTab.ChangeSkillTabName(JOBID.JT_NOVICE, "1st", "2nd", "3rd", "4th")`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if name, ok := manager.SkillDisplayName(28); !ok || name != "Heal" {
		t.Fatalf("skill 28 name = %q %v", name, ok)
	}
	if level, ok := manager.SkillMaxLevel(28); !ok || level != 10 {
		t.Fatalf("skill 28 max level = %d %v", level, ok)
	}
	if name, _ := manager.SkillDisplayName(228); name != "Prepare Potion" {
		t.Fatalf("skill 228 name from SKILL_DESCRIPT = %q", name)
	}
	positions, ok := manager.SkillTreePositions(db.JobAlchemist)
	if !ok || positions[228] != 7 {
		t.Fatalf("alchemist positions = %v %v", positions, ok)
	}
}

// A 30k-entry itemInfo-shaped table must take the data path (no VM compile).
func TestLargeItemInfoTableUsesDataPath(t *testing.T) {
	var b strings.Builder
	b.WriteString("tbl = {\n")
	for id := 1; id <= 30000; id++ {
		fmt.Fprintf(&b, "\t[%d] = { identifiedDisplayName = \"Item %d\", identifiedDescriptionName = { \"line\" }, slotCount = 0 },\n", id, id)
	}
	b.WriteString("}\n")
	chunk, err := parse.Parse(strings.NewReader(b.String()), "t")
	if err != nil {
		t.Fatal(err)
	}
	globals := map[string]luaValue{}
	if err := evalLuaDataChunk(chunk, globals); err != nil {
		t.Fatal(err)
	}
	if len(globals["tbl"].table) != 30000 {
		t.Fatalf("entries = %d", len(globals["tbl"].table))
	}
}

// TestTuroranTablesRealWhenConfigured needs a data dir assembled with
// scripts/data/assemble-turoran-data.sh (English itemInfo + Turoran skill tables).
func TestTuroranTablesRealWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	if _, ok := manager.Find("System/itemInfo.lua"); !ok {
		t.Skip("GORO_DATA_DIR is not an assembled Turoran data dir")
	}
	for id, want := range map[int]string{512: "Apple", 501: "Red Potion"} {
		if name, ok := manager.ItemDisplayName(id, true); !ok || name != want {
			t.Fatalf("item %d = %q, want %q", id, name, want)
		}
	}
	// Renewal item IDs above 65535 resolve too.
	if name, ok := manager.ItemDisplayName(400000, true); !ok || name == "" {
		t.Fatalf("item 400000 missing")
	}
	if name, ok := manager.SkillDisplayName(28); !ok || name != "Heal" {
		t.Fatalf("skill 28 = %q", name)
	}
	if level, _ := manager.SkillMaxLevel(28); level != 10 {
		t.Fatalf("skill 28 max level = %d", level)
	}
	for _, job := range []int{db.JobAlchemist, 4054, 4252} {
		if positions, ok := manager.SkillTreePositions(job); !ok || len(positions) == 0 {
			t.Fatalf("job %d skill tree missing", job)
		}
	}
}
