package main

import (
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type USGSGaugeDataRow struct {
	CubicFeetPerSecond float64 // Maps to USGS parameter 00060 (Discharge, cubic feet per second)
	GaugeHeight        float64 // Maps to USGS parameter 00065 (Gage height, feet)
}

// The data source URL. Take this, append the gaugeID, and that's the URL to get gauge data.
const gaugeUrl = "http://waterdata.usgs.gov/wa/nwis/uv?cb_all_00060_00065=on&cb_00060=on&cb_00065=on&format=rdb&period=1&site_no="

// Handle
func handleGauge(gauge *USGSGaugeConf, configuration *Configuration) {
	// Make the HTTP request
	resp, err := http.Get(fmt.Sprintf("%s%d", gaugeUrl, gauge.Id))
	if err != nil {
		jww.CRITICAL.Println("Error fetching data for gauge", gauge.Id, ". Error was: ", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		jww.CRITICAL.Println("Error reading response body for gauge", gauge.Id, ". Error was: ", err)
		return
	}

	// Get the text file returned by the server in a line-by-line format we can look through.
	bodyTxt := string(body)
	bodyLines := strings.Split(bodyTxt, "\n")

	// Run through the whole file by lines. For now we're just going to keep going until we get the very last row. (The only data we care about is the most recent)
	headerRow := ""
	lastRow := ""
	for _, rowData := range bodyLines {
		// Look for the header row. It'll be the first non-commented row we find. Once we find it we can stop looking.
		if headerRow == "" && len(rowData) > 9 && rowData[:9] == "agency_cd" {
			headerRow = rowData
		}

		if len(rowData) > 5 && rowData[:4] == "USGS" {
			lastRow = rowData
		}
	}

	// Now we find out which columns in the tab-delimited file are the ones we want by parsing the header row
	datetimeCol, cfpsCol, gaugeHeightCol := -1, -1, -1
	headers := strings.Split(headerRow, "\t")
	for index, header := range headers {
		if header == "datetime" {
			datetimeCol = index
		}

		// this column header wasn't datetime, and we know that any column we're interested in will have a _ in it,
		// so we'll split that first, ignoring any columns that might not be what we want.
		headerSplit := strings.Split(header, "_")
		if len(headerSplit) > 1 {
			switch headerSplit[len(headerSplit)-1] {
			case "00060":
				cfpsCol = index
			case "00065":
				gaugeHeightCol = index
			}
		}
	}

	// Sanity check to make sure valid columns found
	if datetimeCol == -1 || cfpsCol == -1 || gaugeHeightCol == -1 {
		jww.CRITICAL.Println("Gauge", gauge.Id, "could not find valid data on any row... Aborting for this gauge.")
		return
	}

	// Now that we know which columns we're looking for yank the data out of there.
	splitLastRow := strings.Split(lastRow, "\t")
	t, err := time.Parse("2006-01-02 15:04", splitLastRow[datetimeCol])
	if err != nil {
		jww.CRITICAL.Println("Gauge", gauge.Id, "could not parse time, aborting for gauge. Error was: ", err)
		return
	}

	// Format our final output TXT!
	timeStr := t.Format("January 02 at 03:04 PM") // Define what we want our output time string to look like
	outStr := fmt.Sprintf("At %s the gauging station on the %s reported %s cubic feet per second, at a height of %s feet.", timeStr, gauge.FriendlyName, splitLastRow[cfpsCol], splitLastRow[gaugeHeightCol])
	jww.DEBUG.Println("Gauge", gauge.Id, "generated text: ", outStr)

	// Write out our result to {OUTPUTDIR}/{GAUGE_ID}.txt
	filename := fmt.Sprintf("%d.txt", gauge.Id)
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", configuration.Settings.RelativeOutputDir, filename), []byte(outStr), 0644)
	if err != nil {
		jww.CRITICAL.Println("Gauge", gauge.Id, "could not write to output file.")
		return
	}
}

/*
// Convert "2" to "2nd", etc...
func ordinalForDayNumber(x int) string {
	suffix := "th"
	switch x {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}
	return strconv.Itoa(x) + suffix
}
*/
