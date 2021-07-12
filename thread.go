package analyzer

import (
	"context"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
	"sync"
)

// Thread 单个线程
type Thread struct {
	c config

	id     int
	input  *chan []byte
	ctx    context.Context
	status int
}

func NewThread(id int, a *Analyzer) Thread {
	thread := Thread{
		c:  a.cfg,
		id: id,

		ctx:    a.ctx,
		status: START,

		input: a.input,
	}

	return thread
}

func (t *Thread) Start() error {
	t.status = OK
	logger.Errorf("%s analyzer thread.id = %d start ok", t.c.name, t.id)

	t.Handler(t.ctx)

	return nil
}

//sync.pool
// Handler 从输入读取数据并进行处理
var luaThreadPool sync.Pool

func newThread() *lua.LState {
	th := luaThreadPool.Get()
	if th == nil {
		co, _ := LState.NewThread()
		return co
	}

	return th.(*lua.LState)
}

func freeThread(co *lua.LState) {
	co.A = nil
	luaThreadPool.Put(co)
}

func (t *Thread) Handler(ctx context.Context) {
	co := lua.State()
	defer co.Close()

	for {
		select {
		case <-ctx.Done():
			logger.Errorf("%s analyzer thread.id = %d start ok", t.c.name, t.id)
		case data, ok := <-*t.input:
			if ok {
				luaAnalyze(co, data)
			}
		}
	}

}

// 执行lua分析脚本
func luaAnalyze(co *lua.LState, msg []byte) {

	co.A = msg
	for _, h := range luaFunc {
		err := xcall.CallByEnv(co, h.fn, xcall.Rock)
		if err != nil {
			logger.Errorf("execute lua script error: %v", err)
			break
		}
	}
	co.A = nil
}
