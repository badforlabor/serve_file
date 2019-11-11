/**
 * Auth :   liubo
 * Date :   2019/11/8 9:35
 * Comment: server和client都通用的函数
 */

package proto

import (
	"crypto/md5"
	"fmt"
	"github.com/schollz/croc/v6/src/utils"
)

func HashFile(fullpath string) string {
	// hash, err := utils.HashFile(fullpath) // 这个不准确！
	hash, err := utils.MD5HashFile(fullpath)
	if err == nil {
		return fmt.Sprintf("%016x", hash)
	}
	return ""
}
func Md5File(fullpath string) string {
	hash, err := utils.MD5HashFile(fullpath)
	if err == nil {
		return fmt.Sprintf("%016x", hash)
	}
	return ""
}
func Md5Value(b []byte) []byte {
	v := md5.New()
	v.Write(b)
	hash := v.Sum(nil)
	//return fmt.Sprintf("%016x", hash)
	return hash
}