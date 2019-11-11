/**
 * Auth :   liubo
 * Date :   2019/11/8 17:31
 * Comment:
 */

package main

import (
	"fmt"
	"os"
	"serve_file/proto"
)

func main() {
	if len(os.Args) != 2 {
		return
	}
	fmt.Println(proto.Md5File( os.Args[1]))
}
