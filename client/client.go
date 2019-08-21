package client

import (
	logger "github.com/ccpaging/log4go"
	"ycnps/common"
	"ycnps/config"
	"ycnps/conn"
	"ycnps/crypt"
	"ycnps/version"
	//	"encoding/base64"
	"encoding/binary"
	//	"errors"
	//	"fmt"
	//	"github.com/cnlh/nps/lib/common"
	//	"github.com/cnlh/nps/lib/config"
	//	"github.com/cnlh/nps/lib/conn"
	//	"github.com/cnlh/nps/lib/crypt"
	//	"github.com/cnlh/nps/lib/version"
	//	"github.com/cnlh/nps/vender/github.com/astaxie/beego/logs"
	//	"github.com/cnlh/nps/vender/github.com/xtaci/kcp"
	//	"github.com/cnlh/nps/vender/golang.org/x/net/proxy"
	"io/ioutil"
	//	"log"
	//	"math"
	//	"math/rand"
	"net"
	//	"net/http"
	//	"net/http/httputil"
	//	"net/url"
	"os"
	"path/filepath"
	//	"strconv"
	//	"strings"
	"time"
)

// func GetTaskStatus(path string) {
// 	cnf, err := config.NewConfig(path)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	c, err := NewConn(cnf.CommonConfig.Tp, cnf.CommonConfig.VKey, cnf.CommonConfig.Server, common.WORK_CONFIG, cnf.CommonConfig.ProxyUrl)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	if _, err := c.Write([]byte(common.WORK_STATUS)); err != nil {
// 		log.Fatalln(err)
// 	}
// 	//read now vKey and write to server
// 	if f, err := common.ReadAllFromFile(filepath.Join(common.GetTmpPath(), "npc_vkey.txt")); err != nil {
// 		log.Fatalln(err)
// 	} else if _, err := c.Write([]byte(crypt.Md5(string(f)))); err != nil {
// 		log.Fatalln(err)
// 	}
// 	var isPub bool
// 	binary.Read(c, binary.LittleEndian, &isPub)
// 	if l, err := c.GetLen(); err != nil {
// 		log.Fatalln(err)
// 	} else if b, err := c.GetShortContent(l); err != nil {
// 		log.Fatalln(err)
// 	} else {
// 		arr := strings.Split(string(b), common.CONN_DATA_SEQ)
// 		for _, v := range cnf.Hosts {
// 			if common.InStrArr(arr, v.Remark) {
// 				log.Println(v.Remark, "ok")
// 			} else {
// 				log.Println(v.Remark, "not running")
// 			}
// 		}
// 		for _, v := range cnf.Tasks {
// 			ports := common.GetPorts(v.Ports)
// 			if v.Mode == "secret" {
// 				ports = append(ports, 0)
// 			}
// 			for _, vv := range ports {
// 				var remark string
// 				if len(ports) > 1 {
// 					remark = v.Remark + "_" + strconv.Itoa(vv)
// 				} else {
// 					remark = v.Remark
// 				}
// 				if common.InStrArr(arr, remark) {
// 					log.Println(remark, "ok")
// 				} else {
// 					log.Println(remark, "not running")
// 				}
// 			}
// 		}
// 	}
// 	os.Exit(0)
// }

//var errAdd = errors.New("The server returned an error, which port or host may have been occupied or not allowed to open.")

