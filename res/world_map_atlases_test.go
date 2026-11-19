package res

import "testing"

func TestWorldMapAtlasesEpisode98(t *testing.T) {
	atlases := WorldMapAtlasesForEpisode(WorldMapEpisode)
	if len(atlases) != 5 {
		t.Fatalf("atlases = %d", len(atlases))
	}
	want := []string{"Midgard", "Eastern Kingdoms", "Far Lands", "New World", "Isgard"}
	for i, name := range want {
		if atlases[i].Name != name || len(atlases[i].Maps) == 0 {
			t.Fatalf("atlas %d = %q maps %d", i, atlases[i].Name, len(atlases[i].Maps))
		}
	}
	for _, section := range atlases[0].Maps {
		if section.EpTo <= WorldMapEpisode {
			t.Fatalf("episode filter kept %s ep_to %v", section.ID, section.EpTo)
		}
	}
}

func TestWorldMapAtlasTilesDoNotShareARegion(t *testing.T) {
	atlases := WorldMapAtlasesForEpisode(WorldMapEpisode)
	entries := atlases[0].Entries()
	if len(entries) < 2 || entries[0].Region == entries[1].Region {
		t.Fatal("continent tiles share a warp region")
	}
	var field3 string
	for _, atlas := range atlases {
		for _, section := range atlas.Maps {
			if section.Name == "Splendide Field 03" {
				field3 = section.ID
			}
		}
	}
	if field3 != "spl_fild03" {
		t.Fatalf("splendide field 03 id = %q", field3)
	}
}
