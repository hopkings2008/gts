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
	m["qzoneuser"] = &proxy.TargetInfo{
		Target:  "http://users.qzone.qq.com",
		MaxConn: 10,
		MaxRps:  80,
	}
	m["qzoneb1"] = &proxy.TargetInfo{
		Target:  "http://b1.qzone.qq.com",
		MaxConn: 10,
		MaxRps:  80,
	}
	m["qzoneb11"] = &proxy.TargetInfo{
		Target:  "http://b11.qzone.qq.com",
		MaxConn: 10,
		MaxRps:  80,
	}
	m["qqtaotao"] = &proxy.TargetInfo{
		Target:  "http://taotao.qq.com",
		MaxConn: 10,
		MaxRps:  80,
	}
	p, _ := proxy.NewGtsProxy(m)
	p.AddIps("60.205.127.147", "101.201.53.131", "101.201.81.33", "123.57.62.42", "115.183.72.43", "111.198.71.191", "101.200.200.179", "127.0.0.1")
	p.Run()
}
