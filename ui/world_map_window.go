package ui

import (
	"fmt"
	"image"
	"math"
	"slices"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
	xdraw "golang.org/x/image/draw"
)

// WorldMapWindow is the illustrated atlas and the Warpra warper. Artwork and
// previews are cached; the closed window does no presentation work.
type WorldMapWindow struct {
	Window
	Confirm       ConfirmModal
	Picker        DungeonFloorPicker
	art           image.Image
	scaled        image.Image
	board         *worldMapWidget
	previews      map[string]image.Image
	showNames     bool
	memberBuf     []worldMapMember
	dotBuf        []worldMapDot
	marks         []worldMapExploreMark
	markBuf       []worldMapExploreMark
	catalog       session.WarpraCatalog
	onOpen        func()
	onNeedList    func()
	onWarp        func(network.WarpraLoc)
	onBuy         func(network.WarpraLoc)
	pending       network.WarpraLoc
	hasPending    bool
	warpAfter     bool
	lastDungeon   string
	pendingMap    string
	pendingRegion int
	atlases       []res.WorldMapAtlas
	atlasIndex    int
	sections      []res.WorldMapEntry
}

type worldMapMember struct {
	ID      uint32
	Name    string
	MapName string
}

type worldMapDot struct {
	ID   uint32
	X, Y int
}

func (w *WorldMapWindow) Toggle(ctx Context) error {
	if w.IsOpen() {
		w.Close()
		return nil
	}
	return w.OpenWindow(ctx)
}

func (w *WorldMapWindow) OpenWindow(ctx Context) error {
	w.atlases = res.LiveWorldMapAtlases()
	if w.art == nil && ctx.Resources != nil && len(w.atlases) > 0 {
		if img, _, err := res.LoadImage(ctx.Resources, res.WorldMapImageCandidates(w.atlases[0].ID)); err == nil {
			w.art = img
			w.sections = w.atlases[0].Entries()
			w.atlasIndex = 0
		}
	}
	if w.art == nil {
		if ctx.Resources == nil || len(ctx.Resources.WorldMapEntries()) == 0 {
			return fmt.Errorf("mapPosTable.txt is missing or empty")
		}
		var err error
		w.art, _, err = res.LoadImage(ctx.Resources, []string{
			"data/texture/유저인터페이스/worldmap.bmp", "data/texture/interface/worldmap.bmp", "worldmap.bmp",
		})
		if err != nil {
			return fmt.Errorf("worldmap.bmp could not be loaded")
		}
	}
	screenW, screenH := ctx.ScreenSize()
	frameW, frameH := w.frameSize(screenW, screenH)
	w.EnsureWindow(frameW, frameH)
	w.SetSize(frameW, frameH)
	w.setTitleButtonCount(2)
	mapSize := w.mapSize(screenW, screenH)
	if w.scaled == nil || w.scaled.Bounds().Size() != mapSize {
		w.scaled = scaleWorldMapImage(w.art, mapSize)
	}
	if w.previews == nil {
		w.previews = make(map[string]image.Image)
	}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(w.reloadLayout)
	w.UpdatePresentation(ctx)
	w.Publish(ctx)
	if w.onOpen != nil {
		w.onOpen()
	}
	return nil
}

func (w *WorldMapWindow) BindWarpra(onOpen, onNeedList func(), onWarp, onBuy func(network.WarpraLoc)) {
	w.onOpen = onOpen
	w.onNeedList = onNeedList
	w.onWarp = onWarp
	w.onBuy = onBuy
}

func (w *WorldMapWindow) reloadLayout() {
	if w.art == nil || !w.IsOpen() {
		return
	}
	hovered := -1
	if w.board != nil {
		hovered = w.board.hovered
	}
	screenW, screenH := w.ctx.ScreenSize()
	frameW, frameH := w.frameSize(screenW, screenH)
	w.SetSize(frameW, frameH)
	mapSize := w.mapSize(screenW, screenH)
	if w.scaled == nil || w.scaled.Bounds().Size() != mapSize {
		w.scaled = scaleWorldMapImage(w.art, mapSize)
	}
	w.SetContent(w.widgetTree(w.ctx))
	if w.board != nil {
		if hovered < 0 || hovered >= len(w.board.entries) {
			hovered = -1
		}
		w.board.hovered = hovered
	}
	w.UpdatePresentation(w.ctx)
}

