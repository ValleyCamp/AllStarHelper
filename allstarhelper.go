package main

import (
	"fmt"
)

func main() {
	// Read config file
	conf, err := getConfigFromFile()
	if err != nil {
		fmt.Println("Configuration Error:", err)
	}

	for _, gaugeConf := range conf.USGSRiver.Gauges {
		// {"id": 12141300, "friendlyName":"Middle Fork near Valley Camp", "cmdCode":"*961" },
		gaugeConf.Handle()
	}
}
