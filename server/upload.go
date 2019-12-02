/**
 * Auth :   liubo
 * Date :   2019/11/8 14:57
 * Comment:
 */

package main

import (
	"fmt"
	"io"
	"os"
	"serve_file/proto"
	"sort"
	"sync"
	"time"
)

type fileMeta struct {
	readMutex sync.Mutex

	// 多线程读写IO的时候，得加锁
	file *os.File
	lastVisitTime time.Time
}

type FileManager struct {

	fileList []string

	fileMap map[string]*fileMeta
	fileMapMutex sync.RWMutex
	checkTimer *time.Timer

	debugFile *os.File
	debugFileMutex sync.Mutex
	idList []int

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
	self.fileMapMutex.Lock()
	for _, v := range self.fileMap {
		v.file.Sync()
		v.file.Close()
		v.file = nil
	}
	self.fileMap = make(map[string]*fileMeta)
	self.fileMapMutex.Unlock()

}
func (self *FileManager) timeCheckFile() {

	return

	const duration = time.Minute

	self.fileMapMutex.RLock()
	var now = time.Now()
	var valid = false
	for _, v := range self.fileMap {
		if  v.lastVisitTime.Add(duration).Before(now) {
			valid = true
			break
		}
	}
	self.fileMapMutex.RUnlock()

	if valid {
		self.fileMapMutex.Lock()
		for k, v := range self.fileMap {
			if  v.lastVisitTime.Add(duration).Before(now) {
				v.readMutex.Lock()
				v.file.Sync()
				v.file.Close()
				v.file = nil
				v.readMutex.Unlock()

				delete(self.fileMap, k)
			}
		}
		self.fileMapMutex.Unlock()
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

	self.debugFileMutex.Lock()

	if self.debugFile != nil {
		self.debugFile.Sync()
	}

	if self.debugFile == nil {
		file, _ := os.Stat(filename)
		filename1 := "debug.pak"
		os.Remove(filename1)
		self.debugFile, _ = os.Create(filename1)
		self.debugFile.Truncate(file.Size())
		self.debugFile.Sync()
	}

	self.debugFileMutex.Unlock()

	self.fileMapMutex.RLock()
	v, ok := self.fileMap[filename]
	if ok {
		v.lastVisitTime = time.Now()
	}
	self.fileMapMutex.RUnlock()
	if ok {
		return v
	}

	self.fileMapMutex.Lock()
	v, ok = self.fileMap[filename]
	if !ok {
		file, err := os.Open(filename)
		if err == nil {
			v = &fileMeta{file:file, lastVisitTime:time.Now()}
			self.fileMap[filename] = v
		} else {
			v = nil
		}
	}
	self.fileMapMutex.Unlock()

	return v
}
func (self *FileManager) getRaw(fid int32, dataSize int32, offset int64) []byte {
	buff := make([]byte, dataSize)

	f := self.getFile(int(fid))
	if f != nil {
		f.readMutex.Lock()
		n, err := f.file.ReadAt(buff, offset)
		f.readMutex.Unlock()

		if err == nil || err == io.EOF {
			buff = buff[:n]
			return buff
		}
	}

	return nil
}

func (self *FileManager) debugWriteRaw(msg proto.OneChunk) {
	self.debugFileMutex.Lock()
	defer self.debugFileMutex.Unlock()

	self.idList = append(self.idList, int(msg.ChunkId))
	sort.Ints(self.idList)

	if self.debugFile != nil {
		offset := int64(msg.ChunkId * msg.ConstChunkDataSize)
		self.debugFile.WriteAt(msg.Data, offset)
	}
	if msg.ChunkId == 51160 {
		fmt.Println("1234")
	}

}
func (self *FileManager) debugDone() {
	self.debugFileMutex.Lock()
	defer self.debugFileMutex.Unlock()
	if self.debugFile != nil {
		self.debugFile.Sync()
		self.debugFile.Close()
		self.debugFile = nil
	}

	for i:=0; i<len(self.idList); i++ {
		fmt.Println(self.idList[i])
	}
	fmt.Println("总共：", len(self.idList))
	self.idList = self.idList[0:0]
}