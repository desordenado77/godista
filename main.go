package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pborman/getopt"
)

const CONFIGENV = "GODISTA_CONF"
const CONFIG_FILENAME = "config.json"
const GODISTA_SETTINGS_FOLDER_PATTERN = "%s/.godista/"

var GODISTA_SETTINGS_FOLDER string

type ClientCfg struct {
	PathForServer string `json:"path_for_server"`
	PathForClient string `json:"path_for_client"`
}

type ServerCfg struct {
	Ip   string `json:"ip_file"`
	Port int    `json:"port"`
}

type AppCfg struct {
	Name   string `json:"name"`
	Cmd    string `json:"cmd"`
	Params string `json:"params"`
}

type Config struct {
	Client ClientCfg `json:"client"`
	Server ServerCfg `json:"server"`
	Apps   []AppCfg  `json:"apps"`
}

type Godista struct {
	conf       Config
	ip         string
	currentApp *AppCfg
}

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLogs(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func (godista *Godista) findApp(cmd string) (app *AppCfg) {
	for i, s := range godista.conf.Apps {
		if cmd == s.Name {
			return &godista.conf.Apps[i]
		}
	}
	return nil
}

func (godista *Godista) replacePath(str string) (path string) {
	// 1. Get absolute path for str
	// 2. Replace path_for_client with path_for_server
	abs, err := filepath.Abs(str)
	if err != nil {
		Error.Println("Error getting Absolute path:", str)
		Error.Println("Error:", err)
		os.Exit(1)
	}
	abs = strings.Replace(abs, godista.conf.Client.PathForClient, godista.conf.Client.PathForServer, 1)
	abs = strings.Replace(abs, "/", pathSeparator(), -1)

	return abs

}

func (godista *Godista) ParseConfig(s bool, p string) (err error) {

	// read config file from:
	// 1. GODISTA_CONF env variable
	// 2. .godista/config.json in ~/
	// 3. in ./config.json

	var jsonFile *os.File

	var c string
	var exist bool

	if p != "" {
		c = p
		exist = true
	} else {
		c, exist = os.LookupEnv(CONFIGENV)
	}

	if exist {
		jsonFile, err = os.Open(c + "/" + CONFIG_FILENAME)
		if err != nil {
			Warning.Println(err)
		}
	} else {
		Warning.Println(CONFIGENV + " environment variable not found")
		err = os.ErrNotExist
	}

	if err != nil {
		jsonFile, err = os.Open(GODISTA_SETTINGS_FOLDER + CONFIG_FILENAME)
		if err != nil {
			Warning.Println(err)
			jsonFile, err = os.Open("./" + CONFIG_FILENAME)
			if err != nil {
				Warning.Println(err)
			}
		}
	}

	if err != nil {
		Error.Println("Unable to open config file")
		return err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		Error.Println("Unable to read config file: " + jsonFile.Name())
		return err
	}
	err = json.Unmarshal(byteValue, &godista.conf)

	if err != nil {
		Error.Println("Unable to read config file: " + jsonFile.Name())
		Error.Println(err)
		return err
	}

	var ipPath string
	if s {
		ipPath = godista.conf.Client.PathForServer + pathSeparator() + godista.conf.Server.Ip
		ipPath = strings.Replace(ipPath, "/", pathSeparator(), -1)
	} else {
		ipPath = godista.conf.Client.PathForClient + "/" + godista.conf.Server.Ip
	}

	dat, err := ioutil.ReadFile(ipPath)
	if err != nil {
		Error.Println("Unable to read file: " + godista.conf.Server.Ip)
		Error.Println(err)
		return err
	}

	godista.ip = string(dat)

	Trace.Println("Server IP address", godista.ip)

	Trace.Println(godista.conf)

	return nil
}

func (godista *Godista) IPMenu(r *bufio.Reader) {
	var s []net.IP

	fmt.Println("\n------------------------------------")
	fmt.Println("Select IP address:")
	fmt.Println("")
	ifaces, _ := net.Interfaces()

	k := 0
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			s = append(s, ip)
			// process IP address
			fmt.Printf("%d) %s\n", k, ip)
			k = k + 1
		}
	}
	fmt.Println("")
	fmt.Println("Select IP Address?")
	text, _ := r.ReadString('\n')
	text = strings.Replace(strings.ToLower(text), "\n", "", -1)
	i, err := strconv.Atoi(text)
	if err == nil && i >= 0 && i < len(s) {
		fmt.Println("IP Selected", s[i])

		ip_byte := []byte(string(s[i].String()))
		err := ioutil.WriteFile(godista.conf.Client.PathForServer+pathSeparator()+godista.conf.Server.Ip, ip_byte, 0644)
		if err != nil {
			Error.Println("Error Writing to file", err)
			os.Exit(1)
		}

	} else {
		fmt.Println("Incorrect IP selected")
	}

}

