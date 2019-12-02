/**
 * Auth :   liubo
 * Date :   2019/11/5 21:15
 * Comment: 下载文件的
 */

package main

import (
	"bytes"
	"fmt"
	"os"
	"serve_file/proto"
	"sync"
	"sync/atomic"
)

type downloader struct {
	socketList []*oneClient
	info proto.FileChunkInfo

	fileName string
	file *os.File
	fileLock sync.Mutex

	// 按文件块请求数据
	chunkList []int
	chunkReqCount int32
	chunkRespCount int32

	onDone proto.FunctionEvent
}
func (self *downloader) download(filename string, info proto.FileChunkInfo) {
	if self.file != nil {
		log.Warnf("正在下载文件！%s", self.file.Name())
		return
	}


	self.fileName = filename
	self.info = info

	self.file, _ = os.OpenFile(filename, os.O_CREATE | os.O_RDWR, os.ModePerm)
	if self.file == nil {
		log.Warnf("下载失败：%s", filename)
		return
	}

	// 如果文件大小不一样，那么删掉，重新创建，并填充各个块的数据为0
	fs, e := self.file.Stat()
	if e != nil || fs.Size() != self.info.FileSize {
		self.file.Sync()
		self.file.Close()
		os.Remove(filename)

		self.file, _ = os.Create(filename)
		if self.file == nil {
			log.Warnf("下载失败：%s", filename)
			return
		}
		self.file.Truncate(self.info.FileSize)
		self.file.Sync()
	}

	// 断点续传
	idList := make([]int, 0)
	self.info = info
	var chunkCount = int((self.info.FileSize + proto.ChunkSize - 1) / proto.ChunkSize)
	emptyBuffer := make([]byte, proto.ChunkSize)
	for i:=0; i< chunkCount; i++ {
		data := make([]byte, proto.ChunkSize)
		size, err := self.file.ReadAt(data, proto.ChunkSize)
		if err != nil {
			break
		}

		if bytes.Equal(emptyBuffer[:size], data[:size]) {
			idList = append(idList, i)
		}
	}

	// 如果所有的块都有数据，那么检查hash值一样不，如果不一样，那么重新下载
	if len(idList) == 0 {
		v := proto.HashFile(filename)
		if v == self.info.HashValue {
			self.done()
			return
		} else {
			idList = idList[0:0]
			for i:=0; i< chunkCount; i++ {
				idList = append(idList, i)
			}
		}
	}

	// 用socketlist去批量下载各个“块”
	for _, v := range self.socketList {
		v.CloseClient()
	}

	self.socketList = make([]*oneClient, proto.DownloadThreadCount)
	for i:=0; i<proto.DownloadThreadCount; i++ {
		one := newClient(self)
		one.onConn = func(client *oneClient) {
			self.reqSendChunk(client)
		}

		self.socketList[i] = one
		one.OpenClient(fmt.Sprintf( ":%d" , proto.Port + i + 1))
	}
	self.chunkList = idList
	self.chunkReqCount = int32(len(idList))
	self.chunkRespCount = 0
}
func (self *downloader) isWorking() bool {
	return self.file != nil
}
func (self *downloader) done() {
	log.Infof("下载完毕:%s", self.fileName)
	if self.info.HashValue != proto.HashFile(self.fileName) {
		log.Infof("    hash value not equal:%s", self.fileName)
	} else {
		log.Infof("    hash value equal!")
	}

	for _, v := range self.socketList {
		v.CloseClient()
	}
	self.socketList = make([]*oneClient, 0)


	self.fileLock.Lock()
	if self.file != nil {
		self.file.Sync()
		self.file.Close()
	}
	self.file = nil
	self.fileLock.Unlock()

	self.onDone.Broadcast()
}
func (self *downloader) writeTrunk(conn *oneClient, msg *proto.OneChunk) {
	if self.file != nil {
		offset := int64(msg.ChunkId * msg.ConstChunkDataSize)
		self.fileLock.Lock()
		n, err := self.file.WriteAt(msg.Data, offset)
		self.file.Sync()
		self.fileLock.Unlock()

		if err != nil {
			log.Warnf("写文件错误:%s", err.Error())
		}
		if n != len(msg.Data) {
			log.Warnf("写文件错误:%d, %d", n, len(msg.Data))
		}

		b := make([]byte, n)
		self.fileLock.Lock()
		self.file.ReadAt(b, offset)
		self.fileLock.Unlock()

		if !bytes.Equal( proto.Md5Value(b), msg.Md5Value ) {
			log.Warnf("写文件错误:%d, %d", offset, msg.ChunkId)
		}

		if true {
			if !bytes.Equal( proto.Md5Value(b), proto.Md5Value(msg.Data)) {
				log.Warnf("写文件错误:%d, %d", offset, msg.ChunkId)
			}
			md5v := proto.Md5Value(msg.Data)
			md5v2 := proto.Md5Value(b)
			if !bytes.Equal( msg.Md5Value, md5v) {
				log.Warnf("2 写文件错误:%d, %d, %016x, %016x, %016x", offset, msg.ChunkId, msg.Md5Value, md5v, md5v2)
			}
			if len(b) != len(msg.Data) {
				log.Warnf("3 写文件错误:%d, %d", offset, msg.ChunkId)
			}
			//ioutil.WriteFile("1.data", b, os.ModePerm)
			//ioutil.WriteFile("2.data", msg.Data, os.ModePerm)
		}

		atomic.AddInt32(&self.chunkRespCount, 1)
		if self.chunkRespCount == int32(len(self.chunkList)) {
			self.done()
		} else {
			self.reqSendChunk(conn)
		}
	} else {
		log.Warnf("接收到写文件失败！")
	}
}
func (self *downloader) reqSendChunk(conn *oneClient) {
	if self.chunkReqCount <= 0 {
		return
	}

	atomic.AddInt32(&self.chunkReqCount, -1)

	var resp proto.OneChunk
	resp.FileId = int32(self.info.FileId)
	resp.ConstChunkDataSize = proto.ChunkSize
	resp.ChunkId = int32(self.chunkList[self.chunkReqCount])
	conn.reqChunk(&resp)
}