/**
 * Auth :   liubo
 * Date :   2019/11/5 20:08
 * Comment:
 */

package proto


import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	_ "github.com/davyxu/cellnet/codec/binary"
	_ "github.com/davyxu/cellnet/codec/json"
	_ "github.com/davyxu/cellnet/codec/protoplus"
	"github.com/davyxu/cellnet/util"
	"reflect"
)

const (
	ChunkSize           = 1024 * 32
	Port                = 10110
	DownloadThreadCount = 1
)

/**************************************************************************/
/* 以下是协议                                                             */
/**************************************************************************/
func init() {

	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("json"),
		Type:  reflect.TypeOf((*FileChunkInfo)(nil)).Elem(),
		ID:    int(util.StringHash("FileChunkInfo")),
	})

	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*OneChunk)(nil)).Elem(),
		ID:    int(util.StringHash("OneChunk")),
	})

	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("json"),
		Type:  reflect.TypeOf((*CommonCommand)(nil)).Elem(),
		ID:    int(util.StringHash("CommonCommand")),
	})

	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("json"),
		Type:  reflect.TypeOf((*FileListResp)(nil)).Elem(),
		ID:    int(util.StringHash("FileListResp")),
	})
}
type FileChunkInfo struct {
	FileId     int
	FileSize   int64
	HashValue  string
}
type OneChunk struct {
	FileId             int32
	ChunkId            int32
	Data               []byte
	Md5Value		   []byte
	ConstChunkDataSize int32
}
type CommonCommand struct {
	Cmd string
	Param1 string
}
type FileListResp struct {
	Files []string
}
