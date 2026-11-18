package res

import (
	"bytes"
	"errors"

	glua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// isLua51Bytecode reports whether data is a precompiled Lua 5.1 chunk, as
// most kRO .lub files are. Everything else is treated as Lua source.
func isLua51Bytecode(data []byte) bool {
	return bytes.HasPrefix(data, []byte("\x1bLua"))
}

// executeLuaChunk runs a client Lua table file whether it is compiled Lua 5.1
// bytecode (translated in process to gopher-lua, no decompiler) or plain Lua
// source, such as the translated tables Turoran ships as loose overrides.
// Globals are seeded from and written back to globals.
func executeLuaChunk(data []byte, globals map[string]luaValue) error {
	if isLua51Bytecode(data) {
		return executeLua51Bytecode(data, globals)
	}
	return executeLuaSource(data, globals, nil)
}

// executeLuaChunkWithSetup is executeLuaChunk with host functions installed
// by setup before the chunk runs.
func executeLuaChunkWithSetup(data []byte, globals map[string]luaValue, setup func(*glua.LState)) error {
	if isLua51Bytecode(data) {
		return executeLua51BytecodeWithSetup(data, globals, setup)
	}
	return executeLuaSource(data, globals, setup)
}

func executeLuaSource(data []byte, globals map[string]luaValue, setup func(*glua.LState)) error {
	if globals == nil {
		return errors.New("nil Lua globals")
	}
	source := bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	// The lexer is byte based, so CP949 string literals survive unchanged.
	chunk, err := parse.Parse(bytes.NewReader(source), "lua")
	if err != nil {
		return err
	}
	// Pure data tables skip the VM entirely (see evalLuaDataChunk). Work on a
	// copy so a chunk that turns out to need the VM starts from clean globals.
	scratch := make(map[string]luaValue, len(globals))
	for key, value := range globals {
		scratch[key] = value
	}
	if err := evalLuaDataChunk(chunk, scratch); err == nil {
		for key, value := range scratch {
			globals[key] = value
		}
		return nil
	}
	state := glua.NewState(glua.Options{SkipOpenLibs: true})
	defer state.Close()
	openLuaResourceLibs(state)
	if setup != nil {
		setup(state)
	}
	for key, value := range globals {
		if value.kind != luaNil {
			state.SetGlobal(key, luaValueToGopherLua(state, value))
		}
	}
	proto, err := glua.Compile(chunk, "lua")
	if err != nil {
		return err
	}
	fn := state.NewFunctionFromProto(proto)
	if err := luaCallByParam(state, fn); err != nil {
		return err
	}
	state.G.Global.ForEach(func(key glua.LValue, value glua.LValue) {
		name, ok := key.(glua.LString)
		if !ok || luaBuiltinGlobal(string(name)) {
			return
		}
		if converted, ok := gopherLuaToLuaValue(value); ok {
			globals[string(name)] = converted
		}
	})
	return nil
}

// luaField returns table[key] for a string key.
func luaField(table luaValue, key string) luaValue {
	if table.kind != luaTable {
		return luaValue{}
	}
	return table.table[key]
}

// luaStringList returns the string values of an array-like table in order.
func luaStringList(table luaValue) []string {
	if table.kind != luaTable {
		return nil
	}
	var out []string
	for i := 1; ; i++ {
		value, ok := table.table[i]
		if !ok {
			break
		}
		if value.kind == luaString {
			out = append(out, value.str)
		}
	}
	return out
}
