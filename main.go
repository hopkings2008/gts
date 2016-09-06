package main

import (
	//"fmt"

	"github.com/zouyu/gts/proxy"
)

func main() {
	m := make(map[string]*proxy.TargetInfo)
	m["weichat"] = &proxy.TargetInfo{
		Target:  "http://mmsns.qpic.cn",
		MaxConn: 0,
		MaxRps:  60,
	}
	p, _ := proxy.NewGtsProxy(m)
	p.Run()
}
