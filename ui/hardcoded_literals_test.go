package ui

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestNoHardcodedUILiterals keeps every UI size and color in ui-layout.ini.
//
// UI window code (ui, ui/rotheme and the game/login screens) must take each
// size, position offset, padding, row height and color from a layout token.
// The compiled defaults live only in the defaults tables: files named
// theme.go or theme_*.go, where every package-level value must be registered
// as a token (rotheme.Section(...).Int/Float/Color/RGBA/String or
// rotheme.Register*). Anywhere else this test fails on:
//
//   - color literals: color.RGBA{...}, color.NRGBA{...}, widget.RGBA8(...),
//     widget.RGBA(...), widget.Hex(...);
//   - numbers other than 0 and 1 passed to layout and drawing calls (Width,
//     Height, Gap, Padding*, Size, geometry.NewRect, geometry.Pt, Stroke*,
//     Draw*, maxInt/minInt bounds, ...), or converted with float32/float64/int;
//   - numeric constants and package-level variables;
//   - +/- offsets by a number other than 1;
//   - size fields (Width:, Height:, X:, Left:, ...) set from a number;
//   - a number assigned to a size-named variable or field (pad := 6).
//
// A value that is not a visual size or color (a packet flag, skill ID, text
// length limit, timer, list capacity) may stay inline with a reason on the
// same line or the line above:  // theme:ok <reason>
// A theme:ok comment directly above a const or var block covers the block.
//
// Set GORO_LITERAL_FILES to a comma list of base names to check only those
// files (useful while converting a window).
func TestNoHardcodedUILiterals(t *testing.T) {
	violations, annotated, unregistered := scanUILiteralDirs(t, uiLiteralDirs)
	var only map[string]bool
	if env := strings.TrimSpace(os.Getenv("GORO_LITERAL_FILES")); env != "" {
		only = map[string]bool{}
		for _, name := range strings.Split(env, ",") {
			only[strings.TrimSpace(name)] = true
		}
	}
	keep := func(pos string) bool {
		if only == nil {
			return true
		}
		file := pos
		if i := strings.Index(file, ":"); i >= 0 {
			file = file[:i]
		}
		return only[filepath.Base(file)]
	}
	var shown []string
	for _, v := range violations {
		if keep(v) {
			shown = append(shown, v)
		}
	}
	for _, v := range unregistered {
		if keep(v) {
			shown = append(shown, v)
		}
	}
	t.Logf("%d theme:ok exceptions (non-visual values kept inline)", len(annotated))
	if testing.Verbose() {
		for _, a := range annotated {
			t.Log("theme:ok ", a)
		}
	}
	if len(shown) > 0 {
		perFile := map[string]int{}
		for _, v := range shown {
			perFile[strings.SplitN(v, ":", 2)[0]]++
		}
		files := make([]string, 0, len(perFile))
		for f := range perFile {
			files = append(files, f)
		}
		sort.Strings(files)
		var summary strings.Builder
		for _, f := range files {
			fmt.Fprintf(&summary, "  %s: %d\n", f, perFile[f])
		}
		t.Fatalf("%d hardcoded UI sizes or colors outside the theme defaults (move them to a theme_<section>.go token, or mark a non-visual value with // theme:ok <reason>):\n%s\n%s",
			len(shown), summary.String(), strings.Join(shown, "\n"))
	}
}

// Each entry is a package directory, or a glob for UI files in a package that
// also holds non-UI code (the map loading screen lives in game/map; its
// tokens are in rotheme/theme_loading.go).
var uiLiteralDirs = []string{".", "rotheme", filepath.Join("..", "game", "login"), filepath.Join("..", "game", "map", "loading.go")}

