package analyzer

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"path/filepath"
	"time"
)

type Analyzer struct {
	lua.Super

	cfg config

	input  *chan []byte // 数据来源
	thread []*Thread

	received uint64 // 接收到的数据量
	parsed   uint64 // 正确处理的数据量

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
	a.thread = make([]*Thread, a.cfg.thread)
	a.ctx, a.cancel = context.WithCancel(context.Background())

	for i := 0; i < a.cfg.thread; i++ {
		a.thread[i] = NewThread(i, a)
		go a.thread[i].Start()
		time.Sleep(50 * time.Millisecond)
	}

	go a.Heartbeat()
	go a.SyncRule()

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
			logger.Debugf("%s analyzer thread.id=%d running", a.cfg.name, id)
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
			logger.Debugf("%s analyzer received %d, parsed %d", a.cfg.name, a.received, a.received)
		}
	}
}

// SyncRule 监控lua规则文件，有改动则更新
func (a *Analyzer) SyncRule() {
	co := lua.NewState()
	defer co.Close()

	if err := compilePath(co, a.cfg.script); err != nil {
		logger.Errorf("compile lua file path %s error: %v", a.cfg.script, err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf("new fs notify watcher error: %v", err)
		return
	}
	defer watcher.Close()

	dir := filepath.Dir(a.cfg.script)
	err = watcher.Add(dir)
	if err != nil {
		logger.Errorf("add directory to fs notify watcher error: %v", err)
		return
	}

	logger.Infof("fs notify watcher directory: %s", a.cfg.script)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			logger.Errorf("directory %s was changed: %s", a.cfg.script, event.String())
			luaFunc = make([]*hook, 0)
			if err = compilePath(co, a.cfg.script); err != nil {
				logger.Errorf("compile lua file path %s error: %v", a.cfg.script, err)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Errorf("directory watcher error: %v", err)
		}
	}
}

func (a *Analyzer) State() lua.LightUserDataStatus {
	if a.thread == nil {
		return lua.CLOSE
	}

	inactive := 0
	for _, v := range a.thread {
		if v.status != OK {
			inactive++
		}
	}

	if inactive == a.cfg.thread {
		return lua.CLOSE
	}

	return lua.RUNNING
}

func (a *Analyzer) Type() string {
	return ANA
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
