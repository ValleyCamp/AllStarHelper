package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Settings      AppSettings         `json:"settings"`
	USGSRiver     USGSRiverConfig     `json:"usgsriver"`
	WXUnderground WXUndergroundConfig `json:"wxunder"`
	WeatherMoss   WeatherMossConfig   `json:"weathermoss"`
}

type AppSettings struct {
	RelativeOutputDir string `json:"relative_outputdir"`
}

type USGSRiverConfig struct {
	Gauges       []USGSGaugeConf `json:"gauges"`
	CmdCodeAbout string          `json:"cmdCodeAbout"`
}

type USGSGaugeConf struct {
	Id           int    `json:"id"`
	FriendlyName string `json:"friendlyName"`
	CmdCode      string `json:"cmdCode"`
}

type WXUndergroundConfig struct {
	ApiKey       string                     `json:"api_key"`
	CmdCodeAbout string                     `json:"cmdCodeAbout"`
	Stations     []WXUndergroundStationConf `json:"stations"`
}

type WXUndergroundStationConf struct {
	Id           string `json:"id"`
	FriendlyName string `json:"frinedlyName"`
	CmdCode      string `json:"cmdCode"`
}

type WeatherMossConfig struct {
}

func getConfigFromFile() (Configuration, error) {
	file, _ := os.Open("allstarhelper_config.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)

	return configuration, err
}
