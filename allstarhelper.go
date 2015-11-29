package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"os"
)

func main() {
	// Note at this point only WARN or above is actually logged to file, and ERROR or above to console.
	jww.SetLogFile("allstarhelper.log")

	// Read config file
	conf, err := getConfigFromFile()
	if err != nil {
		jww.FATAL.Println("Configuration Error:", err)
		os.Exit(0)
	}

	for _, gaugeConf := range conf.USGSRiver.Gauges {
		// {"id": 12141300, "friendlyName":"Middle Fork near Valley Camp", "cmdCode":"*961" },
		gaugeConf.Handle(&conf)
	}

}
