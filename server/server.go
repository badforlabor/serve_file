/**
 * Auth :   liubo
 * Date :   2019/11/5 20:09
 * Comment: 文件服务器
 */

package main

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/examples/chat/proto"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/golog"
	"os"
	"reflect"
	p2 "serve_file/proto"
	"strconv"
)

const (

)

type IFileServer interface {
	Serve(filepath string)
	Start(addr string)
	Stop()
}
func NewFileServer() IFileServer {
	s := &fileServer{}
	return s
}

var log = golog.New("fileserver")

type fileServer struct {
	queue cellnet.EventQueue
	peer cellnet.GenericPeer

	_fileManager *FileManager
}

func (self *fileServer) Serve(filepath string) {
	self._fileManager.fileList = append(self._fileManager.fileList, filepath)
}
func (self *fileServer) Start(addr string) {

	// 创建一个事件处理队列，整个服务器只有这一个队列处理事件，服务器属于单线程服务器
	queue := cellnet.NewEventQueue()
	self.queue = queue

	// 创建一个tcp的侦听器，名称为server，连接地址为127.0.0.1:8801，所有连接将事件投递到queue队列,单线程的处理（收发封包过程是多线程）
	p := peer.NewGenericPeer("tcp.Acceptor", "server", addr, queue)
	self.peer = p

	// 设定封包收发处理的模式为tcp的ltv(Length-Type-Value), Length为封包大小，Type为消息ID，Value为消息内容
	// 每一个连接收到的所有消息事件(cellnet.Event)都被派发到用户回调, 用户使用switch判断消息类型，并做出不同的处理
	proc.BindProcessorHandler(p, "tcp.ltv", self.onMsg)

	// 开始侦听
	p.Start()

	// 事件队列开始循环
	queue.StartLoop()
}
func (self *fileServer) Stop() {
	self.queue.StopLoop()
}
func (self *fileServer) onMsg(ev cellnet.Event) {
	switch msg := ev.Message().(type) {
	// 有新的连接
	case *cellnet.SessionAccepted:
		log.Infof("server accepted")
	// 有连接断开
	case *cellnet.SessionClosed:
		log.Infof("session closed: ", ev.Session().ID())
	// 收到某个连接的ChatREQ消息
	case *proto.ChatREQ:
		// 准备回应的消息
		ack := proto.ChatACK{
			Content: msg.Content,       // 聊天内容
			Id:      ev.Session().ID(), // 使用会话ID作为发送内容的ID
		}

		// 在Peer上查询SessionAccessor接口，并遍历Peer上的所有连接，并发送回应消息（即广播消息）
		self.peer.(cellnet.SessionAccessor).VisitSession(func(ses cellnet.Session) bool {

			ses.Send(&ack)

			return true
		})

	case *p2.OneChunk:
		// 客户端请求这个chunk的内容
		self._fileManager.SingleLock.Lock()
		var resp = *msg
		if len(msg.Data) == 0 {
			data := self._fileManager.getRaw(msg.FileId, resp.ConstChunkDataSize, int64(msg.ChunkId * msg.ConstChunkDataSize))
			if data != nil {
				resp.Data = data
				resp.Md5Value = p2.Md5Value(resp.Data)
				ev.Session().Send(&resp)
				self._fileManager.debugWriteRaw(resp)
			} else {
				log.Warnf("无法获取文件：%d, %d", msg.FileId, msg.ChunkId)
			}
		} else {
			log.Warnf("服务器收到信息错误! %d, %d", msg.FileId, msg.ChunkId)
		}
		self._fileManager.SingleLock.Unlock()

	case *p2.CommonCommand:
		log.Infof("%s", msg.Cmd, msg.Param1)
		switch msg.Cmd {
		case "list":
			var filelist p2.FileListResp
			filelist.Files = append(filelist.Files, self._fileManager.fileList...)
			ev.Session().Send(&filelist)

		case "done":
			self._fileManager.debugDone()

		case "get":
			fileidx, _ := strconv.Atoi(msg.Param1)
			if fileidx >= 0 && fileidx < len(self._fileManager.fileList) {
				var fileinfo p2.FileChunkInfo
				var filename = self._fileManager.fileList[fileidx]
				file, err := os.Stat(filename)
				if err == nil {
					fileinfo.FileId = fileidx
					fileinfo.FileSize = file.Size()
					fileinfo.HashValue = p2.HashFile(filename)
				}
				if fileinfo.FileSize > 0 && len(fileinfo.HashValue) > 0 {
					ev.Session().Send(&fileinfo)
				}
			}
		}

	default:
		log.Infof("unknown msg: %s", reflect.TypeOf(ev.Message()).Name())
	}
}