func (w *WorldMapWindow) Rebind(ctx Context) {
	if w.IsOpen() && w.art != nil {
		// Replace callbacks and the presentation snapshot copied across maps.
		w.RebindContent(ctx, w.widgetTree(ctx))
		w.UpdatePresentation(ctx)
	}
}

func (w *WorldMapWindow) Update(ctx Context) bool {
	if w.Confirm.Update(ctx) {
		return true
	}
	if w.Picker.Update(ctx) {
		return true
	}
	if !w.IsOpen() {
		return false
	}
	w.loadSettledPreview()
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *WorldMapWindow) Close() {
	w.Confirm.Close(w.ctx)
	w.Picker.Close()
	w.hasPending = false
	w.warpAfter = false
	w.lastDungeon = ""
	w.Window.Close()
}

func (w *WorldMapWindow) widgetTree(ctx Context) widget.Widget {
	w.board = &worldMapWidget{
		resources: ctx.Resources, entries: w.activeEntries(ctx),
		image: w.scaled, sourceSize: w.art.Bounds().Size(), previews: w.previews,
		hovered: -1, showNames: w.showNames, marks: w.marks,
		onClick: func(index int) { w.clickEntry(ctx, index) },
	}
	w.board.SetVisible(true)
	w.board.SetEnabled(true)
	w.board.SetRepaintBoundary(true)
	layers := newWorldMapLayers(w.board)
	content := widget.Widget(layers)
	if rail := w.regionRail(ctx); rail != nil {
		content = primitives.HBox(rail, layers).Gap(worldMapPreviewEdge)
	}
	return Win(Title("World Map"), CloseButton(true), OnClose(w.Close),
		TitleButton(rotheme.IconButtonSearch, func() {
			w.showNames = !w.showNames
			w.board.showNames = w.showNames
			w.redraw(ctx)
		}),
		Size(float32(w.width), float32(w.baseHeight())), Content(content))
}

func (w *WorldMapWindow) activeEntries(ctx Context) []res.WorldMapEntry {
	if len(w.sections) > 0 {
		return w.sections
	}
	if ctx.Resources == nil {
		return nil
	}
	return ctx.Resources.WorldMapEntries()
}

func (w *WorldMapWindow) hasRegionRail() bool {
	return len(w.atlases) > 1
}

func (w *WorldMapWindow) mapSize(screenW, screenH int) image.Point {
	availW := screenW
	if w.hasRegionRail() {
		availW -= worldMapRailW + int(worldMapPreviewEdge)
	}
	return worldMapSize(w.art.Bounds().Size(), availW, screenH)
}

func (w *WorldMapWindow) frameSize(screenW, screenH int) (int, int) {
	size := w.mapSize(screenW, screenH)
	width := size.X
	if w.hasRegionRail() {
		width += worldMapRailW + int(worldMapPreviewEdge)
	}
	return width, size.Y + ROWindowTitleHeight
}

func (w *WorldMapWindow) selectAtlas(ctx Context, index int) {
	if ctx.Resources == nil || index < 0 || index >= len(w.atlases) {
		return
	}
	img, _, err := res.LoadImage(ctx.Resources, res.WorldMapImageCandidates(w.atlases[index].ID))
	if err != nil {
		return
	}
	w.atlasIndex = index
	w.art = img
	w.scaled = nil
	w.sections = w.atlases[index].Entries()
	if w.IsOpen() {
		w.reloadLayout()
	}
}

