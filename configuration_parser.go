package main

import (
	"encoding/json"
	"os"
)

// Define structs for our configuration file, starting at the top level (whole-file)
// Any errors should be passed back via the err object and handled by the caller

type Configuration struct {
	Settings      AppSettings         `json:"settings"`
	USGSRiver     USGSRiverConfig     `json:"usgsriver"`
	WXUnderground WXUndergroundConfig `json:"wxunder"`
	WeatherMoss   WeatherMossConfig   `json:"weathermoss"`
}

// Type alias for "*xx" command codes
type CmdCode string

type AppSettings struct {
	RelativeOutputDir string `json:"relative_outputdir"`
}

type USGSRiverConfig struct {
	Gauges       []USGSGaugeConf `json:"gauges"`
	CmdCodeAbout CmdCode         `json:"cmdCodeAbout"`
}

type USGSGaugeConf struct {
	Id           int     `json:"id"`
	FriendlyName string  `json:"friendlyName"`
	CmdCode      CmdCode `json:"cmdCode"`
}

type WXUndergroundConfig struct {
	ApiKey       string                     `json:"api_key"`
	CmdCodeAbout CmdCode                    `json:"cmdCodeAbout"`
	Stations     []WXUndergroundStationConf `json:"stations"`
}

type WXUndergroundStationConf struct {
	Id           string  `json:"id"`
	FriendlyName string  `json:"friendlyName"`
	CmdCode      CmdCode `json:"cmdCode"`
}

type WeatherMossConfig struct {
}

// getConfigFromFile does what it says on the box and returns a Configuration object
// representing the config file. TODO: Pass in the filename from parameters/default
func getConfigFromFile() (Configuration, error) {
	file, _ := os.Open("allstarhelper_config.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)

	return configuration, err
}

// Command codes in the JSON are strings like *900. This allows the reader of the JSON file to
// understand what's going on visually, but when it comes to generating conf files we don't
// need the * in front...
func (c CmdCode) GetForConf() string {
	if c[0] == '*' {
		return string(c[1:len(c)])
	} else {
		return string(c)
	}
}
