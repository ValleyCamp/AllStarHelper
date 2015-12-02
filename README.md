# AllStar Helper
This tool converts data from various data sources to text, and then converts that text to voice files which can be played by an [AllStarLink](https://www.allstarlink.org/) node

## Usage:
Put the compiled binary and the config file in your /etc/asterisk/custom/ folder. Run the app.
(Also: Set up a CRON job to run the app every so often, and update your asterisk rpt.conf file to utilize the output generated.)

## Data Sources:
Currently supported data sources are:
* USGS River Gauging Stations
* WeatherUnderground
* [WeatherMoss](https://github.com/ValleyCamp/WeatherMoss/) (Our own weather station API)

## Configuration File
For an example configuration file see allstarhelper_config.json


## Phone Tree:
* 900 -> "To list available weather stations press *920. To list available river gauging stations press *960"
* 920 -> "Available weather stations are: ..." [Read list from configuration file]
* 960 -> "Available river gauging stations are: ..." [Read list from configuration file]
