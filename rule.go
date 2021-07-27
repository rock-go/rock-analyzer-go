package analyzer

import (
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
	"io/fs"
	"io/ioutil"
	"os"
)

// 解析和同步lua脚本

// 解析单个文件
func compileLua(co *lua.LState, filepath string) error {
	co.B = filepath
	err := xcall.DoFileByEnv(co, filepath, xcall.Rock)
	if err != nil {
		return err
	}

	logger.Infof("compile lua file %s succeed", filepath)
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

	logger.Infof("compile lua file path %s succeed", path)

	return nil
}
