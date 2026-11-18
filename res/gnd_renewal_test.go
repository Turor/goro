package res

import (
	"bytes"
	"strings"
	"testing"
)

// renewalGND builds a 1x1 GND with one raw lightmap, one surface and a water
// block, laid out the way kRO's 1.8/1.9 maps are.
func renewalGND(minor byte, splitRecord int) []byte {
	var buf bytes.Buffer
	buf.WriteString("GRGN")
	buf.WriteByte(1)
	buf.WriteByte(minor)
	writeI32(&buf, 1)
	writeI32(&buf, 1)
	writeF32(&buf, 10)

	writeI32(&buf, 1)
	writeI32(&buf, 80)
	buf.WriteString("grass.bmp")
	buf.Write(make([]byte, 80-len("grass.bmp")))

	writeI32(&buf, 1) // lightmap count
	writeI32(&buf, 8)
	writeI32(&buf, 8)
	writeI32(&buf, 1)
	lightmap := make([]byte, 256)
	lightmap[0] = 200                   // alpha of pixel (0,0)
	lightmap[64], lightmap[65] = 11, 22 // rgb of pixel (0,0)
	lightmap[66] = 33
	buf.Write(lightmap)

	writeI32(&buf, 1) // surfaces
	for i := 0; i < 8; i++ {
		writeF32(&buf, 0)
	}
	writeU16(&buf, 0)
	writeU16(&buf, 0)
	buf.Write([]byte{1, 2, 3, 255})

	for i := 0; i < 4; i++ {
		writeF32(&buf, 5)
	}
	writeI32(&buf, 0)
	writeI32(&buf, -1)
	writeI32(&buf, -1)

	writeF32(&buf, 150) // water level
	writeI32(&buf, 2)   // type
	writeF32(&buf, 1)   // wave height
	writeF32(&buf, 2)   // wave speed
	writeF32(&buf, 50)  // wave pitch
	writeI32(&buf, 3)   // anim speed
	writeI32(&buf, 1)   // split width
	writeI32(&buf, 1)   // split height
	buf.Write(make([]byte, splitRecord))
	return buf.Bytes()
}

func TestParseGNDRenewalRawLightmapsAndWater(t *testing.T) {
	for _, tc := range []struct {
		minor  byte
		record int
	}{{8, 4}, {9, 24}} {
		gnd, err := ParseGND(renewalGND(tc.minor, tc.record))
		if err != nil {
			t.Fatalf("1.%d: %v", tc.minor, err)
		}
		lm, ok := gnd.Lightmap(0)
		if !ok || len(gnd.Lightmaps) != 1 {
			t.Fatalf("1.%d: lightmaps=%d", tc.minor, len(gnd.Lightmaps))
		}
		if lm.Alpha[0][0] != 200 || lm.Color[0][0].R != 11 || lm.Color[0][0].G != 22 || lm.Color[0][0].B != 33 {
			t.Fatalf("1.%d: raw lightmap misread: alpha=%d color=%v", tc.minor, lm.Alpha[0][0], lm.Color[0][0])
		}
		if len(gnd.Surfaces) != 1 || gnd.Surfaces[0].Color.B != 1 || len(gnd.Cells) != 1 {
			t.Fatalf("1.%d: surfaces=%v cells=%d", tc.minor, gnd.Surfaces, len(gnd.Cells))
		}
		w := gnd.Water
		if !w.Present || w.Level != -30 || w.Type != 2 || w.AnimSpeed != 3 || w.SplitWidth != 1 || w.SplitHeight != 1 {
			t.Fatalf("1.%d: water %+v", tc.minor, w)
		}
	}
}

func TestParseGNDRejectsImpossibleCountsWithoutAllocating(t *testing.T) {
	data := renewalGND(9, 24)
	// Lightmap count sits right after the 80-byte texture name.
	offset := 4 + 2 + 12 + 8 + 80
	copy(data[offset:], []byte{0xff, 0xff, 0xff, 0x7f})
	_, err := ParseGND(data)
	if err == nil || !strings.Contains(err.Error(), "exceeds remaining") {
		t.Fatalf("expected a bounded count error, got %v", err)
	}
}

func TestGNDRenewalRealArchiveWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	found := 0
	for _, name := range []string{"data\\1@ch1a.gnd", "data\\bl_grass.gnd", "data\\1@exds.gnd", "data\\prontera.gnd"} {
		data, err := manager.ReadFile(name)
		if err != nil {
			continue
		}
		found++
		gnd, err := ParseGND(data)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		if len(gnd.Cells) != gnd.Width*gnd.Height || len(gnd.Lightmaps) == 0 || len(gnd.Surfaces) == 0 {
			t.Fatalf("%s: %dx%d cells=%d lightmaps=%d surfaces=%d", name, gnd.Width, gnd.Height, len(gnd.Cells), len(gnd.Lightmaps), len(gnd.Surfaces))
		}
		if gnd.VersionMinor >= 8 && (!gnd.Water.Present || gnd.Water.SplitWidth <= 0) {
			t.Fatalf("%s v1.%d: missing water block %+v", name, gnd.VersionMinor, gnd.Water)
		}
	}
	if found == 0 {
		t.Skip("no renewal GND samples under GORO_DATA_DIR")
	}
}

// TestDataCheckRealMapFormatsWhenConfigured walks every map, model, sprite and
// palette in GORO_DATA_DIR and fails on any parser rejection.
func TestDataCheckRealMapFormatsWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	report := manager.CheckArchives(DataCheckOptions{Filter: func(name string) bool {
		lower := strings.ToLower(name)
		for _, ext := range []string{".rsw", ".gnd", ".gat", ".rsm", ".rsm2", ".act", ".spr", ".pal", ".imf", ".str"} {
			if strings.HasSuffix(lower, ext) {
				return true
			}
		}
		return false
	}})
	if report.Files == 0 {
		t.Skip("GORO_DATA_DIR has no map or sprite files")
	}
	for i, failure := range report.Failures {
		if i >= 20 {
			break
		}
		t.Errorf("%s (%s): %s", failure.Name, failure.Kind, failure.Err)
	}
	if len(report.Failures) > 0 {
		t.Fatalf("%d of %d files failed", len(report.Failures), report.Files)
	}
}
