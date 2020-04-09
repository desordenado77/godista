// +build windows

package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func (godista *Godista) runCommand(cmdStr string, c net.Conn) {
	var cmd *exec.Cmd
	fmt.Println("-------- ", cmdStr)

	cmdArray := strings.SplitN(cmdStr, " ", 2)

	currentApp := godista.findApp(cmdArray[0])
	cmd = exec.Command(currentApp.Cmd, cmdArray[1:]...)
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
