// +build linux

package main

import (
	"fmt"
	"net"
)

func runCommand(cmd string, c net.Conn) {
	fmt.Println("-------- ", cmd)

	c.Write([]byte(""))
}

func pathSeparator() string {
	return "/"
}
