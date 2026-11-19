// Command goro-effects-export writes the compiled EffectSpecs, SkillEffects
// and SkillActions tables as JSON for the streaming data path
// (data/contentdata/effectdata/effects.json).
package main

import (
	"fmt"
	"os"

	"github.com/kivutar/goro/db"
)

func main() {
	out := "effects.json"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	specs, skills, actions := db.CompiledEffectTables()
	data, err := db.MarshalEffectTablesJSON(db.EffectTables{
		Effects:      specs,
		SkillEffects: skills,
		SkillActions: actions,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d effects, %d skill effects, %d skill actions)\n",
		out, len(specs), len(skills), len(actions))
}
