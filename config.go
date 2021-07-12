package analyzer

import (
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/utils"
	"reflect"
)

const (
	START = iota
	CLOSE
	ERROR
	OK
)

var (
	LState  *lua.LState
	luaFunc = make([]*hook, 0)
	ANA     = reflect.TypeOf((*Analyzer)(nil)).String()
)

type hook struct {
	filename string
	fn       *lua.LFunction
}

type config struct {
	name      string `lua:"name,null" type:"string"`
	thread    int
	input     Input  // 输入接口
	script    string // lua脚本路径
	heartbeat int
}

// Input 接口获取输入来源的通道,通常为kafka消费者，es搜索结果等，这些模块需要实现该接口
type Input interface {
	GetBuffer() *chan []byte
	GetName() string
}

func newConfig(L *lua.LState) *config {
	tab := L.CheckTable(1)
	cfg := &config{}

	tab.ForEach(func(key lua.LValue, val lua.LValue) {
		switch key.String() {
		case "name":
			cfg.name = utils.CheckProcName(val, L)
		case "thread":
			cfg.thread = utils.LValueToInt(val, 1)
		case "input":
			cfg.input = checkInput(L, val)
		case "script":
			cfg.script = utils.LValueToStr(val, "resource/script/")
		case "heartbeat":
			cfg.heartbeat = utils.LValueToInt(val, 10)
		}
	})

	if cfg.input == nil {
		L.RaiseError("%s analyzer input is nil", cfg.name)
		return nil
	}
	return cfg
}

func checkInput(L *lua.LState, val lua.LValue) Input {
	data := lua.CheckLightUserData(L, val)
	if input, ok := data.Value.(interface{}).(Input); ok {
		return input
	}

	return nil
}
