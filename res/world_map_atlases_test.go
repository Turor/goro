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
