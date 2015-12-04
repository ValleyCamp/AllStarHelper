package main

import (
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"io"
	"os"
)

// writeOutputConfFileForConfiguration takes a configuration object parsed from our app's json configuration
// and writes a .conf file which can be imported into rpt.conf.
// This file will completely destroy and re-write the .conf file, rather than try and validate it.
func writeOutputConfFileForConfiguration(c Configuration) {
	jww.INFO.Println("Writing allstarhelper.conf file...")

	filename := fmt.Sprintf("%s/allstarhelper_cmdTree.conf", c.Settings.RelativeOutputDir)
	f, err := os.Create(filename)
	if err != nil {
		jww.CRITICAL.Println("Could not create", filename)
		return
	}
	defer f.Close()

	for _, gaugeConf := range c.USGSRiver.Gauges {
		l := fmt.Sprintf("%s=playback,/etc/asterisk/custom/allstarhelper_output/%d\n", gaugeConf.CmdCode.GetForConf(), gaugeConf.Id)
		_, err = io.WriteString(f, l)
	}

	for _, wxunderStationConf := range c.WXUnderground.Stations {
		l := fmt.Sprintf("%s=playback,/etc/asterisk/custom/allstarhelper_output/%s\n", wxunderStationConf.CmdCode.GetForConf(), wxunderStationConf.Id)
		_, err = io.WriteString(f, l)
	}

	if err != nil {
		jww.CRITICAL.Println("Some conf data could not be written to", filename, ". File may be malformed!")
	}
}

// ensureRptConfImport takes a gander at asterisk's rpt.conf file to make sure that it's set up to import
// our auto-generated allstarhelper.conf file
func ensureRptConfImport(path string) {

}
