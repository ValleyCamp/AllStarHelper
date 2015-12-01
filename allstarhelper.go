package main

import (
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func main() {
	// Note at this point only WARN or above is actually logged to file, and ERROR or above to console.
	jww.SetLogFile("allstarhelper.log")

	// Set default logging verbosity. // TODO: add verbose flag to output more info
	jww.SetLogThreshold(jww.LevelWarn)
	jww.SetStdoutThreshold(jww.LevelError)

	jww.INFO.Println("Starting run at", time.Now().Format("2006-01-02 15:04:05"))

	// Read config file
	appconf, err := getConfigFromFile()
	if err != nil {
		jww.FATAL.Println("Configuration Error:", err)
		os.Exit(0)
	}

	// Before we do anything make sure output directory exists
	err = os.MkdirAll(appconf.Settings.RelativeOutputDir, 0711)
	if err != nil {
		jww.FATAL.Println("could not create output directory. Permissions issue?")
		os.Exit(0)
	}

	// Dispatch a thread to handle each of the gauges from the conf file
	gaugeDone := make(chan bool)
	for _, gaugeConf := range appconf.USGSRiver.Gauges {
		jww.DEBUG.Println("Dispatching Handler gauge for conf:", gaugeConf)
		go func(curConf USGSGaugeConf) {
			gaugeId := strconv.Itoa(curConf.Id)
			gaugeRes := getTextForGauge(&curConf, &appconf) // Not copying appConf as we never change it... TODO: Make actually thread safe.
			writeOutputTextFile(appconf.Settings.RelativeOutputDir, gaugeId, gaugeRes)
			gaugeDone <- true
		}(gaugeConf)
	}

	// wait until all gauges and stations are done processing before we exit
	for _ = range appconf.USGSRiver.Gauges {
		<-gaugeDone
	}

	jww.INFO.Println("Done creating all files, exiting at", time.Now().Format("2006-01-02 15:04:05"))
}

// writeOutputFile writes {path}/{id}.txt with the contents of outStr, overwriting any existing file.
// Although this is being called from a goroutine we don't need ot synchronize as we should never be trying to write
// out to the same file from multiple routines.
func writeOutputTextFile(path string, id string, outStr string) {
	outPath := fmt.Sprintf("%s/%s.txt", path, id)
	jww.DEBUG.Println("Writing file", outPath)
	err := ioutil.WriteFile(outPath, []byte(outStr), 0644)
	if err != nil {
		jww.CRITICAL.Println("Could not write to output file for id:", id)
	}
}
