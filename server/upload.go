/**
 * Auth :   liubo
 * Date :   2019/11/8 14:57
 * Comment:
 */

package main

import (
	"fmt"
	"golib/ultraio"
	"io"
	"os"
	"serve_file/proto"
	"sort"
	"strconv"
	"sync"
	"time"
)

type fileMeta struct {
	file *os.File
	lastVisitTime time.Time
}

type FileManager struct {

	fileList []string

	fileMap map[string]*fileMeta
	checkTimer *time.Timer

	bDoDebug bool
	debugFile *os.File
	debugFileMutex sync.Mutex
	idList []int
	idMap map[int32]interface{}

	SingleLock sync.Mutex
}

func (self *FileManager) start() {

	self.fileMap = make(map[string]*fileMeta)
	self.timeCheckFile()

}
func (self *FileManager) stop() {
	if self.checkTimer != nil {
		self.checkTimer.Stop()
	}

	for k, v := range self.fileMap {
		if v == nil {
			continue
		}
		v.file.Sync()
		v.file.Close()
		self.fileMap[k] = nil
	}

}
func (self *FileManager) timeCheckFile() {

	const duration = time.Minute

	var now = time.Now()

	for k, v := range self.fileMap {
		if v == nil {
			continue
		}
		if v.lastVisitTime.Add(duration).Before(now) {
			self.fileMap[k] = nil

			// 有可能在“正在读取文件内容”的时候关闭，但是也不准备用锁了。
			// 因为，这种情况比较少见，另外尽管会出现读取失败，但是告知客户端，客户端只要再次发起，就ok了呀。
			v.file.Sync()
			v.file.Close()
		}
	}

	self.checkTimer = time.AfterFunc(duration, self.timeCheckFile)
}
func (self *FileManager) getFile(id int) *fileMeta {

	if id >= len(self.fileList) {
		return nil
	}

	filename := self.fileList[id]
	if len(filename) == 0 {
		return nil
	}

	v, ok := self.fileMap[filename]
	if ok && v != nil{
		v.lastVisitTime = time.Now()
	}

	if ok {
		return v
	} else {
		return nil
	}
}
func (self *FileManager) getRaw(fid int32, dataSize int32, offset int64) []byte {
	buff := make([]byte, dataSize)

	f := self.getFile(int(fid))
	if f != nil {
		n, err := f.file.ReadAt(buff, offset)

		if n == 0 {
			fmt.Println("读取文件错误！", offset)
		}

		if err == nil || err == io.EOF {
			buff = buff[:n]
			return buff
		}
	}

	return nil
}

func (self *FileManager) debugWriteRaw(msg proto.OneChunk) {
	if !self.bDoDebug {
		return
	}

	self.debugFileMutex.Lock()
	defer self.debugFileMutex.Unlock()

	self.idList = append(self.idList, int(msg.ChunkId))
	sort.Ints(self.idList)

	_, ok := self.idMap[msg.ChunkId]
	if ok {
		log.Warnf("重复的chunkid！")
	}
	if self.idMap == nil {
		self.idMap = make(map[int32]interface{})
	}
	self.idMap[msg.ChunkId]= nil

	if self.debugFile != nil {
		offset := int64(msg.ChunkId * msg.ConstChunkDataSize)
		n, err := self.debugFile.WriteAt(msg.Data, offset)
		self.debugFile.Sync()
		if err != nil {
			log.Warnf("写文件错误", err.Error())
		} else {
			if n != len(msg.Data) {
				log.Warnf("写文件错误", n, len(msg.Data))
			}
		}

		totalSum := int64(0)
		for i:=0; i<len(msg.Data); i++ {
			if msg.Data[i] < 0 {
				totalSum += int64(msg.Data[i]) * -1
			} else {
				totalSum += int64(msg.Data[i]) * 1
			}
		}
		if totalSum == 0 || len(msg.Data) == 0 {
			log.Warnf("没有数据！", msg.ChunkId, len(msg.Data))
		}
		ultraio.AppendFile("debug.txt", fmt.Sprintf("[%d]\t[%s]\n", msg.ChunkId, strconv.FormatInt(totalSum, 10)))
	} else {
		log.Warnf("没有文件！")
	}
}
func (self *FileManager) debugDone() {
	if !self.bDoDebug {
		return
	}

	self.debugFileMutex.Lock()
	defer self.debugFileMutex.Unlock()
	if self.debugFile != nil {
		self.debugFile.Sync()
		self.debugFile.Close()
		self.debugFile = nil
	}

	if len(self.idList) > 0 && len(self.idList) - 1 != self.idList[len(self.idList)-1] {
		log.Warnf("数据错误！")
	}
	fmt.Println("总共：", len(self.idList))
	self.idList = self.idList[0:0]
}
func (self *FileManager) beforeDownfile(id int) {

	filename := self.fileList[id]
	if len(filename) == 0 {
		return
	}

	{
		v, ok := self.fileMap[filename]
		if !ok || v == nil {
			file, err := os.Open(filename)
			if err == nil {
				v = &fileMeta{file:file, lastVisitTime:time.Now()}
				self.fileMap[filename] = v
			} else {
				v = nil
			}
		}
		if v != nil {
			v.lastVisitTime = time.Now()
		}
	}

	if self.bDoDebug {
		self.idMap = make(map[int32]interface{})
		self.idList = make([]int, 0)


		if self.debugFile == nil {
			file, _ := os.Stat(filename)
			filename1 := "debug.pak"
			os.Remove(filename1)
			self.debugFile, _ = os.Create(filename1)
			self.debugFile.Truncate(file.Size())
			self.debugFile.Sync()
		}
	}
}