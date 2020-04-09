// +build linux

package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func (godista *Godista) runCommand(cmdStr string, c net.Conn) {
	var cmd *exec.Cmd

	fmt.Printf("-------- '%s'\n", cmdStr)
	commands := strings.Split(cmdStr, " ")
	params := commands[1:]

	cmd = exec.Command(commands[0], params...)
	err := cmd.Start()
	if err != nil {
		Error.Println(err)
		c.Write([]byte(err.Error() + "\n"))
	} else {
		c.Write([]byte("\n"))
	}
}

func pathSeparator() string {
	return "/"
}
