package analyzer

import (
	"github.com/rock-go/rock/lua"
	"github.com/valyala/fastjson"
	"strings"
)

// Parser lua脚本里用来解析一条message
type Parser struct {
	lua.NoReflect
	lua.Super

	codec      string            // 消息类型
	chunkSlice [][]byte          // 存放字符串分割后的结果
	chunkMap   map[string][]byte // 存放json解析结果
	meta       lua.UserKV
}

func newLuaParser(L *lua.LState) int {
	p := &Parser{}
	p.meta = lua.NewUserKV()
	L.Push(lua.NewLightUserData(p))
	return 1
}

func (p *Parser) Index(L *lua.LState, key string) lua.LValue {
	lv := p.meta.Get(key)
	if lv != lua.LNil {
		return lv
	}

	switch key {
	case "contain":
		lv = L.NewFunction(p.LContain)
	case "split":
		lv = L.NewFunction(p.LSplit)
	case "parse_json":
		lv = L.NewFunction(p.LParseJson)
	case "json_get":
		lv = L.NewFunction(p.LGetJson)
	case "slice_get":
		lv = L.NewFunction(p.LGetSlice)
	case "msg":
		lv = L.NewFunction(p.LMsg)
	case "b2s":
		lv = L.NewFunction(p.LB2S)
	default:
		return lua.LNil
	}

	p.meta.Set(key, lv)
	return lv
}

// 通过fastjson 解析
func (p *Parser) parseJson(data []byte, pattern []string) error {
	var fjp fastjson.Parser

	v, err := fjp.ParseBytes(data)
	if err != nil {
		return err
	}

	// e.g.: []string{"a","a.b","a.b.cfg"}
	for _, fieldPath := range pattern {
		var fields []string

		if !strings.Contains(fieldPath, ".") {
			fields = []string{fieldPath}
		} else {
			fields = strings.Split(fieldPath, ".")
		}

		if fields == nil {
			continue
		}

		var obj = v
		var dst []byte
		for _, f := range fields {
			if obj == nil {
				break
			}
			obj = obj.Get(f)
		}

		if obj == nil {
			continue
		}

		p.chunkMap[fieldPath] = obj.MarshalTo(dst)
	}

	return nil
}
