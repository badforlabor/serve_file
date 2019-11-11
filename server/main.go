/**
 * Auth :   liubo
 * Date :   2019/11/8 15:26
 * Comment:
 */

package main

import (
	"fmt"
	"serve_file/proto"
)

func main() {

	_fileManager := &fileManager{}
	_fileManager.start()

	_server := fileServer{}
	_server._fileManager = _fileManager
	_server.Start(fmt.Sprintf(":%d", proto.Port))
	_server.Serve("D:/workspace3/psl/PSL/Saved/StagedBuilds/WindowsNoEditor/PSLVR/Content/Paks/PSLVR-WindowsNoEditor.pak")

	var multiServer []*fileServer
	for i:=0; i<proto.DownloadThreadCount; i++ {
		mf := &fileServer{}
		mf._fileManager = _fileManager
		mf.Start(fmt.Sprintf(":%d", proto.Port + i + 1))
		multiServer = append(multiServer, mf)
	}

	_server.queue.Wait()

	_server.Stop()
	for _, v := range multiServer {
		v.Stop()
	}

	_fileManager.stop()

}