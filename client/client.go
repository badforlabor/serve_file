/**
 * Auth :   liubo
 * Date :   2019/11/5 20:09
 * Comment: 客户端
 */

package main

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/examples/chat/proto"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/golog"
	"golib/ultraio"
	"os"
	"reflect"
	p2 "serve_file/proto"
	"time"
)

type IFileClient interface {
	OpenClient(addr string)
	CloseClient()
}

var log = golog.New("fileclient")
type oneClient struct {
	queue cellnet.EventQueue
	peer cellnet.TCPConnector

	worker *downloader
	pendingFile string

	onConn func(*oneClient)

	savedServerFiles []string

	testMode bool
	testFileName string
}
func NewClient() IFileClient {
	return newClient(nil)
}
func newClient(worker *downloader) *oneClient {
	v := &oneClient{}
	v.worker = worker
	if v.worker == nil {
		v.worker = &downloader{}
	}
	return v
}

func (self *oneClient) OpenClient(addr string) {
	self.worker.onDone.Add(self, self.onDownloadDone)

	// 创建一个事件处理队列，整个客户端只有这一个队列处理事件，客户端属于单线程模型
	queue := cellnet.NewEventQueue()
	self.queue = queue

	// 创建一个tcp的连接器，名称为client，连接地址为127.0.0.1:8801，将事件投递到queue队列,单线程的处理（收发封包过程是多线程）
	p := peer.NewGenericPeer("tcp.Connector", "client", addr, queue)
	self.peer = p.(cellnet.TCPConnector)
	self.peer.SetReconnectDuration(time.Second)

	// 设定封包收发处理的模式为tcp的ltv(Length-Type-Value), Length为封包大小，Type为消息ID，Value为消息内容
	// 并使用switch处理收到的消息
	proc.BindProcessorHandler(p, "tcp.ltv", self.onMsg)

	// 开始发起到服务器的连接
	p.Start()

	// 事件队列开始循环
	queue.StartLoop()
}

func (self *oneClient) sendMsg(msg interface{}) {
	var p = self.peer
	p.(interface {
		Session() cellnet.Session
	}).Session().Send(msg)
}

func (self *oneClient) CloseClient() {
	self.worker.onDone.Remove(self, self.onDownloadDone)
	self.queue.StopLoop()
}
func (self *oneClient) onMsg(ev cellnet.Event) {
	switch msg := ev.Message().(type) {
	case *cellnet.SessionConnected:
		log.Infof("client connected")
		if self.onConn != nil {
			self.onConn(self)
		}
	case *cellnet.SessionClosed:
		log.Infof("client error")
	case *proto.ChatACK:
		log.Infof("sid%d say: %s", msg.Id, msg.Content)
	case *p2.CommonCommand:
		log.Infof("%s", msg.Cmd, msg.Param1)
		switch msg.Cmd {

		}
	case *p2.FileChunkInfo:
		self.worker.download(self.pendingFile, *msg)

	case *p2.OneChunk:
		if len(msg.Data) > 0 {
			self.worker.writeTrunk(self, msg)
		} else {
			log.Warnf("下载trunk失败", msg.ChunkId)
			// 重新再获取一下，此时服务器出错了，不能再次请求
			// self.reqChunk(msg)
		}

	case *p2.FileListResp:
		for i, v := range msg.Files {
			fmt.Printf("    [%d]:%s\n", i+1, v)
		}
		self.savedServerFiles = msg.Files

	default:
		log.Infof("unknown msg: %s", reflect.TypeOf(ev.Message()).Name())
	}
}
func (self *oneClient) reqChunk(msg *p2.OneChunk) {
	self.peer.Session().Send(msg)
}
func (self *oneClient) reqGet(idx int, topath string) {
	if len(self.pendingFile) > 0 {
		log.Warnf("正在获取文件，请稍后再试")
		return
	}
	if idx < 0 {
		log.Warnf("输入索引错误")
		return
	}

	self.pendingFile = topath
	self.sendMsg(&p2.CommonCommand{Cmd:"get", Param1:fmt.Sprintf("%d", idx - 1)})
}
func (self *oneClient) onDownloadDone() {
	self.pendingFile = ""

	go func() {
		self.sendMsg(&p2.CommonCommand{Cmd:"done"})
		time.Sleep(time.Second)

		if self.testMode {

			// 比较文件
			var b = p2.CompareFile("debug.pak", self.testFileName)
			ultraio.AppendFile("result.txt", fmt.Sprintf("比较文件[%s][%s], 结果:%t\n", "debug.pak", self.testFileName, b))
			time.Sleep(time.Second * 1)

			b = p2.CompareFile("debug.pak", self.savedServerFiles[0])
			ultraio.AppendFile("result.txt", fmt.Sprintf("比较文件[%s][%s], 结果:%t\n\n\n", "debug.pak", self.savedServerFiles[0], b))
			time.Sleep(time.Second * 1)

			var err = os.Remove(self.testFileName)
			if err != nil {
				fmt.Println(err.Error())
			}
			err = os.Remove("debug.pak")
			if err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(time.Second * 3)

			self.Test()
		}
	}()
}
func (self *oneClient) Test() {
	self.testMode = true
	self.testFileName = "test.pak"
	ultraio.AppendFile("result.txt", fmt.Sprintf("start test...\n"))

	self.sendMsg(&p2.CommonCommand{Cmd:"list"})
	time.Sleep(time.Second)

	self.reqGet(1, self.testFileName)
}