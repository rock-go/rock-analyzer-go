package analyzer

import "github.com/rock-go/rock/lua"

func (a *Analyzer) Header(out lua.Printer) {
	out.Printf("type: %s", a.Type())
	out.Printf("uptime: %s", a.U.Format("2006-01-02 15:04:06"))
	out.Println("version: v1.0.0")
	out.Println("")
}

func (a *Analyzer) Show(out lua.Printer) {
	a.Header(out)

	out.Printf("name: %s", a.cfg.name)
	out.Printf("thread: %d", a.cfg.thread)
	out.Printf("input: %s", a.cfg.input.GetName())
	out.Printf("script: %s", a.cfg.script)
	out.Printf("heartbeat: %d", a.cfg.heartbeat)
}

func (a *Analyzer) Help(out lua.Printer) {
	a.Header(out)

	out.Printf(".start() 启动")
	out.Printf(".close() 关闭")
	out.Println("")
}
