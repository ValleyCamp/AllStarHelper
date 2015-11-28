# AllStar Helper
This tool converts data from various data sources to text, and then converts that text to voice files which can be played by an [AllStarLink](https://www.allstarlink.org/) node

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


## TODO:
* Write script to export our config file into AllStar link rpt.conf syntax for easy copy-paste
