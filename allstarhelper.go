package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"os"
)

func main() {
	// Note at this point only WARN or above is actually logged to file, and ERROR or above to console.
	jww.SetLogFile("allstarhelper.log")

	// Read config file
	appconf, err := getConfigFromFile()
	if err != nil {
		jww.FATAL.Println("Configuration Error:", err)
		os.Exit(0)
	}

	gaugeDone := make(chan bool)
	for _, gaugeConf := range appconf.USGSRiver.Gauges {
		jww.DEBUG.Println("Dispatching Handler gauge for conf:", gaugeConf)
		go func(curConf USGSGaugeConf) {
			handleGauge(&curConf, &appconf) // Not copying appConf as we never change it... TODO: Make actually thread safe.
			gaugeDone <- true
		}(gaugeConf)
	}

	// wait until all gauges and stations are done processing before we exit
	for _ = range appconf.USGSRiver.Gauges {
		<-gaugeDone
	}

	jww.INFO.Println("Done. Exiting!")
}
