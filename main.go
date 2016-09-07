package main

import (
	//"fmt"

	"github.com/zouyu/gts/proxy"
)

func main() {
	m := make(map[string]*proxy.TargetInfo)
	m["weichat"] = &proxy.TargetInfo{
		Target:  "http://mmsns.qpic.cn",
		MaxConn: 10,
		MaxRps:  80,
	}
	p, _ := proxy.NewGtsProxy(m)
	p.AddIps("60.205.127.147", "101.201.53.131", "127.0.0.1")
	p.Run()
}