func (w *WorldMapWindow) regionRail(ctx Context) widget.Widget {
	if len(w.atlases) < 2 {
		return nil
	}
	buttons := make([]widget.Widget, 0, len(w.atlases))
	for i, atlas := range w.atlases {
		i, name := i, atlas.Name
		buttons = append(buttons, rotheme.Button(name, func() { w.selectAtlas(ctx, i) }))
	}
	return primitives.Box(buttons...).Width(float32(worldMapRailW)).Gap(worldMapPreviewEdge)
}

// Capture only data the atlas uses, never live session/world pointers in Draw.
// Walking only invalidates the atlas while its current-map preview is visible.
func (w *WorldMapWindow) UpdatePresentation(ctx Context) {
	if !w.IsOpen() || w.board == nil {
		return
	}
	b := w.board
	current := ""
	if ctx.World != nil {
		current = normalizeMinimapMapName(ctx.World.MapName)
	}
	w.memberBuf = w.memberBuf[:0]
	w.dotBuf = w.dotBuf[:0]
	mapW, mapH := 0, 0
	showPosition := b.hovered >= 0 && b.entries[b.hovered].MapName == current
	if showPosition && ctx.World != nil {
		mapW, mapH = minimapWorldSize(ctx.World)
		w.dotBuf = append(w.dotBuf, worldMapDot{X: ctx.World.Player.X, Y: ctx.World.Player.Y})
	}
	if ctx.Session != nil {
		for _, member := range ctx.Session.Party.Members {
			if member.AccountID == 0 || member.AccountID == ctx.Session.AccountID || !member.Online() {
				continue
			}
			name := normalizeMinimapMapName(member.MapName)
			w.memberBuf = append(w.memberBuf, worldMapMember{ID: member.AccountID, Name: member.Name, MapName: name})
			if showPosition && name == current && member.X >= 0 && member.Y >= 0 {
				w.dotBuf = append(w.dotBuf, worldMapDot{ID: member.AccountID, X: member.X, Y: member.Y})
			}
		}
	}
	if ctx.Session != nil {
		w.catalog = ctx.Session.Warpra
	}
	w.markBuf = appendWorldMapExploreMarks(w.markBuf[:0], b.entries, w.catalog)
	if b.currentMap == current && b.mapW == mapW && b.mapH == mapH && slices.Equal(b.members, w.memberBuf) && slices.Equal(b.dots, w.dotBuf) && exploreMarksEqual(b.marks, w.markBuf) {
		return
	}
	atlasChanged := b.currentMap != current || !slices.Equal(b.members, w.memberBuf) || !exploreMarksEqual(b.marks, w.markBuf)
	b.currentMap, b.mapW, b.mapH = current, mapW, mapH
	b.members = slices.Clone(w.memberBuf)
	b.indexMembers()
	b.dots = slices.Clone(w.dotBuf)
	b.marks = append(b.marks[:0], w.markBuf...)
	w.marks = b.marks
	if atlasChanged {
		w.redraw(ctx)
		return
	}
	w.markHighlight(ctx)
}

func exploreMarksEqual(a, b []worldMapExploreMark) bool {
	return slices.Equal(a, b)
}

func (w *WorldMapWindow) SetCatalog(ctx Context, catalog session.WarpraCatalog) {
	w.catalog = catalog
	if w.lastDungeon != "" && w.Picker.IsOpen() {
		if info, ok := catalog.DungeonFloors(w.lastDungeon); ok {
			w.Picker.Refresh(ctx, info, catalog.UnlockPrice(), catalog.Zeny)
		}
	}
	if w.hasPending {
		if next, ok := catalog.Find(w.pending.Kind, w.pending.ID); ok && next.Unlocked {
			w.Confirm.Close(ctx)
			w.hasPending = false
		}
	}
	w.UpdatePresentation(ctx)
	if w.pendingMap != "" && catalog.Ready() && w.board != nil {
		name := w.pendingMap
		w.pendingMap = ""
		for i, entry := range w.board.entries {
			if entry.MapName == name {
				w.clickEntry(ctx, i)
				break
			}
		}
	}
}

