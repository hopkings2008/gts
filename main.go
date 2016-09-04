package main

import (
	//"fmt"

	"github.com/zouyu/gts/proxy"
)

func main() {
	m := make(map[string]string)
	m["weichat"] = "http://mmsns.qpic.cn"
	p, _ := proxy.NewGtsProxy(m)
	p.Run()
}
