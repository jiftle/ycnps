package main

import (
	"flag"
	"ycnps/client"

	logger "github.com/ccpaging/log4go"
)

var (
	serverAddr   = flag.String("server", "", "Server addr (ip:port)")
	configPath   = flag.String("config", "", "Configuration file path")
	verifyKey    = flag.String("vkey", "", "Authentication key")
	logType      = flag.String("log", "stdout", "Log output mode（stdout|file）")
	connType     = flag.String("type", "tcp", "Connection type with the server（kcp|tcp）")
	proxyUrl     = flag.String("proxy", "", "proxy socks5 url(eg:socks5://111:222@127.0.0.1:9007)")
	logLevel     = flag.String("log_level", "7", "log level 0~7")
	registerTime = flag.Int("time", 2, "register time long /h")
	localPort    = flag.Int("local_port", 2000, "p2p local port")
	password     = flag.String("password", "", "p2p password flag")
	target       = flag.String("target", "", "p2p target")
	localType    = flag.String("local_type", "p2p", "p2p target")
	logPath      = flag.String("log_path", "npc.log", "npc log path")
)

func main() {

	defer logger.Close()

	// 解析命令行参数
	flag.Parse()

	// 指定默认的配置文件路径
	*configPath = "conf/npc.conf"

	// 从配置文件启动
	client.StartFromFile(*configPath)
}

func init() {
	logger.LoadConfiguration("conf/log4go.xml")
}