func (w *WorldMapWindow) clickEntry(ctx Context, index int) {
	if w.board == nil || index < 0 || index >= len(w.board.entries) {
		return
	}
	entry := w.board.entries[index]
	if !w.catalog.Ready() {
		w.pendingMap = entry.MapName
		w.pendingRegion = entry.Region
		if w.onNeedList != nil {
			w.onNeedList()
		} else if w.onOpen != nil {
			w.onOpen()
		}
		return
	}
	regionMaps := regionMapNames(w.board.entries, entry.Region)
	action := resolveWorldMapWarp(w.catalog, entry.MapName, entry.Region, regionMaps)
	if !action.OK {
		return
	}
	if action.OpenFloors {
		w.lastDungeon = action.Loc.Map
		w.Picker.OpenInfo(ctx, action.Dungeon, w.catalog.UnlockPrice(), w.catalog.Zeny, func(loc network.WarpraLoc) {
			w.UnlockOrWarp(ctx, loc, true)
		})
		return
	}
	w.Picker.Close()
	w.lastDungeon = ""
	w.UnlockOrWarp(ctx, action.Loc, true)
}

func (w *WorldMapWindow) UnlockOrWarp(ctx Context, loc network.WarpraLoc, autoWarp bool) {
	if loc.Unlocked {
		w.warpAfter = false
		w.hasPending = false
		if w.onWarp != nil {
			w.onWarp(loc)
		}
		return
	}
	if session.WarpraLevelGate(loc) > 0 {
		if w.onWarp != nil {
			w.onWarp(loc)
		}
		return
	}
	w.pending = loc
	w.hasPending = true
	w.warpAfter = autoWarp
	title, body := warpraUnlockMessage(loc, w.catalog.UnlockPrice(), w.catalog.Zeny)
	w.Confirm.Open(ctx, title, body, func() {
		if w.onBuy != nil {
			w.onBuy(w.pending)
		}
	}, func() {
		w.warpAfter = false
		w.hasPending = false
	})
}

func (w *WorldMapWindow) ConsumePendingBuy() (network.WarpraLoc, bool, bool) {
	if !w.hasPending {
		return network.WarpraLoc{}, false, false
	}
	return w.pending, w.warpAfter, true
}

func (w *WorldMapWindow) ClearPending() {
	w.hasPending = false
	w.warpAfter = false
}

func (w *WorldMapWindow) CloseAfterWarp() {
	w.Picker.Close()
	w.lastDungeon = ""
	w.Confirm.Close(w.ctx)
	w.ClearPending()
	w.Window.Close()
}

func (w *WorldMapWindow) redraw(ctx Context) {
	w.board.SetNeedsRedraw(true)
	w.markHighlight(ctx)
}

func (w *WorldMapWindow) markHighlight(ctx Context) {
	if w.board != nil && w.board.previewLayer != nil {
		w.board.previewLayer.SetNeedsRedraw(true)
	}
	invalidateWindowRect(ctx, windowFrameRect(w.x, w.y, w.width, w.height))
}

func worldMapSize(source image.Point, screenW, screenH int) image.Point {
	scale := min(1., float64(max(1, screenW-windowScreenMargin*2))/float64(source.X),
		float64(max(1, screenH-windowScreenMargin*2-ROWindowTitleHeight))/float64(source.Y))
	return image.Pt(max(1, int(float64(source.X)*scale)), max(1, int(float64(source.Y)*scale)))
}

func scaleWorldMapImage(img image.Image, size image.Point) image.Image {
	return scaleWorldMap(img, size, xdraw.ApproxBiLinear)
}

func scaleWorldMapPreview(img image.Image, size image.Point) image.Image {
	return scaleWorldMap(img, size, xdraw.NearestNeighbor)
}

func scaleWorldMap(img image.Image, size image.Point, scaler xdraw.Scaler) image.Image {
	dst := image.NewRGBA(image.Rectangle{Max: size})
	scaler.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Src, nil)
	return dst
}

