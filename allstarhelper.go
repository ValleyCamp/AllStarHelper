package main

import (
	"fmt"
	//"github.com/joshproehl/goflite"
	"flag"
	jww "github.com/spf13/jwalterweatherman"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	flgVerbose := flag.Bool("verbose", false, "Output additional debugging information to both STDOUT and the log file")
	flgWriteConf := flag.Bool("writeconf", false, "Write out configuration file to be imported into app_rpt.conf and exit")
	flag.Parse()

	// Note at this point only WARN or above is actually logged to file, and ERROR or above to console.
	jww.SetLogFile("allstarhelper.log")

	if *flgVerbose {
		jww.SetLogThreshold(jww.LevelDebug)
		jww.SetStdoutThreshold(jww.LevelInfo)
	} else {
		// Set default logging verbosity.
		jww.SetLogThreshold(jww.LevelWarn)
		jww.SetStdoutThreshold(jww.LevelError)
	}

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

	if *flgWriteConf {
		// write out the allstarhelper_cmdTree.conf file
		writeOutputConfFileForConfiguration(appconf)
		return
	}

	// Dispatch a thread to handle each of the gauges from the conf file
	gaugeDone := make(chan bool)
	for _, gaugeConf := range appconf.USGSRiver.Gauges {
		jww.DEBUG.Println("Dispatching Handler gauge for conf:", gaugeConf)
		go func(curConf USGSGaugeConf) {
			gaugeId := strconv.Itoa(curConf.Id)
			gaugeRes := getTextForGauge(&curConf, &appconf) // Not copying appConf as we never change it... TODO: Make actually thread safe.
			writeOutputTextFile(appconf.Settings.RelativeOutputDir, gaugeId, gaugeRes)

			// We wrote the txt file for reference, but we're going to go ahead and just output directly to wave now.
			writeOutputAudioFile(appconf.Settings.RelativeOutputDir, gaugeId, gaugeRes, "slt")

			// Now we have to convert the file. Because Asterisk.
			convertOutputAudioFileForAsterisk(appconf.Settings.RelativeOutputDir, gaugeId)

			// Let function that spun up the goroutine know that one of the threads is done
			gaugeDone <- true
		}(gaugeConf)
	}

	// Dispatch a thread to handle each of the wxunderground stations from the conf file
	wxunderStationDone := make(chan bool)
	for _, wxunderStationConf := range appconf.WXUnderground.Stations {
		jww.DEBUG.Println("Dispatching Handler for WXUnderground station for conf:", wxunderStationConf)
		go func(curConf WXUndergroundStationConf) {
			stationRes := getTextForWXUnderStation(&curConf, &appconf) // Not copying appConf as we never change it... TODO: Make actually thread safe.
			writeOutputTextFile(appconf.Settings.RelativeOutputDir, curConf.Id, stationRes)

			// We wrote the txt file for reference, but we're going to go ahead and just output directly to wave now.
			writeOutputAudioFile(appconf.Settings.RelativeOutputDir, curConf.Id, stationRes, "slt")

			// Now we have to convert the file. Because Asterisk.
			convertOutputAudioFileForAsterisk(appconf.Settings.RelativeOutputDir, curConf.Id)

			// Let function that spun up the goroutine know that one of the threads is done
			wxunderStationDone <- true
		}(wxunderStationConf)
	}

	// wait until all gauges and stations are done processing before we exit
	for _ = range appconf.USGSRiver.Gauges {
		<-gaugeDone
	}
	for _ = range appconf.WXUnderground.Stations {
		<-wxunderStationDone
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

// writeOutputWaveFile writes {path}/{id}.wave with the contents of OutStr, overwriting any existing file.
// The wave file is generated using the flite TTS engine, using the voice file described by useVoice.
// TODO: Handle creating some sort of fallback wave file in case we couldn't generate one? (So the AllStar user knows?)
func writeOutputAudioFile(path string, id string, outStr string, useVoice string) {
	// For now we're going to do away with trying to build flite into the binary and just call the compiled flite binary on the system
	args := []string{"-f", fmt.Sprintf("%s/%s.txt", path, id), "-o", fmt.Sprintf("%s/%s.source.wav", path, id), "-voice", useVoice}
	res, err := exec.Command("flite", args...).Output()
	if err != nil {
		jww.CRITICAL.Println("Could not create wav file for", id, "-- Error was:", err)
	}
	jww.INFO.Println("flite returned the following for", id, ":", res)

	/*
		// Create the wavform
		wav, wgErr := goflite.TextToWave(outStr, useVoice)
		if wgErr != nil {
			jww.CRITICAL.Println("Could not synthesize wav for ", id, ": ", wgErr)
			return
		}

		// create the output writer
		waveName := fmt.Sprintf("%s/%s.wav", path, id)
		f, fErr := os.Create(waveName)
		if fErr != nil {
			jww.CRITICAL.Println("Could not create output wave file for", id)
			return
		}
		defer f.Close() //No mater which branching path we take we want to make sure the handle closes

		// write the output waveform to the file we've opened.
		wwErr := wav.EncodeRIFF(f)
		if wwErr != nil {
			jww.CRITICAL.Println("Could not write wave for", id, ": ", wwErr)
			return
		}
		f.Sync() // Just to be sure we're done writing

		// TODO: Convert from WAV to format used by Asterisk?
	*/
}

// convertOutputAudioFileForAsterisk takes the {id}.source.wav file generated by flite and coneverts it to a
// format that Asterisk will be able to play, removing the source.wav file once this is done.
func convertOutputAudioFileForAsterisk(path string, id string) {
	soxArgs := []string{fmt.Sprintf("%s/%s.source.wav", path, id), "-r", "8k", "-c", "1", fmt.Sprintf("%s/%s.wav", path, id)}
	convRes, convErr := exec.Command("sox", soxArgs...).Output()
	if convErr != nil {
		jww.CRITICAL.Println("Could not convert wav file for", id, "-- Error was:", convErr, "-- ", convRes)
		return
	}

	rmArgs := []string{fmt.Sprintf("%s/%s.source.wav", path, id)}
	rmRes, rmErr := exec.Command("rm", rmArgs...).Output()
	if rmErr != nil {
		jww.CRITICAL.Println("Could not remove source wav file for", id, "-- Error was:", rmErr, "-- ", rmRes)
		return
	}
}
