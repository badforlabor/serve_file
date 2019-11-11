/**
 * Auth :   liubo
 * Date :   2019/11/8 17:36
 * Comment: 二进制比较文件
 */

package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		return
	}

	equal := true

	f1, err1 := os.Open(os.Args[1])
	f2, err2 := os.Open(os.Args[2])
	if err1 == nil || err2 == nil {

		count := 32 * 1024
		buff1 := make([]byte, 32 * 1024)
		buff2 := make([]byte, 32 * 1024)

		for {
			n1, err1 := f1.Read(buff1)
			n2, err2 := f2.Read(buff2)
			if n1 != n2 || err1 != nil || err2 != nil {
				fmt.Println("1")
				equal = false
				break
			}
			b := bytes.Equal(buff1[:n1], buff2[:n2])
			offset, _ := f1.Seek(0, os.SEEK_CUR)
			if !b {
				fmt.Println("2", offset)
				equal = false
				continue
			}

			if n1 < count {
				if !equal {
					break
				}
				fmt.Println("两个文件相同！", n1, count)
				return
			}
		}
	}
	fmt.Println("两个文件不同")
}
