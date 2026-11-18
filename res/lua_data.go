package res

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yuin/gopher-lua/ast"
)

// errLuaNotData marks a chunk that is more than global table assignments, so
// it has to run in the Lua VM.
var errLuaNotData = errors.New("lua chunk is not a pure data table")

// evalLuaDataChunk evaluates a chunk made only of global assignments of
// literal values: tables, strings, numbers, booleans, and NAME / NAME.FIELD
// references to globals already set (SKID.AL_HEAL, JOBID.JT_NOVICE). That is
// the shape of every big client table (itemInfo, skillinfolist,
// skilldescript, skilltreeview). Walking gopher-lua's AST directly avoids its
// compiler, whose constant de-duplication is quadratic: the 23 MB Turoran
// itemInfo takes ~53 s to compile and ~0.4 s this way.
//
// Unlike the VM, a table key that resolves to nil (a SKID name this client's
// skillid does not define) skips that entry instead of failing the file.
func evalLuaDataChunk(chunk []ast.Stmt, globals map[string]luaValue) error {
	for _, stmt := range chunk {
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != len(assign.Rhs) {
			return errLuaNotData
		}
		values := make([]luaValue, len(assign.Rhs))
		for i, expr := range assign.Rhs {
			value, err := evalLuaDataExpr(expr, globals)
			if err != nil {
				return err
			}
			values[i] = value
		}
		for i, target := range assign.Lhs {
			switch target := target.(type) {
			case *ast.IdentExpr:
				globals[target.Value] = values[i]
			case *ast.AttrGetExpr:
				object, err := evalLuaDataExpr(target.Object, globals)
				if err != nil {
					return err
				}
				key, err := evalLuaDataExpr(target.Key, globals)
				if err != nil {
					return err
				}
				if object.kind != luaTable {
					return fmt.Errorf("line %d: assignment into a non-table", target.Line())
				}
				if k := luaKey(key); k != nil {
					object.table[k] = values[i]
				}
			default:
				return errLuaNotData
			}
		}
	}
	return nil
}

func evalLuaDataExpr(expr ast.Expr, globals map[string]luaValue) (luaValue, error) {
	switch expr := expr.(type) {
	case *ast.StringExpr:
		return luaValue{kind: luaString, str: expr.Value}, nil
	case *ast.NumberExpr:
		num, err := parseLuaNumber(expr.Value)
		if err != nil {
			return luaValue{}, fmt.Errorf("line %d: %w", expr.Line(), err)
		}
		return luaValue{kind: luaNumber, num: num}, nil
	case *ast.TrueExpr:
		return luaValue{kind: luaBool, num: 1}, nil
	case *ast.FalseExpr:
		return luaValue{kind: luaBool}, nil
	case *ast.NilExpr:
		return luaValue{}, nil
	case *ast.UnaryMinusOpExpr:
		value, err := evalLuaDataExpr(expr.Expr, globals)
		if err != nil {
			return luaValue{}, err
		}
		if value.kind != luaNumber {
			return luaValue{}, errLuaNotData
		}
		value.num = -value.num
		return value, nil
	case *ast.IdentExpr:
		return globals[expr.Value], nil
	case *ast.AttrGetExpr:
		object, err := evalLuaDataExpr(expr.Object, globals)
		if err != nil {
			return luaValue{}, err
		}
		if object.kind != luaTable {
			return luaValue{}, fmt.Errorf("line %d: index of a non-table value", expr.Line())
		}
		key, err := evalLuaDataExpr(expr.Key, globals)
		if err != nil {
			return luaValue{}, err
		}
		if k := luaKey(key); k != nil {
			return object.table[k], nil
		}
		return luaValue{}, nil
	case *ast.TableExpr:
		table := luaValue{kind: luaTable, table: make(map[interface{}]luaValue, len(expr.Fields))}
		next := 1
		for _, field := range expr.Fields {
			value, err := evalLuaDataExpr(field.Value, globals)
			if err != nil {
				return luaValue{}, err
			}
			if field.Key == nil {
				if value.kind != luaNil {
					table.table[next] = value
				}
				next++
				continue
			}
			key, err := evalLuaDataExpr(field.Key, globals)
			if err != nil {
				return luaValue{}, err
			}
			if k := luaKey(key); k != nil && value.kind != luaNil {
				table.table[k] = value
			}
		}
		return table, nil
	default:
		return luaValue{}, errLuaNotData
	}
}

func parseLuaNumber(text string) (float64, error) {
	text = strings.TrimSpace(text)
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "0x") {
		n, err := strconv.ParseUint(lower[2:], 16, 64)
		return float64(n), err
	}
	return strconv.ParseFloat(text, 64)
}
