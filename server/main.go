/**
 * Auth :   liubo
 * Date :   2019/11/8 15:26
 * Comment:
 */

package main

import (
	"fmt"
	"github.com/davyxu/golog"
	"serve_file/proto"
)

func main() {
	golog.SetOutputToFile("log/server.log", golog.OutputFileOption{
		MaxFileSize: 1000,
	})

	log.SetLevel(golog.Level_Info)
	golog.VisitLogger(".*", func(logger *golog.Logger) bool {
		logger.SetLevel(golog.Level_Info)
		return true
	})

	_fileManager := &FileManager{}
	_fileManager.start()
	//_fileManager.bDoDebug = true

	_server := fileServer{}
	_server._fileManager = _fileManager
	_server.Start(fmt.Sprintf(":%d", proto.Port))
	//_server.Serve("D:/workspace3/psl/PSL/Saved/StagedBuilds/WindowsNoEditor/PSLVR/Content/Paks/PSLVR-WindowsNoEditor.pak")
	_server.Serve("F:/workspace2/nomanscity-eng/CityLevel_Finish/Content/Movies/mp4.mp4")

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