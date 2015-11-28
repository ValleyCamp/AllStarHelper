package main

import (
	"fmt"
)

type USGSGaugeDataRow struct {
	CubicFeetPerSecond float64 // Maps to USGS parameter 00060 (Discharge, cubic feet per second)
	GaugeHeight        float64 // Maps to USGS parameter 00065 (Gage height, feet)
}

const gaugeUrl = "http://waterdata.usgs.gov/wa/nwis/uv?cb_all_00060_00065=on&cb_00060=on&cb_00065=on&format=rdb&period=1&site_no="

func (gauge *USGSGaugeConf) Handle() {
	fmt.Println("Dispatching for gauge", gauge.Id)

}
