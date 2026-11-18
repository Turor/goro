package ui

import (
	"fmt"
	"strings"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/ui/rotheme"
)

type BookWindow struct {
	Window
	title      string
	content    res.BookContent
	lines      []itemInfoTextLine
	page       int
	totalPages int
}

func (w *BookWindow) Open(ctx Context, itemID uint32, title string) error {
	if ctx.Resources == nil {
		return fmt.Errorf("book resources unavailable")
	}
	content, err := ctx.Resources.LoadBook(itemID)
	if err != nil {
		return fmt.Errorf("read book %d: %w", itemID, err)
	}
	w.title = strings.TrimSpace(title)
	if w.title == "" {
		w.title = fmt.Sprintf("Book #%d", itemID)
	}
	w.content = content
	w.lines = wrapBookTextLines(content.Lines)
	if !bookLinesHaveContent(w.lines) {
		w.lines = []itemInfoTextLine{parseItemInfoTextLine("This book is empty.")}
	}
	w.page = 0
	w.totalPages = maxInt(1, (len(w.lines)+bookWindowLinesPerPage-1)/bookWindowLinesPerPage)
	w.EnsureWindow(bookWindowWidth, bookWindowHeight)
	w.SetSize(bookWindowWidth, bookWindowHeight)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(bookWindowWidth, bookWindowHeight)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
	return nil
}

func bookLinesHaveContent(lines []itemInfoTextLine) bool {
	for _, line := range lines {
		if strings.TrimSpace(itemInfoLinePlainText(line)) != "" {
			return true
		}
	}
	return false
}

func (w *BookWindow) Update(ctx Context) bool {
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *BookWindow) Rebind(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *BookWindow) widgetTree(ctx Context) widget.Widget {
	background := Color(w.content.Background)
	lineWidgets := make([]widget.Widget, 0, bookWindowLinesPerPage)
	for _, line := range w.pageLines() {
		lineWidgets = append(lineWidgets, itemInfoTextLineWidget(line))
	}
	for len(lineWidgets) < bookWindowLinesPerPage {
		lineWidgets = append(lineWidgets, itemInfoTextLineWidget(itemInfoTextLine{}))
	}
	return Win(
		Title(w.windowTitle()),
		CloseButton(true),
		OnClose(func() {
			w.Window.Close()
			w.Publish(ctx)
		}),
		Size(float32(bookWindowWidth), float32(bookWindowHeight)),
		Background(background),
		Content(
			primitives.Box(
				primitives.Box(lineWidgets...).
					Gap(0).
					Height(float32(bookWindowLinesPerPage*bookWindowLineHeight)),
				primitives.Expanded(primitives.Box()),
				primitives.HBox(
					rotheme.ButtonDisabled("Previous", w.page == 0, func() { w.setPage(ctx, w.page-1) }),
					primitives.Expanded(primitives.Box()),
					rotheme.ButtonDisabled("Next", w.page+1 >= w.totalPages, func() { w.setPage(ctx, w.page+1) }),
				).
					CrossAlign(primitives.CrossAxisCenter).
					Height(float32(bookWindowNavigationH)),
			).
				PaddingXY(float32(bookWindowPaddingX), float32(bookWindowPaddingY)).
				Background(background),
		),
	)
}

func (w *BookWindow) setPage(ctx Context, page int) {
	page = clampWindowInt(page, 0, maxInt(0, w.totalPages-1))
	if page == w.page {
		return
	}
	w.page = page
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *BookWindow) pageLines() []itemInfoTextLine {
	start := w.page * bookWindowLinesPerPage
	if start < 0 || start >= len(w.lines) {
		return nil
	}
	end := minInt(start+bookWindowLinesPerPage, len(w.lines))
	return w.lines[start:end]
}

func (w *BookWindow) windowTitle() string {
	return fmt.Sprintf("%s  (%d/%d)", w.title, w.page+1, maxInt(1, w.totalPages))
}

func wrapBookTextLines(lines []string) []itemInfoTextLine {
	parsed := make([]itemInfoTextLine, 0, len(lines))
	for _, line := range lines {
		parsed = append(parsed, parseItemInfoTextLine(line))
	}
	return wrapItemInfoTextLines(parsed, bookWindowLineRunes)
}
