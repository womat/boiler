package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/womat/debug"

	"boiler/pkg/config"
)

// defaultInterval defines the default of dataCollectionInterval and backupInterval (in seconds)
const (
	defaultDataCollectionInterval = 60
	defaultBackupInterval         = 60
	defaultPort                   = 4000
	defaultDataFile               = "/opt/womat/data/" + MODULE + ".yaml"
	defaultDebugFile              = "stderr"
	defaultDebugFlag              = "standard"
)

type yamlDebug struct {
	File string `yaml:"file"`
	Flag string `yaml:"flag"`
}

type yamlWebserver struct {
	Port        int             `yaml:"port"`
	Webservices map[string]bool `yaml:"webservices"`
}

type yamlStruct struct {
	DataCollectionInterval int           `yaml:"datacollectioninterval"`
	BackupInterval         int           `yaml:"backupintervall"`
	DataFile               string        `yaml:"datafile"`
	Debug                  yamlDebug     `yaml:"debug"`
	WebServer              yamlWebserver `yaml:"webserver"`
	UVS232URL              string        `yaml:"uvs232url"`
	MeterURL               string        `yaml:"meterurl"`
}

func initConfig() {
	var err error
	var flags = config.Flag{
		"version":   {FlagType: config.FlagBool, Usage: "print version and exit", DefaultValue: false},
		"debugFile": {FlagType: config.FlagString, Usage: "log file eg. " + filepath.Join("/opt/womat/log/"+MODULE+".log") + `(default "stderr")`, DefaultValue: ""},
		"debugFlag": {FlagType: config.FlagString, Usage: `"enable debug information (standard | trace | debug) (default "standard")`, DefaultValue: ""},
		"config":    {FlagType: config.FlagString, Usage: "Config File", DefaultValue: filepath.Join("/opt/womat/conf/" + MODULE + ".yaml")},
	}
	var configFile = yamlStruct{
		DataCollectionInterval: defaultDataCollectionInterval,
		BackupInterval:         defaultBackupInterval,
		DataFile:               defaultDataFile,
		Debug:                  yamlDebug{File: defaultDebugFile, Flag: defaultDebugFlag},
		WebServer:              yamlWebserver{Port: defaultPort, Webservices: map[string]bool{"version": false, "currentdata": false}},
	}

	Config = Configuration{Webserver: WebserverConf{Webservices: map[string]bool{}}}

	config.Parse(flags)
	if flags.Bool("version") {
		fmt.Printf("Version: %v\n", VERSION)
		os.Exit(0)
	}

	if err := readConfigFile(flags.String("config"), &configFile); err != nil {
		log.Fatalf("Error reading config file, %v", err)
	}
	if Config.Debug.File, Config.Debug.Flag, err = getDebugConfig(flags, configFile.Debug); err != nil {
		log.Fatalf("unable to open debug file, %v", err)
	}
	Config.DataCollectionInterval = time.Duration(configFile.DataCollectionInterval) * time.Second
	Config.BackupInterval = time.Duration(configFile.BackupInterval) * time.Second
	Config.UVS232URL = configFile.UVS232URL
	Config.MeterURL = configFile.MeterURL
	Config.DataFile = configFile.DataFile
	Config.Webserver.Port = configFile.WebServer.Port
	for s, b := range configFile.WebServer.Webservices {
		Config.Webserver.Webservices[s] = b
	}
}

func getDebugConfig(f config.Flag, d yamlDebug) (writer io.WriteCloser, flag int, err error) {
	var fileName, flagString string

	if s := f.String("debugFile"); s != "" {
		fileName = s
	} else {
		fileName = d.File
	}

	if s := f.String("debugFlag"); s != "" {
		flagString = s
	} else {
		flagString = d.Flag
	}

	// defines Debug section of global.Config
	switch flagString {
	case "trace":
		flag = debug.Full
	case "debug":
		flag = debug.Warning | debug.Info | debug.Error | debug.Fatal | debug.Debug
	case "standard":
		flag = debug.Standard
	}

	switch fileName {
	case "stderr":
		writer = os.Stderr
	case "stdout":
		writer = os.Stdout
	default:
		if writer, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			return
		}
	}

	return
}

func readConfigFile(fileName string, c *yamlStruct) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(c); err != nil {
		return err
	}

	return nil
}
