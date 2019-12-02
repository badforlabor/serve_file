/**
 * Auth :   liubo
 * Date :   2019/11/8 17:36
 * Comment: 二进制比较文件
 */

package main

import (
	"bytes"
	"fmt"
	"io"
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
				}
				continue
			}

			if n1 < count {
				if !equal {
					break
				}
				fmt.Println("两个文件相同！")
				return
			}
		}
	}
	fmt.Println("两个文件不同")
}
