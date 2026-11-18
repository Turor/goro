package rotheme

import (
	"testing"

	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

func TestFieldWidgetsFollowLayoutTokens(t *testing.T) {
	t.Cleanup(ResetLayout)
	ctx := widget.NewContext()
	loose := geometry.Loose(geometry.Sz(800, 800))

	defaultRow := TableRowHeight
	defaultAreaW := textAreaWidth
	defaultAreaH := textAreaHeight
	defaultFill := selectListRowFill
	defaultViewW := tableViewFallbackW
	defaultViewRow := tableViewRowHeight
	defaultHeader := Default.Typography.TextSize + TableHeaderPadY*2
	defaultShadow := widget.RGBA8(82, 108, 138, 34)

	table := Table([]TableRow{{
		{Text: "A", Width: 40, Align: widget.TextAlignLeft},
	}}).(*tableWidget)
	if got := table.Layout(ctx, loose).Height; got != defaultRow {
		t.Fatalf("table height = %.1f, want %.1f", got, defaultRow)
	}
	area := TextArea("hi", 0, nil)
	if got := area.Layout(ctx, loose); got.Width != defaultAreaW || got.Height != defaultAreaH {
		t.Fatalf("text area size = %.1f x %.1f, want %.1f x %.1f", got.Width, got.Height, defaultAreaW, defaultAreaH)
	}
	view := TableView(TableViewRowCount(1))
	viewSize := view.Layout(ctx, geometry.Expand())
	if viewSize.Width != defaultViewW || viewSize.Height != defaultHeader+defaultViewRow {
		t.Fatalf("table view size = %.1f x %.1f, want %.1f x %.1f", viewSize.Width, viewSize.Height, defaultViewW, defaultHeader+defaultViewRow)
	}

	canvas := &uitest.MockCanvas{}
	bounds := geometry.NewRect(0, 0, 40, 16)
	SelectListPainter{}.PaintItemBackground(canvas, listview.ItemPaintState{Bounds: bounds})
	if len(canvas.Rects) != 1 || canvas.Rects[0].Color != defaultFill {
		t.Fatalf("list fill = %+v, want %+v", canvas.Rects, defaultFill)
	}
	field := &uitest.MockCanvas{}
	fieldBounds := geometry.NewRect(0, 0, 80, 22)
	TextFieldPainter{}.PaintTextField(field, &textfield.PaintState{Bounds: fieldBounds})
	if len(field.Rects) < 2 || field.Rects[1].Color != defaultShadow {
		t.Fatalf("shadow = %+v, want %+v", field.Rects, defaultShadow)
	}

	for _, set := range []struct{ section, key, raw string }{
		{"table", "row_height", "28"},
		{"table", "view_row_height", "48"},
		{"table", "view_fallback_width", "300"},
		{"text_area", "width", "360"},
		{"listview", "row_fill", "#e02020"},
		{"textfield", "shadow_color", "#112233"},
	} {
		if err := SetToken(set.section, set.key, set.raw); err != nil {
			t.Fatalf("set %s.%s: %v", set.section, set.key, err)
		}
	}

	if got := table.Layout(ctx, loose).Height; got != 28 {
		t.Fatalf("reloaded table height = %.1f, want 28", got)
	}
	if got := area.Layout(ctx, loose).Width; got != 360 {
		t.Fatalf("reloaded text area width = %.1f, want 360", got)
	}
	viewSize = view.Layout(ctx, geometry.Expand())
	if viewSize.Width != 300 || viewSize.Height != defaultHeader+48 {
		t.Fatalf("reloaded table view size = %.1f x %.1f, want 300 x %.1f", viewSize.Width, viewSize.Height, defaultHeader+48)
	}
	canvas.Rects = nil
	SelectListPainter{}.PaintItemBackground(canvas, listview.ItemPaintState{Bounds: bounds})
	if got := canvas.Rects[0].Color; got != widget.RGBA8(0xe0, 0x20, 0x20, 255) {
		t.Fatalf("reloaded list fill = %+v", got)
	}
	field.Rects = nil
	TextFieldPainter{}.PaintTextField(field, &textfield.PaintState{Bounds: fieldBounds})
	if got := field.Rects[1].Color; got != widget.RGBA8(0x11, 0x22, 0x33, 34) {
		t.Fatalf("reloaded shadow = %+v", got)
	}

	ResetLayout()
	if got := table.Layout(ctx, loose).Height; got != defaultRow {
		t.Fatalf("restored table height = %.1f, want %.1f", got, defaultRow)
	}
	if got := area.Layout(ctx, loose); got.Width != defaultAreaW || got.Height != defaultAreaH {
		t.Fatalf("restored text area size = %.1f x %.1f, want %.1f x %.1f", got.Width, got.Height, defaultAreaW, defaultAreaH)
	}
	viewSize = view.Layout(ctx, geometry.Expand())
	if viewSize.Width != defaultViewW || viewSize.Height != defaultHeader+defaultViewRow {
		t.Fatalf("restored table view size = %.1f x %.1f, want %.1f x %.1f", viewSize.Width, viewSize.Height, defaultViewW, defaultHeader+defaultViewRow)
	}
	canvas.Rects = nil
	SelectListPainter{}.PaintItemBackground(canvas, listview.ItemPaintState{Bounds: bounds})
	if canvas.Rects[0].Color != defaultFill {
		t.Fatalf("restored list fill = %+v, want %+v", canvas.Rects[0].Color, defaultFill)
	}
}
