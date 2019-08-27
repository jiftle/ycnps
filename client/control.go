package client

import (
	//"bufio"
	"ycnps/common"
	"ycnps/mux"
	//	"github.com/cnlh/nps/lib/config"
	//	"github.com/cnlh/nps/lib/conn"
	//	"github.com/cnlh/nps/lib/crypt"
	//	"github.com/cnlh/nps/vender/github.com/astaxie/beego/logs"
	//	"github.com/cnlh/nps/vender/github.com/xtaci/kcp"
	//"net"
	// "net/http"
	logger "github.com/ccpaging/log4go"
	//	"strconv"
	"net"
	"time"
	"ycnps/config"
	"ycnps/conn"
)

type TRPClient struct {
	svrAddr        string
	bridgeConnType string
	proxyUrl       string
	vKey           string
	p2pAddr        map[string]string
	tunnel         *mux.Mux
	signal         *conn.Conn
	ticker         *time.Ticker
	cnf            *config.Config
}

//new client　创建一个结构体
func NewRPClient(svraddr string, vKey string, bridgeConnType string, proxyUrl string, cnf *config.Config) *TRPClient {
	return &TRPClient{
		svrAddr:        svraddr, //地址
		p2pAddr:        make(map[string]string, 0),
		vKey:           vKey,           //客户端
		bridgeConnType: bridgeConnType, //桥接类型
		proxyUrl:       proxyUrl,       //代理地址
		cnf:            cnf,            //配置
	}
}

//start 启动
func (s *TRPClient) Start() {
retry:
	// 连接服务器
	c, err := NewConn(s.bridgeConnType, s.vKey, s.svrAddr, common.WORK_MAIN, s.proxyUrl)
	if err != nil {
		logger.Error("连接服务器失败,%s. 5秒后重连.", err)
		time.Sleep(time.Second * 5)
		goto retry
	}
	logger.Info("连接成功,server %s", s.svrAddr)

	//monitor the connection 监视连接状态
	go s.ping()

	s.signal = c

	//start a channel connection 开始一个通道协程,侦听服务端连接
	go s.newChan()

	//start health check if the it's open 如果打开开启健康检查
	if s.cnf != nil && len(s.cnf.Healths) > 0 {
		// go heathCheck(s.cnf.Healths, s.signal)
	}

	//msg connection, eg udp
	//	s.handleMain()
}

// //handle main connection
// func (s *TRPClient) handleMain() {
// 	for {
// 		flags, err := s.signal.ReadFlag()
// 		if err != nil {
// 			logger.Error("Accept server data error %s, end this service", err.Error())
// 			break
// 		}
// 		switch flags {
// 		case common.NEW_UDP_CONN:
// 			//read server udp addr and password
// 			if lAddr, err := s.signal.GetShortLenContent(); err != nil {
// 				logger.Warn(err)
// 				return
// 			} else if pwd, err := s.signal.GetShortLenContent(); err == nil {
// 				var localAddr string
// 				//The local port remains unchanged for a certain period of time
// 				if v, ok := s.p2pAddr[crypt.Md5(string(pwd)+strconv.Itoa(int(time.Now().Unix()/100)))]; !ok {
// 					tmpConn, err := common.GetLocalUdpAddr()
// 					if err != nil {
// 						logs.Error(err)
// 						return
// 					}
// 					localAddr = tmpConn.LocalAddr().String()
// 				} else {
// 					localAddr = v
// 				}
// 				go s.newUdpConn(localAddr, string(lAddr), string(pwd))
// 			}
// 		}
// 	}
// 	s.Close()
// }

// Whether the monitor channel is closed
func (s *TRPClient) ping() {
	s.ticker = time.NewTicker(time.Second * 5)
loop:
	for {
		select {
		case <-s.ticker.C:
			if s.tunnel != nil && s.tunnel.IsClose {
				s.Close()
				break loop
			}
		} // select
	} //for
}

//mux tunnel
func (s *TRPClient) newChan() {
	tunnel, err := NewConn(s.bridgeConnType, s.vKey, s.svrAddr, common.WORK_CHAN, s.proxyUrl)
	if err != nil {
		logger.Error("connect to ", s.svrAddr, "error:", err)
		return
	}
	logger.Info("新建Chan，%v", tunnel)
	s.tunnel = mux.NewMux(tunnel.Conn, s.bridgeConnType)
	for {
		// 接收客户端连接
		src, err := s.tunnel.Accept()
		if err != nil {
			logger.Warn("接收客户端连接失败,%s", err)
			s.Close()
			break
		}
		go s.handleChan(src)
	}
}

func (s *TRPClient) handleChan(src net.Conn) {
	lk, err := conn.NewConn(src).GetLinkInfo()
	if err != nil {
		src.Close()
		logger.Error("get connection info from server error ", err)
		return
	}
	logger.Info("来自服务端的新建连接信息,lk=%v\n", lk)
	//	//host for target processing
	//	lk.Host = common.FormatAddress(lk.Host)
	//	//if Conn type is http, read the request and log
	//	if lk.ConnType == "http" {
	//		if targetConn, err := net.Dial(common.CONN_TCP, lk.Host); err != nil {
	//			logger.Warn("connect to %s error %s", lk.Host, err.Error())
	//			src.Close()
	//		} else {
	//			srcConn := conn.GetConn(src, lk.Crypt, lk.Compress, nil, false)
	//			go func() {
	//				common.CopyBuffer(srcConn, targetConn)
	//				srcConn.Close()
	//				targetConn.Close()
	//			}()
	//			for {
	//				if r, err := http.ReadRequest(bufio.NewReader(srcConn)); err != nil {
	//					srcConn.Close()
	//					targetConn.Close()
	//					break
	//				} else {
	//					logger.Trace("http request, method %s, host %s, url %s, remote address %s", r.Method, r.Host, r.URL.Path, r.RemoteAddr)
	//					r.Write(targetConn)
	//				}
	//			}
	//		}
	//		return
	// 	}
	// 	//connect to target if conn type is tcp or udp
	// 	if targetConn, err := net.Dial(lk.ConnType, lk.Host); err != nil {
	// 		logger.Warn("connect to %s error %s", lk.Host, err.Error())
	// 		src.Close()
	// 	} else {
	// 		logger.Trace("new %s connection with the goal of %s, remote address:%s", lk.ConnType, lk.Host, lk.RemoteAddr)
	// 		conn.CopyWaitGroup(src, targetConn, lk.Crypt, lk.Compress, nil, nil, false, nil)
	// 	}
}

// 关闭
func (s *TRPClient) Close() {
	if s.tunnel != nil {
		s.tunnel.Close()
	}
	if s.signal != nil {
		s.signal.Close()
	}
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