func TestHardcodedUILiteralScannerCatchesEachRule(t *testing.T) {
	dir := t.TempDir()
	src := `package x

import (
	"image/color"
	"github.com/gogpu/ui/widget"
)

const rowH = 32
const packetFlag = 0x10 // theme:ok packet bit
// theme:ok skill IDs
const (
	skillA = 26
	skillB = 27
)
var tint = color.RGBA{R: 1, G: 2, B: 3, A: 255}
var c2 = widget.RGBA8(1, 2, 3, 4)

func f(b box, x int) {
	b.Width(120).Gap(4).Height(float32(rowH))
	_ = x + 6
	_ = x + 1
	_ = x / 2
	pad := 6
	_ = rect{Width: 40}
	_ = float32(18)
	_ = pad
	_ = widget.Color{}
}
`
	if err := os.WriteFile(filepath.Join(dir, "window.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	theme := `package x

var shopW = 420
var orphan = 7

func init() { register("shop", "width", &shopW) }
`
	if err := os.WriteFile(filepath.Join(dir, "theme_shop.go"), []byte(theme), 0o644); err != nil {
		t.Fatal(err)
	}
	violations, annotated, unregistered := scanUILiteralDirs(t, []string{dir})
	want := []string{"const rowH", "color.RGBA", "widget.RGBA8", "Width(120)", "Gap(4)", "x + 6", "pad := 6", "Width: 40", "float32(18)"}
	if len(violations) != len(want) {
		t.Fatalf("got %d violations, want %d:\n%s", len(violations), len(want), strings.Join(violations, "\n"))
	}
	if len(annotated) != 3 {
		t.Fatalf("got %d theme:ok exceptions, want 3: %v", len(annotated), annotated)
	}
	if len(unregistered) != 1 || !strings.Contains(unregistered[0], "orphan") {
		t.Fatalf("unregistered = %v", unregistered)
	}
}

var (
	// layoutCallNames are calls whose numeric arguments are sizes, offsets or
	// colors.
	layoutCallNames = map[string]bool{
		"Width": true, "Height": true, "MinWidth": true, "MinHeight": true, "MaxWidth": true, "MaxHeight": true,
		"Size": true, "Gap": true, "Spacing": true, "Padding": true, "PaddingXY": true, "PaddingLTRB": true,
		"PaddingLeft": true, "PaddingRight": true, "PaddingTop": true, "PaddingBottom": true, "Margin": true,
		"Radius": true, "Rounded": true, "CornerRadius": true, "BorderStyle": true, "BorderWidth": true,
		"FontSize": true, "LineHeight": true, "Offset": true, "Position": true, "SetAutoPosition": true,
		"OpenAt": true, "SetSize": true, "EnsureWindow": true, "NewWindow": true, "setOpacity": true,
		"NewRect": true, "Rect": true, "Pt": true, "Sz": true, "UniformInsets": true, "Insets": true, "TranslateXY": true,
		"DrawLine": true, "DrawRect": true, "StrokeRect": true, "StrokeRoundRect": true, "DrawRoundRect": true,
		"FillRect": true, "FillRoundRect": true, "DrawCircle": true, "StrokeCircle": true, "FillCircle": true,
		"DrawText": true, "DrawImage": true, "TableViewRowHeight": true, "TableViewHeaderHeight": true,
		"ItemHeight": true, "RowHeight": true, "ColumnWidth": true, "IconSize": true,
		"maxInt": true, "minInt": true, "max": true, "min": true, "clampWindowInt": true,
		"Lerp": true, "WithAlpha": true, "Alpha": true,
	}
	sizeFieldNames = map[string]bool{
		"X": true, "Y": true, "W": true, "H": true, "Width": true, "Height": true, "Left": true, "Right": true,
		"Top": true, "Bottom": true, "Size": true, "Radius": true, "Pad": true, "Padding": true, "Gap": true,
		"MinWidth": true, "MinHeight": true, "MaxWidth": true, "MaxHeight": true, "Thickness": true,
	}
	sizeNameRE = regexp.MustCompile(`(?i)(width|height|^[wh]$|[a-z][WH]$|pad|gap|margin|offset|inset|indent|radius|border|spacing|size|rowh|lineh|^[xy]$|[a-z][XY]$|^d[xy]$|top|left|right|bottom|thick|stroke|alpha|cell|slot|font)`)
	themeOKRE  = regexp.MustCompile(`theme:ok\s+\S`)
)

func isDefaultsTable(path string) bool {
	base := filepath.Base(path)
	return base == "theme.go" || strings.HasPrefix(base, "theme_")
}

// scanUILiteralDirs returns hardcoded literal violations, theme:ok exceptions
// and defaults-table values that are not registered as tokens.
func scanUILiteralDirs(t *testing.T, dirs []string) (violations, annotated, unregistered []string) {
	t.Helper()
	for _, dir := range dirs {
		pattern := filepath.Join(dir, "*.go")
		if strings.ContainsAny(dir, "*?[") || strings.HasSuffix(dir, ".go") {
			pattern = dir
		}
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) == 0 {
			t.Fatalf("literal guard: %s matches no Go files", pattern)
		}
		sort.Strings(files)
		fset := token.NewFileSet()
		registered := map[string]bool{}
		type themeVar struct{ name, pos string }
		var themeVars []themeVar
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			collectRegistered(file, registered)
			if isDefaultsTable(path) {
				for _, decl := range file.Decls {
					gen, ok := decl.(*ast.GenDecl)
					if !ok || gen.Tok != token.VAR {
						continue
					}
					for _, spec := range gen.Specs {
						vs := spec.(*ast.ValueSpec)
						if !specHasLiteral(vs) {
							continue
						}
						for _, name := range vs.Names {
							themeVars = append(themeVars, themeVar{name.Name, relPos(fset, name.Pos())})
						}
					}
				}
				continue
			}
			s := &literalScan{fset: fset, file: file, okLines: themeOKLines(fset, file)}
			s.scan()
			violations = append(violations, s.violations...)
			annotated = append(annotated, s.annotated...)
		}
		for _, v := range themeVars {
			if !registered[v.name] {
				unregistered = append(unregistered, fmt.Sprintf("%s: defaults-table value %s is not registered as a layout token", v.pos, v.name))
			}
		}
	}
	return violations, annotated, unregistered
}

func relPos(fset *token.FileSet, pos token.Pos) string {
	p := fset.Position(pos)
	name := p.Filename
	if rel, err := filepath.Rel(".", name); err == nil {
		name = rel
	}
	return fmt.Sprintf("%s:%d", name, p.Line)
}

// collectRegistered records every &ident passed to a call (token registration
// takes a pointer to the variable).
func collectRegistered(file *ast.File, registered map[string]bool) {
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, arg := range call.Args {
			unary, ok := arg.(*ast.UnaryExpr)
			if !ok || unary.Op != token.AND {
				continue
			}
			switch x := unary.X.(type) {
			case *ast.Ident:
				registered[x.Name] = true
			case *ast.SelectorExpr:
				// &theme.Field registers the struct var's field; count the root.
				for {
					inner, ok := x.X.(*ast.SelectorExpr)
					if !ok {
						break
					}
					x = inner
				}
				if id, ok := x.X.(*ast.Ident); ok {
					registered[id.Name] = true
				}
			}
		}
		return true
	})
}

func specHasLiteral(vs *ast.ValueSpec) bool {
	found := false
	for _, v := range vs.Values {
		ast.Inspect(v, func(n ast.Node) bool {
			if _, ok := n.(*ast.FuncLit); ok {
				return false
			}
			switch x := n.(type) {
			case *ast.BasicLit:
				if x.Kind == token.INT || x.Kind == token.FLOAT {
					found = true
				}
			case *ast.CompositeLit:
				found = true
			}
			return !found
		})
	}
	return found
}

// themeOKLines maps line numbers covered by a "theme:ok <reason>" comment: the
// comment's own line, the line after it, and a whole const/var block when the
// comment sits directly above it.
func themeOKLines(fset *token.FileSet, file *ast.File) map[int]string {
	lines := map[int]string{}
	for _, group := range file.Comments {
		for _, c := range group.List {
			if !themeOKRE.MatchString(c.Text) {
				continue
			}
			reason := strings.TrimSpace(c.Text[strings.Index(c.Text, "theme:ok")+len("theme:ok"):])
			reason = strings.TrimSuffix(reason, "*/")
			line := fset.Position(c.Pos()).Line
			lines[line] = reason
			lines[line+1] = reason
		}
	}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Doc == nil {
			continue
		}
		docLine := fset.Position(gen.Doc.End()).Line
		reason, ok := lines[docLine]
		if !ok {
			continue
		}
		for l := fset.Position(gen.Pos()).Line; l <= fset.Position(gen.End()).Line; l++ {
			if _, set := lines[l]; !set {
				lines[l] = reason
			}
		}
	}
	return lines
}

type literalScan struct {
	fset       *token.FileSet
	file       *ast.File
	okLines    map[int]string
	violations []string
	annotated  []string
	seen       map[token.Pos]bool
}

func (s *literalScan) report(pos token.Pos, what string) {
	if s.seen == nil {
		s.seen = map[token.Pos]bool{}
	}
	if s.seen[pos] {
		return
	}
	s.seen[pos] = true
	line := s.fset.Position(pos).Line
	if reason, ok := s.okLines[line]; ok {
		s.annotated = append(s.annotated, fmt.Sprintf("%s: %s (%s)", relPos(s.fset, pos), what, reason))
		return
	}
	s.violations = append(s.violations, fmt.Sprintf("%s: %s", relPos(s.fset, pos), what))
}

func (s *literalScan) src(n ast.Node) string {
	start, end := s.fset.Position(n.Pos()).Offset, s.fset.Position(n.End()).Offset
	data, err := os.ReadFile(s.fset.Position(n.Pos()).Filename)
	if err != nil || end > len(data) || start < 0 || end-start > 120 {
		return fmt.Sprintf("%T", n)
	}
	return string(data[start:end])
}

func numericLit(e ast.Expr) (*ast.BasicLit, bool) {
	if u, ok := e.(*ast.UnaryExpr); ok && (u.Op == token.SUB || u.Op == token.ADD) {
		e = u.X
	}
	if p, ok := e.(*ast.ParenExpr); ok {
		e = p.X
	}
	bl, ok := e.(*ast.BasicLit)
	if !ok || (bl.Kind != token.INT && bl.Kind != token.FLOAT) {
		return nil, false
	}
	return bl, true
}

func litValue(bl *ast.BasicLit) float64 {
	if bl.Kind == token.INT {
		v, err := strconv.ParseInt(strings.ReplaceAll(bl.Value, "_", ""), 0, 64)
		if err == nil {
			return float64(v)
		}
	}
	v, _ := strconv.ParseFloat(strings.ReplaceAll(bl.Value, "_", ""), 64)
	return v
}

func trivialLit(bl *ast.BasicLit) bool {
	v := litValue(bl)
	return v == 0 || v == 1
}

func callName(call *ast.CallExpr) (pkg, name string) {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if id, ok := fn.X.(*ast.Ident); ok {
			return id.Name, fn.Sel.Name
		}
		return "", fn.Sel.Name
	case *ast.Ident:
		return "", fn.Name
	}
	return "", ""
}

func exprName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return x.Sel.Name
	}
	return ""
}

// numbersIn reports numeric literals inside e, skipping x/2 and x*2 halving,
// function literals and nested calls (those are checked on their own).
func (s *literalScan) numbersIn(e ast.Expr, what string) {
	ast.Inspect(e, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncLit, *ast.CallExpr, *ast.IndexExpr:
			return false
		case *ast.BinaryExpr:
			if x.Op == token.QUO || x.Op == token.MUL || x.Op == token.REM || x.Op == token.SHL || x.Op == token.SHR {
				if bl, ok := numericLit(x.Y); ok && (litValue(bl) == 2 || x.Op != token.MUL) {
					s.numbersIn(x.X, what)
					return false
				}
			}
		case *ast.BasicLit:
			if bl, ok := numericLit(x); ok && !trivialLit(bl) {
				s.report(bl.Pos(), what)
			}
		}
		return true
	})
}

func (s *literalScan) scan() {
	ast.Inspect(s.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok != token.CONST && x.Tok != token.VAR {
				return true
			}
			for _, spec := range x.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, v := range vs.Values {
					name := vs.Names[min(i, len(vs.Names)-1)].Name
					if bl, ok := numericLit(v); ok && !trivialLit(bl) {
						s.report(bl.Pos(), fmt.Sprintf("%s %s = %s", x.Tok, name, bl.Value))
					} else if bin, ok := v.(*ast.BinaryExpr); ok && x.Tok == token.CONST {
						s.numbersIn(bin, fmt.Sprintf("%s %s = %s", x.Tok, name, s.src(bin)))
					}
				}
			}
		case *ast.CompositeLit:
			if sel, ok := x.Type.(*ast.SelectorExpr); ok && len(x.Elts) > 0 {
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == "color" {
					s.report(x.Pos(), "color literal "+s.src(x))
					return false
				}
			}
			for _, elt := range x.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if key, ok := kv.Key.(*ast.Ident); ok && sizeFieldNames[key.Name] {
					if bl, ok := numericLit(kv.Value); ok && !trivialLit(bl) {
						s.report(bl.Pos(), "size field "+s.src(kv))
					}
				}
			}
		case *ast.CallExpr:
			pkg, name := callName(x)
			if pkg == "widget" && (strings.HasPrefix(name, "RGB") || name == "Hex") {
				s.report(x.Pos(), "color literal "+s.src(x))
				return false
			}
			if pkg == "color" && (name == "RGBA" || name == "NRGBA") {
				s.report(x.Pos(), "color literal "+s.src(x))
				return false
			}
			if pkg == "" && (name == "float32" || name == "float64" || name == "int") && len(x.Args) == 1 {
				if bl, ok := numericLit(x.Args[0]); ok && !trivialLit(bl) {
					s.report(bl.Pos(), "number "+s.src(x))
				}
			}
			if layoutCallNames[name] {
				for _, arg := range x.Args {
					s.numbersIn(arg, "layout value in "+name+"("+s.src(arg)+")")
				}
			}
		case *ast.BinaryExpr:
			if x.Op == token.ADD || x.Op == token.SUB {
				for _, operand := range []ast.Expr{x.X, x.Y} {
					if bl, ok := numericLit(operand); ok && !trivialLit(bl) {
						s.report(bl.Pos(), "offset "+s.src(x))
					}
				}
			}
		case *ast.AssignStmt:
			for i, rhs := range x.Rhs {
				if i >= len(x.Lhs) {
					break
				}
				name := exprName(x.Lhs[i])
				if name == "" || !sizeNameRE.MatchString(name) {
					continue
				}
				if bl, ok := numericLit(rhs); ok && !trivialLit(bl) {
					s.report(bl.Pos(), "size "+s.src(x))
				}
			}
		case *ast.ReturnStmt:
			// return 24 from a size helper is caught by the call sites above
			// only when inlined, so flag numeric returns in size-named funcs.
		}
		return true
	})
	// Numeric returns from size-named functions (func rowHeight() int { return 24 }).
	for _, decl := range s.file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || !sizeNameRE.MatchString(fn.Name.Name) {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if _, ok := n.(*ast.FuncLit); ok {
				return false
			}
			ret, ok := n.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			for _, r := range ret.Results {
				if bl, ok := numericLit(r); ok && !trivialLit(bl) {
					s.report(bl.Pos(), "size returned by "+fn.Name.Name+"()")
				}
			}
			return true
		})
	}
}
