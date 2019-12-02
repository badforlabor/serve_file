/**
 * Auth :   liubo
 * Date :   2019/11/8 15:26
 * Comment:
 */

package main

import (
	"bufio"
	"fmt"
	"os"
	"serve_file/proto"
	"strconv"
	"strings"
)

func main() {
	c := newClient(nil)
	c.OpenClient(fmt.Sprintf(":%d", proto.Port))

	ReadConsole(func(s string) {

		cmds := strings.Split(s, " ")
		if len(cmds) == 0 {
			return
		}

		if cmds[0] == "list" {
			c.sendMsg(&proto.CommonCommand{Cmd:cmds[0]})
		} else if cmds[0] == "get" {
			if len(cmds) == 3 {
				idx, _ := strconv.Atoi( cmds[1])
				c.reqGet(idx, cmds[2])
			}
		} else if cmds[0] == "done" {
			c.sendMsg(&proto.CommonCommand{Cmd:cmds[0]})
		} else if cmds[0] == "test" {
			c.Test()
		}
	})

	c.CloseClient()
}

func ReadConsole(callback func(string)) {

	for {

		// 从标准输入读取字符串，以\n为分割
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			break
		}

		// 去掉读入内容的空白符
		text = strings.TrimSpace(text)

		if text == "exit" {
			break
		}

		callback(text)
	}

}