type worldMapWidget struct {
	widget.WidgetBase
	resources    *res.Manager
	entries      []res.WorldMapEntry
	image        image.Image
	sourceSize   image.Point
	previews     map[string]image.Image
	hovered      int
	showNames    bool
	named        *image.RGBA
	namedFor     image.Image
	currentMap   string
	members      []worldMapMember
	memberByMap  map[string]uint32
	dots         []worldMapDot
	marks        []worldMapExploreMark
	onClick      func(int)
	highlight    *worldMapHighlight
	previewLayer *worldMapPreview
	hoverArmed   bool
	mapW, mapH   int
}

func (w *worldMapWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := w.image.Bounds().Size()
	result := constraints.Constrain(geometry.Sz(float32(size.X), float32(size.Y)))
	w.SetBounds(geometry.FromPointSize(w.Position(), result))
	return result
}

func (w *worldMapWidget) mapRect(entry res.WorldMapEntry) geometry.Rect {
	b := w.Bounds()
	sx, sy := b.Width()/float32(w.sourceSize.X), b.Height()/float32(w.sourceSize.Y)
	r := entry.Bounds
	return geometry.NewRect(b.Min.X+float32(r.Min.X)*sx, b.Min.Y+float32(r.Min.Y)*sy, float32(r.Dx())*sx, float32(r.Dy())*sy)
}

func (w *worldMapWidget) mapAt(point geometry.Point) int {
	// Later sections are drawn on top. A point inside two cells uses the later
	// one. A point in the gap between neighbors stays with the nearer cell.
	hit := -1
	best := worldMapHitSlack * worldMapHitSlack
	for i, entry := range w.entries {
		dist := worldMapRectDistance(w.mapRect(entry), point)
		if dist <= best {
			hit = i
			best = dist
		}
	}
	return hit
}

func worldMapRectDistance(r geometry.Rect, p geometry.Point) float32 {
	x, y := p.X, p.Y
	if x < r.Min.X {
		x = r.Min.X
	} else if x > r.Max.X {
		x = r.Max.X
	}
	if y < r.Min.Y {
		y = r.Min.Y
	} else if y > r.Max.Y {
		y = r.Max.Y
	}
	dx, dy := p.X-x, p.Y-y
	return dx*dx + dy*dy
}

func (w *worldMapWidget) Event(ctx widget.Context, e event.Event) bool {
	mouse, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	index := w.hovered
	switch mouse.MouseType {
	case event.MouseEnter, event.MouseMove, event.MousePress:
		// Hover-enter is delivered directly, whereas moves pass through parent
		// transforms. GlobalPosition gives both paths the same coordinates.
		point := mouse.GlobalPosition.Sub(w.ScreenBounds().Min).Add(w.Bounds().Min)
		index = w.mapAt(point)
	case event.MouseLeave:
		index = -1
	}
	if index != w.hovered {
		w.hovered = index
		w.markHoverDirty(ctx)
	}
	if mouse.MouseType == event.MousePress && mouse.Button == event.ButtonLeft && w.hovered >= 0 && w.onClick != nil {
		w.onClick(w.hovered)
	}
	// Clicks never issue a movement command; they warp through Warpra instead.
	return true
}

func (w *WorldMapWindow) loadSettledPreview() {
	b := w.board
	if b == nil || b.hovered < 0 || b.hovered >= len(b.entries) {
		if b != nil {
			b.hoverArmed = false
		}
		return
	}
	if !b.hoverArmed {
		b.hoverArmed = true
		return
	}
	name := b.entries[b.hovered].MapName
	if _, cached := b.previews[name]; cached {
		return
	}
	b.loadPreview(name)
	if b.previewLayer != nil {
		b.previewLayer.SetNeedsRedraw(true)
	}
}

