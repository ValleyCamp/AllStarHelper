package main

import (
	"encoding/json"
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"net/http"
	"strconv"
	"time"
)

type WXUGRes struct {
	Response           WXUGRResResponse           `json:"response"`
	CurrentObservation WXUGRResCurrentObservation `json:"current_observation"`
}

type WXUGRResResponse struct {
	Version string                `json:"version"`
	Error   WXUGRResResponseError `json:"error"`
}

type WXUGRResResponseError struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type WXUGRResCurrentObservation struct {
	StationId        string  `json:"station_id"`
	LocalEpoch       string  `json:"local_epoch"`
	TempF            float64 `json:"temp_f"`
	ObservationEpoch string  `json:"observation_epoch"`
	RelativeHumidity string  `json:"relative_humidity"`
	FeelsLikeF       string  `json:"feelslike_f"`
	DewPointF        float64 `json:"dewpoint_f"`
	WindDir          string  `json:"wind_dir"`
	WindMph          float64 `json:"wind_mph"`
	WindString       string  `json:"wind_string"`
	PrecipOneHrInch  string  `json:"precip_1hr_in"`
	PrecipTodayInch  string  `json:"precip_today_in"`
}

// buildStationFetchUrl builds the WUnderground fetch URL to be used for fetching data.
// This should be abstracted further if we ever need to support multiple URL types, or different data formats...
func buildStationFetchUrl(apikey string, stationid string) string {
	str := fmt.Sprintf("http://api.wunderground.com/api/%s/conditions/q/pws:%s.json", apikey, stationid)
	return str
}

// getTextForWXUnderStation MUST return a text string, which is what will be read aloud as the text for this gauge.
// It may be an error message or the actual result string, but it must be able to be read by the speech synthesizer.
func getTextForWXUnderStation(station *WXUndergroundStationConf, configuration *Configuration) string {
	fetchUrl := buildStationFetchUrl(configuration.WXUnderground.ApiKey, station.Id)
	jww.DEBUG.Println("Station", station.Id, "fetching from URL:", fetchUrl)
	resp, err := http.Get(fetchUrl)
	if err != nil {
		jww.CRITICAL.Println("Station", station.Id, " threw errror fetching station data. Error was: ", err)
		return fmt.Sprintf("Could not fetch data for station at %s.", station.FriendlyName)
	}
	defer resp.Body.Close()

	// $SED -n 's/.*temp_f":\([^0-9.]*\)/ the temperature was \1/p;s/.*relative_humidity":"\(.*\)\%"/ degrees.<break strength="strong"\/> The humidity was \1/p;s/.*wind_dir":"\(.*\)"/ percent.<break strength="strong"\/> The wind direction was \1/p;s/.*wind_mph":\([^0-9.]*\)/.<break strength="strong"\/> The wind speed was \1/p;s/.*wind_gust_mph":"\(.*\)"/ miles per hour.<break strength="strong"\/> Wind gusts were \1/p;s/.*pressure_in":"\(.*\)"/ miles per hour.<break strength="strong"\/> The pressure was \1/p;s/.*dewpoint_f":\([^0-9.]*\)/ inches.<break strength="strong"\/> The dew point was \1/p;s/.*precip_today_in":"\(.*\)"/ degrees.<break strength="strong"\/> The precipitation today was \1/p'  | tr -d '\n' )

	apiResult := WXUGRes{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&apiResult)
	if jsonErr != nil {
		jww.CRITICAL.Println("Station", station.Id, "could not parse JSON response. Error was: ", jsonErr)
		return fmt.Sprintf("Could not parse data for station at %s.", station.FriendlyName)
	}

	if apiResult.Response.Error.Type != "" {
		jww.CRITICAL.Println("Weather Underground API Error. Errortype:", apiResult.Response.Error.Type, ", Description:", apiResult.Response.Error.Description)
		return fmt.Sprintf("Could not fetch data from Weather Underground.")
	}

	jww.INFO.Println("Station", station.Id, "Parsed Result:", apiResult)

	c := apiResult.CurrentObservation

	// Convert the UNIX Epoch time the api gave us from a string of numbers to a format we want
	timeInt, err := strconv.ParseInt(c.ObservationEpoch, 10, 64)
	observationTime := time.Unix(timeInt, 0)
	timeStr := observationTime.Format("January 02 at 03:04 PM")

	// Munge the JSON from weather underground which is returning floats as strings and such into sensible values.
	hourPrecipInches, err := strconv.ParseFloat(c.PrecipTodayInch, 64)
	if err != nil {
		jww.CRITICAL.Println("Weather Underground returned a value in the precip_1hr_in field that we weren't expecting!")
		return "Error in Weather Underground data."
	}

	dayPrecipInches, err := strconv.ParseFloat(c.PrecipTodayInch, 64)
	if err != nil {
		jww.CRITICAL.Println("Weather Underground returned a value in the precip_today_in field that we weren't expecting!")
		return "Error in Weather Underground data."
	}

	// Construct final output string
	outStr := fmt.Sprintf("On %s the temperature at %s was %0.1f degrees farenheit, with a relative humidity of %s. Percieved temperature was %s degrees farenheit, dewpoint of %0.1f. Winds were %s. Precipitation for the day was %0.1f inches, with %0.1f inches in the last hour. End of report.", timeStr, station.FriendlyName, c.TempF, c.RelativeHumidity, c.FeelsLikeF, c.DewPointF, c.WindString, hourPrecipInches, dayPrecipInches)
	jww.DEBUG.Println("Station", station.Id, "Got final string: ", outStr)

	return outStr
}
