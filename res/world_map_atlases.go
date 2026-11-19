package res

// Renewal world-map atlases. Episode 98 is the live client setting.
const WorldMapEpisode = 98

type WorldMapAtlas struct {
	ID     string
	Name   string
	EpFrom float64
	EpTo   float64
	Maps   []WorldMapSection
}

type WorldMapSection struct {
	ID     string
	Name   string
	EpFrom float64
	EpTo   float64
	Top    int
	Left   int
	Width  int
	Height int
}

func WorldMapAtlasVisible(epFrom, epTo, episode float64) bool {
	return episode >= epFrom && episode < epTo
}

func WorldMapAtlasesForEpisode(episode float64) []WorldMapAtlas {
	out := make([]WorldMapAtlas, 0, len(worldMapAtlases))
	for _, atlas := range worldMapAtlases {
		if !WorldMapAtlasVisible(atlas.EpFrom, atlas.EpTo, episode) {
			continue
		}
		kept := atlas
		kept.Maps = nil
		for _, section := range atlas.Maps {
			if WorldMapAtlasVisible(section.EpFrom, section.EpTo, episode) {
				kept.Maps = append(kept.Maps, section)
			}
		}
		out = append(out, kept)
	}
	return out
}

var worldMapAtlases = []WorldMapAtlas{
	{
		ID: "worldmap.jpg", Name: "Midgard", EpFrom: 0.0, EpTo: 99.0,
		Maps: []WorldMapSection{
			{ID: "hu_fild01", Name: "Hugel Field 01 - Entrance to Thanatos Tower", EpFrom: 10.3, EpTo: 99.0, Top: 0, Left: 695, Width: 57, Height: 57},
			{ID: "hu_fild02", Name: "Hugel Field 02", EpFrom: 10.3, EpTo: 99.0, Top: 0, Left: 753, Width: 58, Height: 57},
			{ID: "hu_fild03", Name: "Hugel Field 03 - Royal Hunting Ground", EpFrom: 10.3, EpTo: 13.1, Top: 0, Left: 811, Width: 59, Height: 57},
			{ID: "hu_fild04", Name: "Hugel Field 04", EpFrom: 10.3, EpTo: 99.0, Top: 58, Left: 753, Width: 58, Height: 58},
			{ID: "hu_fild05", Name: "Hugel Field 05 - Entrance to The Abyss Lake", EpFrom: 10.3, EpTo: 99.0, Top: 58, Left: 812, Width: 58, Height: 58},
			{ID: "hu_fild06", Name: "Hugel Field 06", EpFrom: 10.4, EpTo: 99.0, Top: 58, Left: 871, Width: 56, Height: 58},
			{ID: "hu_fild07", Name: "Hugel Field 07", EpFrom: 10.4, EpTo: 13.1, Top: 116, Left: 811, Width: 59, Height: 57},
			{ID: "hugel", Name: "Hugel", EpFrom: 10.4, EpTo: 99.0, Top: 0, Left: 871, Width: 56, Height: 57},
			{ID: "ein_fild01", Name: "Einbroch Field 01", EpFrom: 10.1, EpTo: 99.0, Top: 58, Left: 576, Width: 59, Height: 58},
			{ID: "ein_fild02", Name: "Einbroch Field 02", EpFrom: 10.1, EpTo: 13.1, Top: 116, Left: 575, Width: 60, Height: 57},
			{ID: "ein_fild03", Name: "Einbroch Field 03", EpFrom: 10.1, EpTo: 99.0, Top: 174, Left: 460, Width: 57, Height: 60},
			{ID: "ein_fild04", Name: "Einbroch Field 04", EpFrom: 10.1, EpTo: 99.0, Top: 174, Left: 518, Width: 57, Height: 60},
			{ID: "ein_fild05", Name: "Einbroch Field 05", EpFrom: 10.1, EpTo: 99.0, Top: 174, Left: 576, Width: 59, Height: 60},
			{ID: "ein_fild06", Name: "Einbroch Field 06", EpFrom: 10.1, EpTo: 99.0, Top: 174, Left: 636, Width: 58, Height: 60},
			{ID: "ein_fild07", Name: "Einbroch Field 07", EpFrom: 10.1, EpTo: 99.0, Top: 235, Left: 636, Width: 58, Height: 56},
			{ID: "ein_fild08", Name: "Einbroch Field 08", EpFrom: 10.1, EpTo: 99.0, Top: 292, Left: 518, Width: 57, Height: 59},
			{ID: "ein_fild09", Name: "Einbroch Field 09", EpFrom: 10.1, EpTo: 99.0, Top: 292, Left: 576, Width: 59, Height: 59},
			{ID: "ein_fild10", Name: "Einbroch Field 10", EpFrom: 10.1, EpTo: 13.1, Top: 292, Left: 636, Width: 58, Height: 59},
			{ID: "einbroch", Name: "Einbroch, the city of steel", EpFrom: 10.1, EpTo: 99.0, Top: 235, Left: 518, Width: 57, Height: 56},
			{ID: "einbech", Name: "Einbech, the mining village", EpFrom: 10.1, EpTo: 99.0, Top: 245, Left: 576, Width: 59, Height: 46},
			{ID: "yuno_fild01", Name: "Yuno Field 01 - Border Posts", EpFrom: 5.0, EpTo: 99.0, Top: 235, Left: 812, Width: 58, Height: 56},
			{ID: "yuno_fild03", Name: "Yuno Field 03 - El Mes Plateau", EpFrom: 5.0, EpTo: 99.0, Top: 117, Left: 707, Width: 45, Height: 56},
			{ID: "yuno_fild04", Name: "Yuno Field 04 - El Mes Plateau", EpFrom: 5.0, EpTo: 99.0, Top: 137, Left: 636, Width: 58, Height: 36},
			{ID: "yuno_fild05", Name: "Yuno Field 05 - El Mes Plateau", EpFrom: 5.0, EpTo: 13.1, Top: 57, Left: 635, Width: 59, Height: 59},
			{ID: "yuno_fild06", Name: "Yuno Field 06 - El Mes Plateau", EpFrom: 10.12, EpTo: 99.0, Top: 58, Left: 707, Width: 45, Height: 58},
			{ID: "yuno_fild07", Name: "Yuno Field 07 - El Mes Gorge (Valley of Abyss)", EpFrom: 10.12, EpTo: 99.0, Top: 174, Left: 695, Width: 57, Height: 60},
			{ID: "yuno_fild09", Name: "Yuno Field 09 - Schwaltzwald Guards Camp", EpFrom: 10.12, EpTo: 99.0, Top: 174, Left: 812, Width: 58, Height: 60},
			{ID: "yuno_fild10", Name: "Yuno Field 10", EpFrom: 10.12, EpTo: 13.1, Top: 174, Left: 871, Width: 58, Height: 60},
			{ID: "yuno_fild11", Name: "Yuno Field 11", EpFrom: 10.12, EpTo: 99.0, Top: 235, Left: 695, Width: 57, Height: 56},
			{ID: "yuno_fild12", Name: "Yuno Field 12 - Border Checkpoint", EpFrom: 10.12, EpTo: 99.0, Top: 235, Left: 753, Width: 58, Height: 56},
			{ID: "yuno", Name: "Yuno, the capital of Schwarzwald Republic", EpFrom: 5.0, EpTo: 99.0, Top: 80, Left: 644, Width: 62, Height: 56},
			{ID: "odin_tem01", Name: "Odin Temple F1", EpFrom: 10.4, EpTo: 99.0, Top: 174, Left: 988, Width: 55, Height: 60},
			{ID: "odin_tem02", Name: "Odin Temple F2", EpFrom: 10.4, EpTo: 99.0, Top: 174, Left: 1044, Width: 58, Height: 60},
			{ID: "odin_tem03", Name: "Odin Temple F3", EpFrom: 10.4, EpTo: 99.0, Top: 117, Left: 1044, Width: 58, Height: 56},
			{ID: "lhz_fild01", Name: "Lighthalzen Field 01", EpFrom: 10.2, EpTo: 99.0, Top: 235, Left: 401, Width: 58, Height: 56},
			{ID: "lhz_fild03", Name: "Lighthalzen Field 03", EpFrom: 10.2, EpTo: 99.0, Top: 292, Left: 460, Width: 57, Height: 59},
			{ID: "lighthalzen", Name: "Lighthalzen, the City-State of Prosperity", EpFrom: 10.2, EpTo: 99.0, Top: 292, Left: 401, Width: 58, Height: 54},
			{ID: "ra_fild01", Name: "Rachel Field 01 - Audumra Grass Land", EpFrom: 11.1, EpTo: 99.0, Top: 117, Left: 283, Width: 58, Height: 56},
			{ID: "ra_fild02", Name: "Rachel Field 02 - Oz Gorge", EpFrom: 11.1, EpTo: 13.1, Top: 174, Left: 164, Width: 59, Height: 61},
			{ID: "ra_fild03", Name: "Rachel Field 03 - Ida Plane", EpFrom: 11.1, EpTo: 99.0, Top: 174, Left: 224, Width: 58, Height: 61},
			{ID: "ra_fild04", Name: "Rachel Field 04 - Audumra Grass Land", EpFrom: 11.1, EpTo: 99.0, Top: 174, Left: 283, Width: 58, Height: 60},
			{ID: "ra_fild05", Name: "Rachel Field 05 - Audumra Grass Land", EpFrom: 11.1, EpTo: 99.0, Top: 174, Left: 342, Width: 58, Height: 60},
			{ID: "ra_fild06", Name: "Rachel Field 06 - Fortu Luna", EpFrom: 11.1, EpTo: 99.0, Top: 174, Left: 401, Width: 58, Height: 60},
			{ID: "ra_fild07", Name: "Rachel Field 07 - Oz Gorge", EpFrom: 11.1, EpTo: 13.1, Top: 235, Left: 164, Width: 59, Height: 58},
			{ID: "ra_fild08", Name: "Rachel Field 08 - Ida Plane", EpFrom: 11.1, EpTo: 99.0, Top: 235, Left: 283, Width: 58, Height: 58},
			{ID: "ra_fild09", Name: "Rachel Field 09 - Audumra Grass Land", EpFrom: 11.1, EpTo: 13.1, Top: 235, Left: 342, Width: 58, Height: 56},
			{ID: "ra_fild10", Name: "Rachel Field 10 - Oz Gorge", EpFrom: 11.1, EpTo: 13.1, Top: 292, Left: 107, Width: 56, Height: 59},
			{ID: "ra_fild11", Name: "Rachel Field 11 - Ida Plane", EpFrom: 11.1, EpTo: 13.1, Top: 292, Left: 164, Width: 59, Height: 59},
			{ID: "ra_fild12", Name: "Rachel Field 12 - Ida Plane", EpFrom: 11.1, EpTo: 99.0, Top: 292, Left: 283, Width: 58, Height: 59},
			{ID: "ra_fild13", Name: "Rachel Field 13 - Beach of Tears", EpFrom: 11.1, EpTo: 13.1, Top: 352, Left: 283, Width: 58, Height: 57},
			{ID: "ra_temple", Name: "Rachel Temple (Sesrumnis)", EpFrom: 11.1, EpTo: 99.0, Top: 261, Left: 236, Width: 35, Height: 30},
			{ID: "rachel", Name: "Rachel, Capital of Arunafelz, the Study Nation", EpFrom: 11.1, EpTo: 99.0, Top: 292, Left: 224, Width: 58, Height: 59},
			{ID: "ve_fild01", Name: "Veins Field 01", EpFrom: 11.2, EpTo: 99.0, Top: 352, Left: 164, Width: 59, Height: 57},
			{ID: "ve_fild02", Name: "Veins Field 02", EpFrom: 11.2, EpTo: 99.0, Top: 352, Left: 224, Width: 58, Height: 57},
			{ID: "aru_gld", Name: "Arunafeltz Guild", EpFrom: 12.0, EpTo: 99.0, Top: 410, Left: 224, Width: 58, Height: 58},
			{ID: "ve_fild03", Name: "Veins Field 03", EpFrom: 11.2, EpTo: 99.0, Top: 410, Left: 107, Width: 56, Height: 58},
			{ID: "ve_fild04", Name: "Veins Field 04", EpFrom: 11.2, EpTo: 99.0, Top: 410, Left: 164, Width: 59, Height: 58},
			{ID: "ve_fild05", Name: "Veins Field 05", EpFrom: 11.2, EpTo: 13.1, Top: 470, Left: 106, Width: 58, Height: 58},
			{ID: "ve_fild06", Name: "Veins Field 06", EpFrom: 11.2, EpTo: 99.0, Top: 485, Left: 164, Width: 59, Height: 42},
			{ID: "ve_fild07", Name: "Veins Field 07", EpFrom: 11.2, EpTo: 99.0, Top: 586, Left: 161, Width: 57, Height: 57},
			{ID: "veins", Name: "Veins, the Canyon Village", EpFrom: 11.2, EpTo: 99.0, Top: 528, Left: 161, Width: 57, Height: 57},
			{ID: "aldebaran", Name: "Al De Baran", EpFrom: 0.0, EpTo: 99.0, Top: 292, Left: 812, Width: 58, Height: 59},
			{ID: "alde_gld", Name: "Luina the satellite of Al De Baran", EpFrom: 4.0, EpTo: 99.0, Top: 292, Left: 753, Width: 58, Height: 59},
			{ID: "mjolnir_01", Name: "Mt.Mjolnir 01", EpFrom: 0.0, EpTo: 99.0, Top: 411, Left: 576, Width: 59, Height: 57},
			{ID: "mjolnir_02", Name: "Mt.Mjolnir 02", EpFrom: 0.0, EpTo: 99.0, Top: 411, Left: 636, Width: 58, Height: 57},
			{ID: "mjolnir_03", Name: "Mt.Mjolnir 03", EpFrom: 0.0, EpTo: 99.0, Top: 411, Left: 695, Width: 57, Height: 57},
			{ID: "mjolnir_04", Name: "Mt.Mjolnir 04", EpFrom: 0.0, EpTo: 99.0, Top: 411, Left: 753, Width: 58, Height: 57},
			{ID: "mjolnir_05", Name: "Mt.Mjolnir 05", EpFrom: 0.0, EpTo: 99.0, Top: 411, Left: 812, Width: 58, Height: 57},
			{ID: "mjolnir_06", Name: "Mt.Mjolnir 06", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 636, Width: 58, Height: 58},
			{ID: "mjolnir_07", Name: "Mt.Mjolnir 07", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 695, Width: 57, Height: 58},
			{ID: "mjolnir_08", Name: "Mt.Mjolnir 08", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 753, Width: 58, Height: 58},
			{ID: "mjolnir_09", Name: "Mt.Mjolnir 09", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 753, Width: 58, Height: 58},
			{ID: "mjolnir_10", Name: "Mt.Mjolnir 10", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 812, Width: 58, Height: 58},
			{ID: "mjolnir_11", Name: "Mt.Mjolnir 11", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 871, Width: 56, Height: 58},
			{ID: "mjolnir_12", Name: "Mt.Mjolnir 12", EpFrom: 0.0, EpTo: 99.0, Top: 352, Left: 812, Width: 58, Height: 58},
			{ID: "glast_01", Name: "Entrance to Glast Heim", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 408, Width: 51, Height: 58},
			{ID: "gef_fild00", Name: "Geffen Field 00", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 636, Width: 58, Height: 58},
			{ID: "gef_fild01", Name: "Geffen Field 01", EpFrom: 0.0, EpTo: 99.0, Top: 587, Left: 636, Width: 58, Height: 56},
			{ID: "gef_fild02", Name: "Geffen Field 02", EpFrom: 0.0, EpTo: 99.0, Top: 644, Left: 695, Width: 57, Height: 57},
			{ID: "gef_fild03", Name: "Geffen Field 03", EpFrom: 0.0, EpTo: 99.0, Top: 644, Left: 636, Width: 58, Height: 57},
			{ID: "gef_fild04", Name: "Geffen Field 04", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 576, Width: 59, Height: 58},
			{ID: "gef_fild05", Name: "Geffen Field 05", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 518, Width: 57, Height: 58},
			{ID: "gef_fild06", Name: "Geffen Field 06", EpFrom: 0.0, EpTo: 99.0, Top: 469, Left: 460, Width: 57, Height: 58},
			{ID: "gef_fild07", Name: "Geffen Field 07", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 518, Width: 57, Height: 58},
			{ID: "gef_fild08", Name: "Geffen Field 08", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 460, Width: 57, Height: 58},
			{ID: "gef_fild09", Name: "Geffen Field 09", EpFrom: 0.0, EpTo: 99.0, Top: 587, Left: 576, Width: 59, Height: 56},
			{ID: "gef_fild10", Name: "Geffen Field 10", EpFrom: 0.0, EpTo: 99.0, Top: 644, Left: 576, Width: 59, Height: 57},
			{ID: "gef_fild11", Name: "Geffen Field 11", EpFrom: 0.0, EpTo: 99.0, Top: 702, Left: 576, Width: 59, Height: 58},
			{ID: "gef_fild12", Name: "Geffen Field 12 - Kordt Forest", EpFrom: 0.0, EpTo: 13.1, Top: 587, Left: 460, Width: 58, Height: 56},
			{ID: "gef_fild13", Name: "Geffen Field 13 - Britoniah Guild (Geffen)", EpFrom: 0.0, EpTo: 99.0, Top: 587, Left: 518, Width: 57, Height: 56},
			{ID: "gef_fild14", Name: "Geffen Field 14 - West Orc Village", EpFrom: 0.0, EpTo: 13.1, Top: 644, Left: 518, Width: 57, Height: 57},
			{ID: "geffen", Name: "Geffen", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 576, Width: 59, Height: 58},
			{ID: "prt_fild00", Name: "Prontera Field 00", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 695, Width: 57, Height: 58},
			{ID: "prt_fild01", Name: "Prontera Field 01", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 812, Width: 58, Height: 58},
			{ID: "prt_fild02", Name: "Prontera Field 02", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 871, Width: 56, Height: 58},
			{ID: "prt_fild03", Name: "Prontera Field 03", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 928, Width: 58, Height: 42},
			{ID: "prt_fild04", Name: "Prontera Field 04", EpFrom: 0.0, EpTo: 99.0, Top: 587, Left: 695, Width: 57, Height: 56},
			{ID: "prt_fild05", Name: "Prontera Field 05", EpFrom: 0.0, EpTo: 99.0, Top: 587, Left: 753, Width: 58, Height: 56},
			{ID: "prt_fild06", Name: "Prontera Field 06", EpFrom: 0.0, EpTo: 99.0, Top: 587, Left: 871, Width: 56, Height: 56},
			{ID: "prt_fild07", Name: "Prontera Field 07", EpFrom: 0.0, EpTo: 99.0, Top: 644, Left: 753, Width: 58, Height: 57},
			{ID: "prt_fild08", Name: "Prontera Field 08", EpFrom: 0.0, EpTo: 99.0, Top: 644, Left: 812, Width: 58, Height: 57},
			{ID: "prt_fild09", Name: "Prontera Field 09", EpFrom: 0.0, EpTo: 99.0, Top: 702, Left: 753, Width: 58, Height: 58},
			{ID: "prt_fild10", Name: "Prontera Field 10", EpFrom: 0.0, EpTo: 99.0, Top: 702, Left: 695, Width: 57, Height: 58},
			{ID: "prt_fild11", Name: "Prontera Field 11", EpFrom: 0.0, EpTo: 99.0, Top: 702, Left: 636, Width: 58, Height: 58},
			{ID: "prt_monk", Name: "St. Capitolina Abbey", EpFrom: 0.0, EpTo: 99.0, Top: 528, Left: 987, Width: 55, Height: 42},
			{ID: "prontera", Name: "Prontera", EpFrom: 0.0, EpTo: 99.0, Top: 587, Left: 812, Width: 58, Height: 56},
			{ID: "izlude", Name: "Izlude", EpFrom: 0.0, EpTo: 99.0, Top: 644, Left: 871, Width: 31, Height: 32},
			{ID: "izlu2dun", Name: "Byalan Island", EpFrom: 0.0, EpTo: 99.0, Top: 644, Left: 928, Width: 20, Height: 21},
			{ID: "pay_fild01", Name: "Payon Forest 01", EpFrom: 0.0, EpTo: 99.0, Top: 804, Left: 936, Width: 58, Height: 57},
			{ID: "pay_fild02", Name: "Payon Forest 02", EpFrom: 0.0, EpTo: 99.0, Top: 869, Left: 960, Width: 43, Height: 62},
			{ID: "pay_fild03", Name: "Payon Forest 03", EpFrom: 0.0, EpTo: 99.0, Top: 885, Left: 1004, Width: 54, Height: 46},
			{ID: "pay_fild04", Name: "Payon Forest 04", EpFrom: 0.0, EpTo: 99.0, Top: 702, Left: 871, Width: 56, Height: 58},
			{ID: "pay_fild05", Name: "Payon Forest 05", EpFrom: 0.0, EpTo: 13.1, Top: 932, Left: 960, Width: 43, Height: 57},
			{ID: "pay_fild06", Name: "Payon Forest 06", EpFrom: 0.0, EpTo: 99.0, Top: 932, Left: 1004, Width: 54, Height: 57},
			{ID: "pay_fild07", Name: "Payon Forest 07", EpFrom: 0.0, EpTo: 99.0, Top: 804, Left: 995, Width: 58, Height: 57},
			{ID: "pay_fild08", Name: "Payon Forest 08", EpFrom: 0.0, EpTo: 99.0, Top: 746, Left: 1014, Width: 39, Height: 57},
			{ID: "pay_fild09", Name: "Payon Forest 09", EpFrom: 0.0, EpTo: 99.0, Top: 746, Left: 1054, Width: 59, Height: 57},
			{ID: "pay_fild10", Name: "Payon Forest 10", EpFrom: 0.0, EpTo: 99.0, Top: 804, Left: 1054, Width: 59, Height: 57},
			{ID: "pay_fild11", Name: "Payon Forest 11", EpFrom: 0.0, EpTo: 13.1, Top: 869, Left: 907, Width: 53, Height: 62},
			{ID: "pay_gld", Name: "Payon Forest - Greenwood Lake (Guild)", EpFrom: 4.0, EpTo: 99.0, Top: 764, Left: 910, Width: 58, Height: 58},
			{ID: "pay_arche", Name: "Payon Forest - Archer Village", EpFrom: 0.0, EpTo: 99.0, Top: 694, Left: 985, Width: 45, Height: 51},
			{ID: "payon", Name: "Payon Town, home of Archers", EpFrom: 0.0, EpTo: 99.0, Top: 746, Left: 967, Width: 46, Height: 57},
			{ID: "alberta", Name: "Alberta, City of Merchants", EpFrom: 0.0, EpTo: 99.0, Top: 913, Left: 1059, Width: 54, Height: 51},
			{ID: "alb2trea", Name: "Alberta Island", EpFrom: 0.0, EpTo: 99.0, Top: 920, Left: 1114, Width: 20, Height: 24},
			{ID: "tur_dun01", Name: "Turtle Island", EpFrom: 0.0, EpTo: 99.0, Top: 974, Left: 1163, Width: 29, Height: 30},
			{ID: "moc_fild01", Name: "Morroc Field 01 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 702, Left: 812, Width: 58, Height: 58},
			{ID: "moc_fild02", Name: "Morroc Field 02 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 761, Left: 858, Width: 48, Height: 48},
			{ID: "moc_fild03", Name: "Morroc Field 03 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 810, Left: 883, Width: 52, Height: 58},
			{ID: "moc_fild04", Name: "Morroc Field 04 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 785, Left: 777, Width: 57, Height: 59},
			{ID: "moc_fild05", Name: "Morroc Field 05 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 785, Left: 718, Width: 58, Height: 59},
			{ID: "moc_fild06", Name: "Morroc Field 06 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 785, Left: 660, Width: 57, Height: 59},
			{ID: "moc_fild07", Name: "Morroc Field 07 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 785, Left: 603, Width: 56, Height: 59},
			{ID: "moc_fild08", Name: "Morroc Field 08 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 845, Left: 777, Width: 57, Height: 48},
			{ID: "moc_fild09", Name: "Morroc Field 09 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 845, Left: 718, Width: 58, Height: 48},
			{ID: "moc_fild10", Name: "Morroc Field 10 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 845, Left: 657, Width: 60, Height: 48},
			{ID: "moc_fild11", Name: "Morroc Field 11 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 894, Left: 660, Width: 57, Height: 56},
			{ID: "moc_fild12", Name: "Morroc Field 12 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 894, Left: 603, Width: 56, Height: 56},
			{ID: "moc_fild13", Name: "Morroc Field 13 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 810, Left: 836, Width: 46, Height: 58},
			{ID: "moc_fild14", Name: "Morroc Field 14 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 894, Left: 777, Width: 57, Height: 55},
			{ID: "moc_fild15", Name: "Morroc Field 15 - Sograt Desert", EpFrom: 0.0, EpTo: 12.0, Top: 894, Left: 718, Width: 58, Height: 44},
			{ID: "moc_fild16", Name: "Morroc Field 16 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 939, Left: 718, Width: 58, Height: 57},
			{ID: "moc_fild17", Name: "Morroc Field 17 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 951, Left: 660, Width: 57, Height: 58},
			{ID: "moc_fild18", Name: "Morroc Field 18 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 951, Left: 603, Width: 56, Height: 58},
			{ID: "moc_fild19", Name: "Morroc Field 19 - Sograt Desert", EpFrom: 0.0, EpTo: 99.0, Top: 849, Left: 567, Width: 35, Height: 32},
			{ID: "moc_fild20", Name: "Morroc Field 20 - Sograt Desert - Continental Guard Quarantine", EpFrom: 12.0, EpTo: 99.0, Top: 838, Left: 716, Width: 58, Height: 58},
			{ID: "moc_ruins", Name: "Morroc Ruins", EpFrom: 0.0, EpTo: 99.0, Top: 817, Left: 567, Width: 35, Height: 31},
			{ID: "morocc", Name: "Morroc Town", EpFrom: 0.0, EpTo: 99.0, Top: 845, Left: 606, Width: 49, Height: 48},
			{ID: "cmd_fild01", Name: "Comodo Field 01 - Papuchicha Forest", EpFrom: 3.0, EpTo: 99.0, Top: 835, Left: 367, Width: 58, Height: 58},
			{ID: "cmd_fild02", Name: "Comodo Field 02 - Kokomo beach", EpFrom: 3.0, EpTo: 99.0, Top: 894, Left: 367, Width: 58, Height: 56},
			{ID: "cmd_fild03", Name: "Comodo Field 03 - Zenhai Marsh", EpFrom: 3.0, EpTo: 99.0, Top: 835, Left: 426, Width: 57, Height: 58},
			{ID: "cmd_fild04", Name: "Comodo Field 04 - Kokomo beach", EpFrom: 3.0, EpTo: 99.0, Top: 894, Left: 426, Width: 57, Height: 56},
			{ID: "cmd_fild05", Name: "Comodo Field 05 - Border of Papuchica forest", EpFrom: 3.0, EpTo: 13.1, Top: 835, Left: 484, Width: 59, Height: 58},
			{ID: "cmd_fild06", Name: "Comodo Field 06 - Fortress Saint Darmain (West)", EpFrom: 3.0, EpTo: 99.0, Top: 894, Left: 484, Width: 59, Height: 56},
			{ID: "cmd_fild07", Name: "Comodo Field 07 - Beacon Island, Pharos", EpFrom: 3.0, EpTo: 99.0, Top: 951, Left: 484, Width: 59, Height: 58},
			{ID: "cmd_fild08", Name: "Comodo Field 08 - Fortress Saint Darmain (East)", EpFrom: 3.0, EpTo: 99.0, Top: 894, Left: 544, Width: 58, Height: 56},
			{ID: "cmd_fild09", Name: "Comodo Field 09 - Fortress Saint Darmain (South)", EpFrom: 3.0, EpTo: 99.0, Top: 951, Left: 544, Width: 58, Height: 58},
			{ID: "comodo", Name: "Comodo Town", EpFrom: 3.0, EpTo: 99.0, Top: 835, Left: 297, Width: 59, Height: 58},
			{ID: "um_fild01", Name: "Umbala Forest 01 - Luluka Forest", EpFrom: 0.0, EpTo: 99.0, Top: 776, Left: 239, Width: 57, Height: 58},
			{ID: "um_fild02", Name: "Umbala Forest 02 - Hoomga Forest", EpFrom: 0.0, EpTo: 99.0, Top: 776, Left: 297, Width: 59, Height: 58},
			{ID: "um_fild03", Name: "Umbala Forest 03 - Kalala Swamp", EpFrom: 0.0, EpTo: 99.0, Top: 776, Left: 357, Width: 59, Height: 58},
			{ID: "um_fild04", Name: "Umbala Forest 04 - Hoomga Jungle", EpFrom: 0.0, EpTo: 99.0, Top: 717, Left: 297, Width: 59, Height: 58},
			{ID: "umbala", Name: "Umbala Town", EpFrom: 0.0, EpTo: 99.0, Top: 660, Left: 308, Width: 48, Height: 56},
			{ID: "nameless_n", Name: "Nameless Island", EpFrom: 11.3, EpTo: 99.0, Top: 698, Left: 94, Width: 63, Height: 62},
		},
	},
	{
		ID: "worldmap_localizing1.bmp", Name: "Eastern Kingdoms", EpFrom: 6.0, EpTo: 99.0,
		Maps: []WorldMapSection{
			{ID: "ama_fild01", Name: "Amatsu Field 01", EpFrom: 6.0, EpTo: 99.0, Top: 165, Left: 332, Width: 88, Height: 88},
			{ID: "amatsu", Name: "Amatsu", EpFrom: 6.0, EpTo: 99.0, Top: 256, Left: 281, Width: 88, Height: 88},
			{ID: "gon_fild01", Name: "Gonryun Field 01", EpFrom: 6.0, EpTo: 99.0, Top: 268, Left: 911, Width: 88, Height: 88},
			{ID: "gon_dun01", Name: "Gonryun Dungeon 01", EpFrom: 6.0, EpTo: 99.0, Top: 100, Left: 1030, Width: 88, Height: 88},
			{ID: "gonryun", Name: "Gonryun", EpFrom: 6.0, EpTo: 99.0, Top: 170, Left: 906, Width: 88, Height: 88},
			{ID: "lou_fild01", Name: "Louyang Field 01", EpFrom: 8.12, EpTo: 99.0, Top: 768, Left: 276, Width: 88, Height: 88},
			{ID: "lou_dun01", Name: "Louyang Dungeon 01", EpFrom: 8.12, EpTo: 99.0, Top: 672, Left: 178, Width: 88, Height: 88},
			{ID: "louyang", Name: "Louyang", EpFrom: 8.12, EpTo: 99.0, Top: 673, Left: 276, Width: 88, Height: 88},
			{ID: "ayo_fild01", Name: "Ayothaya Field 01", EpFrom: 8.2, EpTo: 99.0, Top: 734, Left: 920, Width: 88, Height: 88},
			{ID: "ayothaya", Name: "Ayothaya", EpFrom: 8.2, EpTo: 99.0, Top: 724, Left: 827, Width: 88, Height: 88},
		},
	},
	{
		ID: "worldmap_localizing2.bmp", Name: "Far Lands", EpFrom: 12.0, EpTo: 99.0,
		Maps: []WorldMapSection{
			{ID: "mosk_fild02", Name: "Moscovia Field 02", EpFrom: 12.0, EpTo: 99.0, Top: 299, Left: 187, Width: 88, Height: 88},
			{ID: "mosk_dun01", Name: "Moscovia Dungeon 01", EpFrom: 12.0, EpTo: 99.0, Top: 206, Left: 185, Width: 88, Height: 88},
			{ID: "mosk_dun02", Name: "Moscovia Dungeon 02", EpFrom: 12.0, EpTo: 99.0, Top: 111, Left: 207, Width: 88, Height: 88},
			{ID: "mosk_dun03", Name: "Moscovia Dungeon 03", EpFrom: 12.0, EpTo: 99.0, Top: 126, Left: 307, Width: 88, Height: 88},
			{ID: "moscovia", Name: "Moscovia", EpFrom: 12.0, EpTo: 99.0, Top: 332, Left: 394, Width: 110, Height: 110},
			{ID: "bra_fild01", Name: "Brasilis Field 01", EpFrom: 13.2, EpTo: 99.0, Top: 110, Left: 998, Width: 88, Height: 88},
			{ID: "brasilis", Name: "Brasilis", EpFrom: 13.2, EpTo: 99.0, Top: 255, Left: 924, Width: 88, Height: 88},
			{ID: "dew_fild01", Name: "Dewata Field (Tribal Village)", EpFrom: 14.1, EpTo: 99.0, Top: 668, Left: 37, Width: 88, Height: 88},
			{ID: "dew_dun01", Name: "Dewata Dungeon 01", EpFrom: 14.1, EpTo: 99.0, Top: 836, Left: 337, Width: 88, Height: 88},
			{ID: "dewata", Name: "Dewata", EpFrom: 14.1, EpTo: 99.0, Top: 712, Left: 232, Width: 88, Height: 88},
			{ID: "ma_fild01", Name: "Malaya Field 01 - Baryo Mahiwaga", EpFrom: 14.1, EpTo: 99.0, Top: 810, Left: 956, Width: 88, Height: 88},
			{ID: "ma_fild02", Name: "Malaya Field 02 - Forest", EpFrom: 14.1, EpTo: 99.0, Top: 713, Left: 946, Width: 88, Height: 88},
			{ID: "ma_scene01", Name: "Bakonawa Lake", EpFrom: 14.1, EpTo: 99.0, Top: 601, Left: 1076, Width: 88, Height: 88},
			{ID: "malaya", Name: "Malaya", EpFrom: 14.1, EpTo: 99.0, Top: 883, Left: 832, Width: 88, Height: 88},
		},
	},
	{
		ID: "worldmap_dimension.bmp", Name: "New World", EpFrom: 13.1, EpTo: 99.0,
		Maps: []WorldMapSection{
			{ID: "eclage", Name: "Eclage", EpFrom: 14.2, EpTo: 99.0, Top: 185, Left: 393, Width: 91, Height: 91},
			{ID: "ecl_fild01", Name: "Blooming Flower Land", EpFrom: 14.2, EpTo: 99.0, Top: 275, Left: 393, Width: 91, Height: 91},
			{ID: "mora", Name: "Mora Village", EpFrom: 14.1, EpTo: 99.0, Top: 365, Left: 393, Width: 91, Height: 91},
			{ID: "bif_fild01", Name: "Bifrost Bridge", EpFrom: 14.1, EpTo: 99.0, Top: 455, Left: 393, Width: 91, Height: 91},
			{ID: "splendide", Name: "Forest Village Splendide", EpFrom: 13.2, EpTo: 99.0, Top: 544, Left: 393, Width: 91, Height: 91},
			{ID: "spl_fild01", Name: "Splendide Field 01", EpFrom: 13.1, EpTo: 99.0, Top: 455, Left: 483, Width: 91, Height: 91},
			{ID: "spl_fild02", Name: "Splendide Field 02", EpFrom: 13.1, EpTo: 99.0, Top: 544, Left: 483, Width: 91, Height: 91},
			{ID: "spl_fild02", Name: "Splendide Field 03", EpFrom: 13.2, EpTo: 99.0, Top: 636, Left: 483, Width: 91, Height: 91},
			{ID: "mid_camp", Name: "Rune Midgard Allied Forces Post", EpFrom: 13.1, EpTo: 99.0, Top: 544, Left: 573, Width: 91, Height: 91},
			{ID: "man_fild01", Name: "Manuk Field 01", EpFrom: 13.1, EpTo: 99.0, Top: 544, Left: 662, Width: 91, Height: 91},
			{ID: "man_fild03", Name: "Manuk Field 03", EpFrom: 13.2, EpTo: 99.0, Top: 636, Left: 662, Width: 91, Height: 91},
			{ID: "man_fild02", Name: "Manuk Field 02", EpFrom: 13.1, EpTo: 99.0, Top: 544, Left: 753, Width: 91, Height: 91},
			{ID: "manuk", Name: "Mining Village Manuk", EpFrom: 13.2, EpTo: 99.0, Top: 636, Left: 753, Width: 91, Height: 91},
			{ID: "dicastes02", Name: "Dicastes Diel", EpFrom: 13.3, EpTo: 99.0, Top: 451, Left: 978, Width: 91, Height: 91},
			{ID: "dicastes01", Name: "El Dicastes, the Sapha Capital", EpFrom: 13.3, EpTo: 99.0, Top: 544, Left: 978, Width: 91, Height: 91},
			{ID: "dic_fild01", Name: "Outskirts of Kamidal Mountain 01", EpFrom: 13.3, EpTo: 99.0, Top: 636, Left: 978, Width: 91, Height: 91},
			{ID: "dic_fild02", Name: "Outskirts of Kamidal Mountain 02", EpFrom: 13.3, EpTo: 99.0, Top: 729, Left: 978, Width: 91, Height: 91},
		},
	},
	{
		ID: "WorldMap_Isgard.jpg", Name: "Isgard", EpFrom: 0.0, EpTo: 99.0,
		Maps: []WorldMapSection{
			{ID: "jor_tail", Name: "Frozen Tail", EpFrom: 0.0, EpTo: 99.0, Top: 906, Left: 881, Width: 116, Height: 112},
			{ID: "jor_back1", Name: "Frozen Scale Hill", EpFrom: 0.0, EpTo: 99.0, Top: 789, Left: 881, Width: 116, Height: 116},
			{ID: "jor_back2", Name: "Frozen Scale Plains", EpFrom: 0.0, EpTo: 99.0, Top: 672, Left: 881, Width: 116, Height: 116},
			{ID: "jor_back3", Name: "Frozen Scale Glacier", EpFrom: 0.0, EpTo: 99.0, Top: 672, Left: 764, Width: 116, Height: 116},
			{ID: "jor_back4", Name: "Frozen Scale Beach", EpFrom: 0.0, EpTo: 99.0, Top: 556, Left: 764, Width: 116, Height: 114},
			{ID: "jor_back5", Name: "Ancient Ice Canyon East", EpFrom: 0.0, EpTo: 99.0, Top: 438, Left: 647, Width: 116, Height: 116},
			{ID: "jor_back6", Name: "Ancient Ice Canyon West", EpFrom: 0.0, EpTo: 99.0, Top: 438, Left: 530, Width: 116, Height: 116},
			{ID: "jor_twice", Name: "Drift Ice Zone", EpFrom: 0.0, EpTo: 99.0, Top: 204, Left: 530, Width: 116, Height: 115},
		},
	},
}