func (w *worldMapWidget) loadPreview(name string) {
	if _, cached := w.previews[name]; cached {
		return
	}
	var preview image.Image
	if w.resources != nil {
		if img, _, err := res.LoadImage(w.resources, minimapImageCandidates(name)); err == nil {
			size := img.Bounds().Size()
			scale := float64(worldMapPreviewMax) / float64(max(size.X, size.Y))
			preview = scaleWorldMapPreview(img, image.Pt(max(1, int(float64(size.X)*scale)), max(1, int(float64(size.Y)*scale))))
		}
	}
	w.previews[name] = preview // Missing minimaps are normal, and cached too.
}

func (w *worldMapWidget) indexMembers() {
	if len(w.members) == 0 {
		w.memberByMap = nil
		return
	}
	if w.memberByMap == nil {
		w.memberByMap = make(map[string]uint32, len(w.members))
	}
	for name := range w.memberByMap {
		delete(w.memberByMap, name)
	}
	for i := range w.members {
		w.memberByMap[w.members[i].MapName] = w.members[i].ID
	}
}

func (w *worldMapWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	canvas.PushClip(bounds)
	defer canvas.PopClip()
	canvas.DrawImage(w.atlasImage(), bounds.Min)
	for i, entry := range w.entries {
		r := w.mapRect(entry)
		if id, ok := w.memberByMap[entry.MapName]; ok {
			canvas.StrokeRect(r, Color(minimapMemberColor(id)), worldMapMemberStroke)
		}
		if entry.MapName == w.currentMap {
			drawWorldMapStar(canvas, geometry.Pt(r.Min.X+r.Width()/2, r.Min.Y+r.Height()/2))
		}
		if i < len(w.marks) && w.marks[i].Present {
			drawWorldMapExploreMark(canvas, r, w.marks[i].Unlocked)
		}
	}
}

func drawWorldMapExploreMark(canvas widget.Canvas, tile geometry.Rect, unlocked bool) {
	size := worldMapMarkSize
	cx := tile.Max.X - size/2 - worldMapMarkInset
	cy := tile.Min.Y + size/2 + worldMapMarkInset
	if cx < tile.Min.X+size/2 {
		cx = tile.Min.X + tile.Width()/2
	}
	if cy > tile.Max.Y-size/2 {
		cy = tile.Min.Y + tile.Height()/2
	}
	color := Color(ErrorTextColor)
	if unlocked {
		color = Color(GoodTextColor)
	}
	if unlocked {
		canvas.DrawLine(geometry.Pt(cx-worldMapMarkArm, cy), geometry.Pt(cx-worldMapMarkNotch, cy+worldMapMarkDrop), color, worldMapMarkStroke)
		canvas.DrawLine(geometry.Pt(cx-worldMapMarkNotch, cy+worldMapMarkDrop), geometry.Pt(cx+worldMapMarkTip, cy-worldMapMarkDrop), color, worldMapMarkStroke)
		return
	}
	canvas.DrawLine(geometry.Pt(cx-worldMapMarkArm, cy-worldMapMarkArm), geometry.Pt(cx+worldMapMarkArm, cy+worldMapMarkArm), color, worldMapMarkStroke)
	canvas.DrawLine(geometry.Pt(cx+worldMapMarkArm, cy-worldMapMarkArm), geometry.Pt(cx-worldMapMarkArm, cy+worldMapMarkArm), color, worldMapMarkStroke)
}

func drawWorldMapStar(canvas widget.Canvas, center geometry.Point) {
	// A small geometric star stays legible without depending on font coverage.
	var points [5]geometry.Point
	for i := range points {
		// theme:ok five-point star angle
		a := float64(i)*4*math.Pi/5 - math.Pi/2
		points[i] = geometry.Pt(center.X+float32(math.Cos(a))*worldMapStarRadius, center.Y+float32(math.Sin(a))*worldMapStarRadius)
	}
	for i, point := range points {
		canvas.DrawLine(point, points[(i+1)%len(points)], Color(minimapPlayerColor), worldMapStarStroke)
	}
}