// 从配置文件启动
func StartFromFile(path string) {
	cnf, err := config.NewConfig(path)
	if err != nil || cnf.CommonConfig == nil {
		logger.Error("配置文件[%s]加载失败,%s", path, err.Error())
		os.Exit(0)
	}
	logger.Info("加载配置文件[%s]成功", path)

	// 新建连接 c
	c, err := NewConn(cnf.CommonConfig.Tp, cnf.CommonConfig.VKey, cnf.CommonConfig.Server, common.WORK_CONFIG, cnf.CommonConfig.ProxyUrl)
	if err != nil {
		logger.Error(err)
	}

	var isPub bool
	binary.Read(c, binary.LittleEndian, &isPub)
	logger.Info("服务端返回isPub=%v", isPub)

	// get tmp password
	var b []byte
	vkey := cnf.CommonConfig.VKey
	if isPub {
		// send global configuration to server and get status of config setting
		if _, err := c.SendInfo(cnf.CommonConfig.Client, common.NEW_CONF); err != nil {
			logger.Error(err)
			return
			//goto re
		}
		if !c.GetAddStatus() {
			logger.Error("the web_user may have been occupied!")
			return
			//goto re
		}

		if b, err = c.GetShortContent(16); err != nil {
			logger.Error(err)
			return
			//goto re
		}
		vkey = string(b)
	}
	sfilepath := filepath.Join(common.GetTmpPath(), "npc_vkey.txt")
	logger.Info("vkey验证密钥写入到文件%s", sfilepath)
	ioutil.WriteFile(sfilepath, []byte(vkey), 0600)

	//	//send hosts to server
	//	for _, v := range cnf.Hosts {
	//		if _, err := c.SendInfo(v, common.NEW_HOST); err != nil {
	//			logs.Error(err)
	//			goto re
	//		}
	//		if !c.GetAddStatus() {
	//			logs.Error(errAdd, v.Host)
	//			goto re
	//		}
	//	}
	//
	//	//send  task to server
	//	for _, v := range cnf.Tasks {
	//		if _, err := c.SendInfo(v, common.NEW_TASK); err != nil {
	//			logs.Error(err)
	//			goto re
	//		}
	//		if !c.GetAddStatus() {
	//			logs.Error(errAdd, v.Ports, v.Remark)
	//			goto re
	//		}
	//		if v.Mode == "file" {
	//			//start local file server
	//			go startLocalFileServer(cnf.CommonConfig, v, vkey)
	//		}
	//	}
	//
	//	//create local server secret or p2p
	//	for _, v := range cnf.LocalServer {
	//		go StartLocalServer(v, cnf.CommonConfig)
	//	}
	//
	//	c.Close()
	//	if cnf.CommonConfig.Client.WebUserName == "" || cnf.CommonConfig.Client.WebPassword == "" {
	//		logs.Notice("web access login username:user password:%s", vkey)
	//	} else {
	//		logs.Notice("web access login username:%s password:%s", cnf.CommonConfig.Client.WebUserName, cnf.CommonConfig.Client.WebPassword)
	//	}
	//	NewRPClient(cnf.CommonConfig.Server, vkey, cnf.CommonConfig.Tp, cnf.CommonConfig.ProxyUrl, cnf).Start()
	//	CloseLocalServer()
	//	goto re
}

// Create a new connection with the server and verify it
// 创建一个新的连接
func NewConn(tp string, vkey string, server string, connType string, proxyUrl string) (*conn.Conn, error) {
	logger.Info("tp:%s,vkey:%s,server:%s,connType:%s,proxyUrl:%s", tp, vkey, server, connType, proxyUrl)
	var err error
	var connection net.Conn
	//	var sess *kcp.UDPSession
	if tp == "tcp" {
		//		if proxyUrl != "" {
		//			u, er := url.Parse(proxyUrl)
		//			if er != nil {
		//				return nil, er
		//			}
		//			switch u.Scheme {
		//			case "socks5":
		//				n, er := proxy.FromURL(u, nil)
		//				if er != nil {
		//					return nil, er
		//				}
		//				connection, err = n.Dial("tcp", server)
		//			case "http":
		//				connection, err = NewHttpProxyConn(u, server)
		//			}
		//		} else {
		connection, err = net.Dial("tcp", server)
		if err != nil {
			logger.Error("创建tcp测试连接失败,%s", err)
			return nil, err
		}
		logger.Info("创建tcp测试连接成功")
		//		}
	} else {
		//		sess, err = kcp.DialWithOptions(server, nil, 10, 3)
		//		if err == nil {
		//			conn.SetUdpSession(sess)
		//			connection = sess
		//		}
	}

	// 设置超时时间 30秒
	connection.SetDeadline(time.Now().Add(time.Second * 10))
	defer connection.SetDeadline(time.Time{})

	// 创建一个新连接
	logger.Error("新建通讯连接-start")
	c := conn.NewConn(connection)
	if _, err := c.Write([]byte(common.CONN_TEST)); err != nil {
		logger.Error("新建通讯连接失败,%v", err)
		return nil, err
	}
	logger.Info("新建通讯连接-success")
	// 发送客户端的版本信息
	if _, err := c.Write([]byte(crypt.Md5(version.GetVersion()))); err != nil {
		logger.Error("发送客户端版本信息失败,%s", err)
		return nil, err
	}
	// 取服务端返回的32个字节
	if b, err := c.GetShortContent(32); err != nil || crypt.Md5(version.GetVersion()) != string(b) {
		logger.Error("客户端和服务端版本不匹配，客户端当前版本是%s", version.GetVersion())
		return nil, err
	}
	// 发送验证结果
	if _, err := c.Write([]byte(common.Getverifyval(vkey))); err != nil {
		return nil, err
	}

	// 读取标记
	if s, err := c.ReadFlag(); err != nil {
		return nil, err
	} else if s == common.VERIFY_EER { //密钥验证错
		logger.Error("Validation key %s incorrect", vkey)
		os.Exit(0)
	}
	// 发送连接类型
	if _, err := c.Write([]byte(connType)); err != nil {
		return nil, err
	}

	//设置保活
	c.SetAlive(tp)

	return c, nil
}
