/**
 * Auth :   liubo
 * Date :   2019/11/8 14:57
 * Comment:
 */

package main

import (
	"io"
	"os"
	"sync"
	"time"
)

type fileMeta struct {
	readMutex sync.Mutex

	// 多线程读写IO的时候，得加锁
	file *os.File
	lastVisitTime time.Time
}

type fileManager struct {

	fileList []string

	fileMap map[string]*fileMeta
	fileMapMutex sync.RWMutex
	checkTimer *time.Timer
}

func (self *fileManager) start() {

	self.fileMap = make(map[string]*fileMeta)
	self.timeCheckFile()

}
func (self *fileManager) stop() {
	if self.checkTimer != nil {
		self.checkTimer.Stop()
	}
	self.fileMapMutex.Lock()
	for _, v := range self.fileMap {
		v.file.Close()
		v.file = nil
	}
	self.fileMap = make(map[string]*fileMeta)
	self.fileMapMutex.Unlock()

}
func (self *fileManager) timeCheckFile() {

	const duration = time.Minute

	self.fileMapMutex.RLock()
	var now = time.Now()
	var valid = false
	for _, v := range self.fileMap {
		if  v.lastVisitTime.Add(duration).After(now) {
			valid = true
			break
		}
	}
	self.fileMapMutex.RUnlock()

	if valid {
		self.fileMapMutex.Lock()
		for k, v := range self.fileMap {
			if  v.lastVisitTime.Add(duration).After(now) {
				v.readMutex.Lock()
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
func (self *fileManager) getFile(id int) *fileMeta {

	if id >= len(self.fileList) {
		return nil
	}

	filename := self.fileList[id]
	if len(filename) == 0 {
		return nil
	}

	self.fileMapMutex.RLock()
	v, ok := self.fileMap[filename]
	if ok {
		v.lastVisitTime = time.Now()
	}
	self.fileMapMutex.RUnlock()
	if ok {
		return v
	}

	file, err := os.Open(filename)
	if err == nil {
		v = &fileMeta{file:file, lastVisitTime:time.Now()}
		self.fileMapMutex.Lock()
		self.fileMap[filename] = v
		self.fileMapMutex.Unlock()
	} else {
		v = nil
	}

	return v
}
func (self *fileManager) getRaw(fid int32, dataSize int32, offset int64) []byte {
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