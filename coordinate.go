package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

type Coordinate struct {
	x float64
	y float64
}

var latLongRegex *regexp.Regexp

func init() {
	latLongRegex = regexp.MustCompile(`LATITUDE        LONGITUDE         AREA\n DD MM SS.sssss  DDD MM SS.sssss       \n --------------  ---------------   ----\n (?P<latDD>\d+) +(?P<latMM>\d+) +(?P<latSS>\d+\.\d+) +(?P<longDD>\d+)+ (?P<longMM>\d+) +(?P<longSS>\d+\.\d+)`)
}

func (c Coordinate) ToLatLong(datumType string) (string, string) {
	apiUrl := "http://www.ngs.noaa.gov"
	resource := "/cgi-bin/spc_getgp.prl"
	data := url.Values{}
	if datumType == "FEET" {
		data.Add("DatumSelected", "NAD27 (Input SPC units = FEET)")
	} else if datumType == "METERS" {
		data.Add("DatumSelected", "NAD83 (Input SPC units = METERS)")
	}
	data.Add("NorthBox", strconv.FormatFloat(c.x, 'f', -1, 64))
	data.Add("EastBox", strconv.FormatFloat(c.y, 'f', -1, 64))
	data.Add("ZoneBox", StatePlaneCoordinateZone)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(r)
	fmt.Println(resp.Status)

	defer resp.Body.Close()
	if contents, err := ioutil.ReadAll(resp.Body); err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		lat, long := getLatLongFromResponse(string(contents))

		fmt.Printf("%v\t%v\n", lat, long)
	}

	return "a", "b"
}

func getLatLongFromResponse(body string) (float64, float64) {
	match := latLongRegex.FindStringSubmatch(body)
	result := make(map[string]string)
	for i, name := range latLongRegex.SubexpNames() {
		if i != 0 {
			result[name] = match[i]
		}
	}

	latitude := convertDMStoDecimal(result["latDD"], result["latMM"], result["latSS"])
	longitude := convertDMStoDecimal(result["longDD"], result["longMM"], result["longSS"])

	return latitude, longitude
}

func convertDMStoDecimal(dd string, mm string, ss string) float64 {
	d, _ := strconv.ParseFloat(dd, 64)
	m, _ := strconv.ParseFloat(mm, 64)
	s, _ := strconv.ParseFloat(ss, 64)

	return d + m/60 + s/3600
}