func (w *worldMapWidget) previewPanel(entry res.WorldMapEntry) geometry.Rect {
	b := w.Bounds()
	width := min(worldMapPreviewW, b.Width()-worldMapPreviewEdge)
	rows := 0
	for _, member := range w.members {
		if member.MapName == entry.MapName {
			rows++
		}
	}
	rows = min(rows, max(0, int((b.Height()-worldMapPreviewRowsTop)/worldMapPreviewRowH)))
	height := min(worldMapPreviewBaseH+float32(rows)*worldMapPreviewRowH, b.Height()-worldMapPreviewEdge)
	return geometry.NewRect(b.Max.X-width-worldMapPreviewInset, b.Min.Y+worldMapPreviewInset, width, height)
}

func (w *worldMapWidget) drawPreview(canvas widget.Canvas, entry res.WorldMapEntry) {
	b := w.Bounds()
	width := min(worldMapPreviewW, b.Width()-worldMapPreviewEdge)
	rows := 0
	for _, member := range w.members {
		if member.MapName == entry.MapName {
			rows++
		}
	}
	rows = min(rows, max(0, int((b.Height()-worldMapPreviewRowsTop)/worldMapPreviewRowH)))
	height := min(worldMapPreviewBaseH+float32(rows)*worldMapPreviewRowH, b.Height()-worldMapPreviewEdge)
	panel := geometry.NewRect(b.Max.X-width-worldMapPreviewInset, b.Min.Y+worldMapPreviewInset, width, height)
	canvas.DrawRoundRect(panel, worldMapPreviewFill, worldMapPreviewRadius)
	canvas.StrokeRoundRect(panel, rotheme.Default.Colors.WindowBorder, worldMapPreviewRadius, worldMapPreviewBorder)
	labelW := width - worldMapPreviewPadX*2
	label := geometry.NewRect(panel.Min.X+worldMapPreviewPadX, panel.Min.Y+worldMapPreviewLabelY, labelW, worldMapPreviewLabelH)
	rotheme.DrawLabel(canvas, worldMapEntryLabel(entry), label, widget.TextAlignCenter)
	img := w.previews[entry.MapName]
	if img == nil {
		rotheme.DrawText(canvas, "No minimap available", geometry.NewRect(panel.Min.X+worldMapPreviewPadX, panel.Min.Y+worldMapPreviewImageY, labelW, worldMapPreviewImageH),
			rotheme.Default.Typography.TextSize, rotheme.Default.Colors.MutedText, false, widget.TextAlignCenter)
	} else {
		size := img.Bounds().Size()
		at := geometry.Pt(panel.Min.X+(width-float32(size.X))/2, panel.Min.Y+worldMapPreviewImageY)
		canvas.DrawImage(img, at)
		if entry.MapName == w.currentMap {
			rect := minimapRect{x: int(at.X), y: int(at.Y), w: size.X, h: size.Y}
			for _, dot := range w.dots {
				if x, y, ok := minimapCellToScreen(rect, w.mapW, w.mapH, dot.X, dot.Y); ok {
					fill := minimapPlayerColor
					if dot.ID != 0 {
						fill = minimapMemberColor(dot.ID)
					}
					drawMinimapSquareCentered(canvas, x, y, worldMapDotSize, fill)
				}
			}
		}
	}
	y := panel.Min.Y + worldMapPreviewMemberY
	for _, member := range w.members {
		if member.MapName != entry.MapName || rows == 0 {
			continue
		}
		rotheme.DrawText(canvas, trimRunes(member.Name, 24), geometry.NewRect(panel.Min.X+worldMapPreviewPadX, y, labelW, worldMapPreviewRowH), // theme:ok name length limit
			rotheme.Default.Typography.TextSize, rotheme.Default.Colors.Text, false, widget.TextAlignLeft)
		y += worldMapPreviewRowH
		rows--
	}
}

func worldMapEntryLabel(entry res.WorldMapEntry) string {
	if entry.Name != "" {
		return entry.Name
	}
	return entry.MapName
}
