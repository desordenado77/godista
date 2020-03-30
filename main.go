package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
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
	Ip   string `json:"ip"`
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
	abs = strings.Replace(abs, "/", "\\", -1)

	return abs

}

func (godista *Godista) ParseConfig() (err error) {

	// read config file from:
	// 1. GODISTA_CONF env variable
	// 2. .godista/config.json in ~/
	// 3. in ./config.json

	var jsonFile *os.File

	c, exist := os.LookupEnv(CONFIGENV)

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

	Trace.Println(godista.conf)

	return nil
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

	err = godista.ParseConfig()
	if err != nil {
		Error.Println("Exiting")
		os.Exit(1)
	}

	if *optCommand != "" {
		Trace.Println("Non Empty command", *optCommand)
		godista.currentApp = godista.findApp(*optCommand)

		if godista.currentApp == nil {
			Error.Println("Unknown Application", *optCommand)
			os.Exit(1)
		}

		if *optParams != "" {
			regex := godista.currentApp.Params
			Trace.Println("Regex for Command", *optCommand, "is", regex)

			if regex != "" {
				re := regexp.MustCompile(regex)
				
				matches := re.FindAllStringSubmatch(*optParams, -1)

				newParams := *optParams
				for i,e := range matches[0] {
					if i == 0 {
						continue
					}
					newParams = strings.Replace(newParams, e, godista.replacePath(e), 1)
				}

				Trace.Println("New Params:", newParams)
			}
		}

	}

}
