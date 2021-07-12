package analyzer

import (
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
)

func LuaInjectApi(env xcall.Env) {
	env.Set("log_analyzer", lua.NewFunction(createAnalyzerUserData))

	analyzer := lua.NewUserKV()
	analyzer.Set("callback", lua.NewFunction(newCallbackChains))
	analyzer.Set("parser", lua.NewAnyData(newLuaParser()))
	env.Set("analyzer", analyzer)
}

func (a *Analyzer) Index(L *lua.LState, key string) lua.LValue {
	if key == "start" {
		return lua.NewFunction(a.start)
	}
	if key == "close" {
		return lua.NewFunction(a.close)
	}

	return lua.LNil
}

func (a *Analyzer) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "name":
		a.cfg.name = lua.CheckString(L, val)
	case "thread":
		a.cfg.thread = lua.CheckInt(L, val)
	case "input":
		a.cfg.input = checkInput(L, val)
	case "script":
		a.cfg.script = lua.CheckString(L, val)
	case "heartbeat":
		a.cfg.heartbeat = lua.CheckInt(L, val)
	}
}

func createAnalyzerUserData(L *lua.LState) int {

	cfg := newConfig(L)
	proc := L.NewProc(cfg.name, ANA)
	if proc.IsNil() {
		proc.Set(newAnalyzer(cfg))
	} else {
		proc.Value.(*Analyzer).cfg = *cfg
	}
	L.Push(proc)
	return 1
}

func newCallbackChains(L *lua.LState) int {
	filename := L.B.(string)
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	for i := 1; i <= n; i++ {
		fn := L.CheckFunction(i)
		luaFunc = append(luaFunc, &hook{
			filename: filename,
			fn:       fn,
		})
	}

	return 0
}

func (a *Analyzer) start(L *lua.LState) int {
	err := a.Start()
	if err != nil {
		L.RaiseError("%s start error: %v", a.cfg.name, err)
	}

	return 0
}

func (a *Analyzer) close(L *lua.LState) int {
	err := a.Close()
	if err != nil {
		L.RaiseError("%s close error: %v", a.cfg.name, err)
	}

	return 0
}
