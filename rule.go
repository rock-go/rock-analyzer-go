package analyzer

import (
	"github.com/fsnotify/fsnotify"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// 解析和同步lua脚本

// 解析单个文件
func compileLua(co *lua.LState, filepath string) error {
	co.B = filepath
	err := xcall.DoFileByEnv(co, filepath, xcall.Rock)
	if err != nil {
		return err
	}

	return nil
}

// 通过路径解析lua文件
func compilePath(co *lua.LState, path string) error {
	var err error
	var fileInfo fs.FileInfo
	var filesInfo []fs.FileInfo

	fileInfo, err = os.Stat(path)
	if err != nil {
		return err
	}

	// 如果是文件
	if !fileInfo.IsDir() {
		return compileLua(co, path)
	}

	filesInfo, err = ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range filesInfo {
		name := f.Name()
		filePath := path + "/" + name
		err = compilePath(co, filePath)
		if err != nil {
			logger.Errorf("compile lua file %s error: %v", filePath, err)
		}
	}

	return nil
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
