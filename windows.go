// +build windows

package main

import (
	"net"
	"os/exec"
	"strings"
)

func (godista *Godista) runCommand(cmdStr string, c net.Conn) {
	var cmd *exec.Cmd
	Trace.Println("Received: ", cmdStr)

	cmdArray := strings.Split(cmdStr, " ")

	currentApp := godista.findApp(cmdArray[0])

	Trace.Println("Command:", currentApp.Cmd)
	Trace.Println("Params:", cmdArray[1:])

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
