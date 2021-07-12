package analyzer

import (
	"context"
	"fmt"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"time"
)

type Analyzer struct {
	lua.Super

	cfg config

	input  *chan []byte // 数据来源
	thread []Thread
	ctx    context.Context
	cancel context.CancelFunc
}

func newAnalyzer(cfg *config) *Analyzer {
	a := &Analyzer{cfg: *cfg}
	a.S = lua.INIT
	a.T = ANA
	return a
}
func (a *Analyzer) Start() error {
	a.input = a.cfg.input.GetBuffer()
	a.thread = make([]Thread, a.cfg.thread)
	a.ctx, a.cancel = context.WithCancel(context.Background())

	go a.SyncRule()

	for i := 0; i < a.cfg.thread; i++ {
		a.thread[i] = NewThread(i, a)
		go a.thread[i].Start()
		time.Sleep(50 * time.Millisecond)
	}

	go a.Heartbeat()

	a.U = time.Now()
	a.S = lua.RUNNING

	logger.Infof("%s analyzer start successfully", a.cfg.name)

	return nil
}

func (a *Analyzer) Close() error {
	a.S = lua.CLOSE
	if a.cancel != nil {
		a.cancel()
		return nil
	}

	logger.Infof("%s analyzer close successfully", a.cfg.name)
	return nil
}

func (a *Analyzer) Ping() {
	for id, t := range a.thread {
		switch t.status {
		case OK:
			continue
		case CLOSE:
			logger.Errorf("%s analyzer thread.id = %d close", a.cfg.name, id)
		case ERROR:
			go a.thread[id].Start()
		}
	}
}

// Heartbeat 心跳检测
func (a *Analyzer) Heartbeat() {
	tk := time.NewTicker(time.Second * time.Duration(a.cfg.heartbeat))
	defer tk.Stop()

	for {
		select {
		case <-a.ctx.Done():
			logger.Errorf("%s analyzer heartbeat exit", a.cfg.name)
			return
		case <-tk.C:
			a.Ping()
		}
	}
}

func (a *Analyzer) Type() string {
	return "log analyzer"
}

func (a *Analyzer) Name() string {
	return a.cfg.name
}

func (a *Analyzer) Status() string {
	var scripts string
	for _, s := range luaFunc {
		scripts = scripts + " " + s.filename
	}

	return fmt.Sprintf("name: %s, status: %s, uptime: %s, rule scripts: %s",
		a.cfg.name, a.S.String(), a.U.Format("2006-01-02 15:04:06"), scripts)
}
