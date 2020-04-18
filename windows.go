// +build windows

package main

import (
<<<<<<< HEAD
=======
	"bytes"
	"fmt"
>>>>>>> Added an option to wait for the command to finish and get the output
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
	if currentApp.Wait {
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		if err != nil {
			Error.Println(err)
		}
		c.Write(out.Bytes())
		c.Write([]byte("\n"))
		Trace.Println(out.Bytes())
	} else {
		err := cmd.Start()
		if err != nil {
			Error.Println(err)
			c.Write([]byte(err.Error() + "\n"))
		} else {
			c.Write([]byte("\n"))
		}
	}
}

func pathSeparator() string {
	return "\\"
}
