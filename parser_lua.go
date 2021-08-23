package analyzer

import (
	"bytes"
	"github.com/rock-go/rock/lua"
)

// LMsg 获取原始的待处理数据
func (p *Parser) LMsg(co *lua.LState) int {
	if co.A == nil {
		return 0
	}

	co.Push(lua.B2L(co.A.([]byte)))
	return 1
}

// LContain 判断是否包含某个字符串,args 一般为两个参数，第一个为包含原始字符串，第二个为子字符串
func (p *Parser) LContain(co *lua.LState) int {
	if co.GetTop() < 2 {
		co.Push(lua.LTrue)
		return 1
	}

	exist := bytes.Contains(lua.S2B(co.CheckString(1)), []byte(co.CheckString(2)))
	co.Push(lua.LBool(exist))
	return 1
}

// LSplit 分割字符串,args 一般为两个参数，第一个为包含原始字符串，第二个为分隔符
// 返回Parse对象
func (p *Parser) LSplit(co *lua.LState) int {
	n := co.GetTop()
	if n != 2 {
		co.RaiseError("function split must have 2 args, get %d", n)
		return 0
	}

	sep := []byte(co.CheckString(2))
	byteSlice := bytes.Split(lua.S2B(co.CheckString(1)), sep)

	parse := &Parser{chunkSlice: byteSlice}
	co.Push(&lua.LUserData{Value: parse})
	return 1
}

// LParseJson 解析json，args多个参数，第一个为待处理的字符串，其他为待获取的json字段路径
// 返回Parse对象
func (p *Parser) LParseJson(co *lua.LState) int {
	n := co.GetTop()
	if n < 1 {
		co.RaiseError("json parse need at least 1 arg, get %d", n)
		return 0
	}

	pattern := make([]string, n-1)
	for i := 1; i < n; i++ {
		pattern[i-1] = co.CheckString(i + 1)
	}

	parse := &Parser{chunkMap: make(map[string][]byte)}
	if err := parse.parseJson(lua.S2B(co.CheckString(1)), pattern); err != nil {
		co.RaiseError("json parse error: %v", err)
		return 0
	}

	co.Push(&lua.LUserData{Value: parse})
	return 1
}

// LGetJson 获取Parse对象的数据，第一个参数是Value为Parse的userdata，第二个为chunkMap的key
// 返回 string
func (p *Parser) LGetJson(co *lua.LState) int {
	n := co.GetTop()
	if n < 2 {
		return 0
	}

	parse, ok := co.CheckUserData(1).Value.(*Parser)
	if !ok {
		co.RaiseError("the object of parse get function must be *Parser")
		return 0
	}

	res, ok := parse.chunkMap[co.CheckString(2)]
	if !ok {
		return 0
	}
	co.Push(lua.B2L(res))
	return 1

}

// LGetSlice 获取Parse对象的数据，第一个参数是Value为Parse的userdata，第二个为chunkSlice的index
// 返回 string
func (p Parser) LGetSlice(co *lua.LState) int {
	n := co.GetTop()
	if n < 2 {
		return 0
	}

	parse, ok := co.CheckUserData(1).Value.(*Parser)
	if !ok {
		co.RaiseError("the object of parse get function must be *Parser")
		return 0
	}

	index := co.CheckInt(2)
	max := len(parse.chunkSlice) - 1
	if index > max {
		co.RaiseError("index out of bounds, max: %d", max)
		return 0
	}

	res := parse.chunkSlice[index]
	co.Push(lua.B2L(res))
	return 1
}

func (p *Parser) LB2S(co *lua.LState) int {
	ud := co.CheckUserData(1)
	if data, ok := ud.Value.([]byte); ok {
		co.Push(lua.LString(data))
		return 1
	}

	co.RaiseError("the value of userdata must be []byte")
	return 0
}
