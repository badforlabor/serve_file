/**
 * Auth :   liubo
 * Date :   2019/11/8 9:35
 * Comment: server和client都通用的函数
 */

package proto

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/schollz/croc/v6/src/utils"
	"io"
	"io/ioutil"
	"os"
	"strconv"
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

// 比较文件
func CompareFile(file1, file2 string) bool {
	return CompareFileDetail(file1, file2, false)
}
func CompareFileDetail(file1, file2 string, detail bool) bool {
	equal := true

	f1, err1 := os.Open(file1)
	f2, err2 := os.Open(file2)
	if err1 == nil || err2 == nil {

		count := 32 * 1024
		buff1 := make([]byte, count)
		buff2 := make([]byte, count)

		for {
			n1, err1 := f1.Read(buff1)
			n2, err2 := f2.Read(buff2)
			if err1 != nil || err2 != nil {
				if err1 == io.EOF && err2 == io.EOF {

				} else {
					fmt.Println("1")
					equal = false
					break
				}
			}
			b := bytes.Equal(buff1[:n1], buff2[:n2])
			offset1, _ := f1.Seek(0, os.SEEK_CUR)
			offset2, _ := f2.Seek(0, os.SEEK_CUR)
			if !b {
				fmt.Println("2", offset1, offset2)
				b = true
				for i:=0; i<n1; i++ {
					if buff1[i] != buff2[i] {
						b = false
						fmt.Println("二次比较失败！", i, int(buff1[i]), int(buff2[i]))
						break
					}
				}
				if !b {
					fmt.Println("二次比较失败！")
					equal = false
					if detail {
						folder := "./result/"
						os.Mkdir(folder, os.ModePerm)
						ioutil.WriteFile(folder + strconv.Itoa(int(offset1)) + "_src", buff1[:n1], os.ModePerm)
						ioutil.WriteFile(folder + strconv.Itoa(int(offset2)) + "_dst", buff2[:n2], os.ModePerm)
					}
				}
				continue
			}

			if n1 < count {
				if equal {
					fmt.Println("两个文件相同！")
				}
				break
			}
		}

		f1.Close()
		f2.Close()
	}

	return equal
}