/**
 * Auth :   liubo
 * Date :   2019/11/8 17:36
 * Comment: 二进制比较文件
 */

package main

import (
	"fmt"
	"os"
	"serve_file/proto"
)

func main() {
	if len(os.Args) != 3 {
		return
	}

	equal := proto.CompareFile(os.Args[1], os.Args[2])
	if !equal {
		fmt.Println("两个文件不同")
	} else {
		fmt.Println("两个文件相同")
	}
}