func (godista *Godista) MainMenu(r *bufio.Reader) {
	fmt.Println("\n------------------------------------")
	fmt.Println("Main Menu\n")
	fmt.Println("0) Renew IP Address")
	fmt.Println("1) Exit")
	fmt.Println("")
	fmt.Println("Select Option?")
	text, _ := r.ReadString('\n')
	text = strings.Replace(strings.ToLower(text), "\n", "", -1)

	if text == "0" {
		godista.IPMenu(r)
	}

	if text == "1" {
		os.Exit(0)
	}

}

func usage() {
	w := os.Stdout

	getopt.PrintUsage(w)

	fmt.Println("Extended usage goes here")
}

func main() {

	InitLogs(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	usr, err := user.Current()
	if err != nil {
		Error.Println("Error getting current user info: ")
		Error.Println(err)
		os.Exit(1)
	}
	//fmt.Println(usr)

	GODISTA_SETTINGS_FOLDER = fmt.Sprintf(GODISTA_SETTINGS_FOLDER_PATTERN, usr.HomeDir)

	getopt.SetUsage(usage)

	optHelp := getopt.BoolLong("Help", 'h', "Show this message")
	optVerbose := getopt.IntLong("Verbose", 'v', 0, "Set verbosity: 0 to 3. Verbose set to -1 everything goes to stderr. This is used for the cd case in which the output of the application goes to cd.")
	optCommand := getopt.StringLong("Command", 'c', "", "Command to send")
	optParams := getopt.StringLong("Params", 'p', "", "Command parameters")
	optServer := getopt.BoolLong("Server", 's', "Server Mode")
	optConfigPath := getopt.StringLong("config", 'f', "", "Path to config file. Ideal to use for server mode.")

	getopt.Parse()

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}

	vw := ioutil.Discard
	if *optVerbose > 0 {
		vw = os.Stdout
	}

	vi := ioutil.Discard
	if *optVerbose > 1 {
		vi = os.Stdout
	}

	vt := ioutil.Discard
	if *optVerbose > 2 {
		vt = os.Stdout
	}

	if *optVerbose == -1 {
		InitLogs(os.Stderr, os.Stderr, os.Stderr, os.Stderr)
	}

	var godista Godista

	InitLogs(vt, vi, vw, os.Stderr)

	err = godista.ParseConfig(*optServer, *optConfigPath)
	if err != nil {
		Error.Println("Exiting")
		os.Exit(1)
	}

	if *optServer {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(godista.conf.Server.Port))
		if err != nil {
			// handle error
			Error.Println("Network Error:", err)
			os.Exit(1)
		}
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					Error.Println("Network Error:", err)
					os.Exit(1)
				}
				go func(c net.Conn) {
					defer c.Close()
					buf := make([]byte, 2048)
					_, err := c.Read(buf)
					if err != nil {
						Error.Println("Network Error:", err)
						os.Exit(1)
					}
					fmt.Println("Received:", string(buf))

					runCommand(strings.TrimRight(strings.TrimRight(string(buf), "\x00"), "\n"), c)

					c.Close()
				}(conn)
			}
		}()

		reader := bufio.NewReader(os.Stdin)
		for {
			godista.MainMenu(reader)
		}
		os.Exit(0)

	}

	if *optCommand != "" {
		Trace.Println("Non Empty command", *optCommand)
		godista.currentApp = godista.findApp(*optCommand)

		if godista.currentApp == nil {
			Error.Println("Unknown Application", *optCommand)
			os.Exit(1)
		}

		newParams := *optParams
		if *optParams != "" {
			regex := godista.currentApp.Params
			Trace.Println("Regex for Command", *optCommand, "is", regex)

			if regex != "" {
				re := regexp.MustCompile(regex)

				matches := re.FindAllStringSubmatch(*optParams, -1)

				for i, e := range matches[0] {
					if i == 0 {
						continue
					}
					newParams = strings.Replace(newParams, e, godista.replacePath(e), 1)
				}

				Trace.Println("New Params:", newParams)
			}
		}
		conn, err := net.Dial("tcp", godista.ip+":"+strconv.Itoa(godista.conf.Server.Port))
		if err != nil {
			Error.Println("Error connecting to", godista.ip+":"+strconv.Itoa(godista.conf.Server.Port), err)
			os.Exit(1)
		}
		fmt.Fprintf(conn, godista.currentApp.Cmd+" "+newParams+"\n")
		Trace.Println("Sent")
		status, err := bufio.NewReader(conn).ReadString('\n')
		if err == nil {
			fmt.Println(status)
		} else {
			if err != io.EOF {
				Error.Println("Network error:", err)
			}
		}
	}
}
