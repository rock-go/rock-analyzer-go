package analyzer

import (
	"context"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
	"sync"
	"time"
)

// Thread 单个线程
type Thread struct {
	c config

	id    int
	input *chan []byte

	co *lua.LState

	ctx    context.Context
	status int
}

func NewThread(id int, a *Analyzer) Thread {
	thread := Thread{
		c:  a.cfg,
		id: id,

		co: lua.NewState(),

		ctx:    a.ctx,
		status: START,

		input: a.input,
	}

	return thread
}

func (t *Thread) Start() error {
	t.status = OK
	logger.Errorf("%s analyzer thread.id = %d start ok", t.c.name, t.id)

	time.Sleep(500 * time.Millisecond)
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
	for {
		select {
		case <-ctx.Done():
			t.co.Close()
			t.status = CLOSE
			logger.Errorf("%s analyzer thread.id = %d close ok", t.c.name, t.id)
			return
		case data, ok := <-*t.input:
			if ok {
				luaAnalyze(t.co, data)
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
