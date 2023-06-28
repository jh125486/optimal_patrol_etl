package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Crime struct {
	hour     uint8
	weight   uint8
	code     string
	weekday  uint8
	gang     bool
	category string
	location Coordinate
}

type Crimes []Crime

type CrimeCluster struct {
	aggregateWeight uint64
	centroid        Coordinate
}

const timeFormat1 string = "02-Jan-06 15:04:05"
const timeFormat2 string = "1/2/2006 3:04:05 PM"

// visually determined through R histograms
// not verified with lat/long projections
const (
	xLower = 5500000
	xUpper = 7500000
	yLower = 1700000
	yUpper = 1920000
)

var crimeCategories map[string]uint8
var tzLocation *time.Location
var timeFormat string
var wg sync.WaitGroup

func init() {
	tzLocation, _ = time.LoadLocation("America/Los_Angeles")

	if file, err := ioutil.ReadFile("crime_categories.json"); err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	} else {
		if err := json.Unmarshal(file, &crimeCategories); err != nil {
			panic(err)
		}
	}
}

func (_ Crimes) GetHeaders() []string {
	return []string{
		"DoW",
		"Hour",
		"Weight",
		"Gang",
		"X",
		"Y",
	}
}

func (c Crime) ToSlice() []string {
	return []string{
		fmt.Sprint(c.weekday),
		fmt.Sprint(c.hour),
		fmt.Sprint(c.weight),
		fmt.Sprint(c.gang),
		strconv.FormatFloat(c.location.x, 'f', -1, 64),
		strconv.FormatFloat(c.location.y, 'f', -1, 64),
	}
}

func ProcessCrimes() (Crimes, int) {
	csv_files, _ := ioutil.ReadDir("crime_data")

	crimes := make(Crimes, 0)

	filesCh := make(chan string, len(csv_files))
	countCh := make(chan int, len(csv_files))
	resultsCh := make(chan Crime, 100)

	for w := 1; w <= runtime.NumCPU()*3; w++ {
		wg.Add(1)
		go processCrimeFile(filesCh, resultsCh, countCh)
	}

	err := filepath.Walk("crime_data", func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ".csv" {
			filesCh <- path
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	close(filesCh)

	go func() {
		for result := range resultsCh {
			crimes = append(crimes, result)
		}
	}()

	totalCrimes := 0
	go func() {
		for count := range countCh {
			totalCrimes += count
		}
	}()

	wg.Wait()

	return crimes, totalCrimes
}

func processCrimeFile(files <-chan string, resultsCh chan<- Crime, countCh chan<- int) {
	defer wg.Done()
	for filename := range files {
		if file, err := os.Open(filename); err != nil {
			fmt.Println("Error:", err)
		} else {
			defer file.Close()

			reader := csv.NewReader(file)
			reader.FieldsPerRecord = 18
			count := 0
			// remove header line
			reader.Read()
			for {
				if record, err := reader.Read(); err == io.EOF {
					break
				} else if err != nil {
					fmt.Println("Error:", err)
					return
				} else {
					count++
					crimeFromLine(record, resultsCh)
				}
			}
			fmt.Printf("Processed %s: %d crimes\n", filename, count)
			countCh <- count
		}
	}
}

func crimeFromLine(components []string, resultsCh chan<- Crime) {
	// skip any date without a time value
	if len(components[0]) > 9 {
		x, _ := strconv.ParseFloat(components[9], 64)
		y, _ := strconv.ParseFloat(components[10], 64)

		// don't return any Crimes that don't have locations
		coord := Coordinate{x, y}
		t := getTimeFromCrime(components[0])
		if validCoordinate(coord) {
			resultsCh <- Crime{
				hour:     uint8(t.Hour()),
				weekday:  uint8(t.Weekday()),
				code:     components[3],
				gang:     isGangRelated(components[14]),
				location: coord,
				category: components[2],
				weight:   crimeCategories[components[2]],
			}
		}
	}
}

func validCoordinate(c Coordinate) bool {
	return c.x > xLower && c.x < xUpper && c.y > yLower && c.y < yUpper
}

func getTimeFromCrime(s string) time.Time {
	if strings.Contains(s, "-") {
		timeFormat = timeFormat1
	} else if strings.Contains(s, "/") {
		timeFormat = timeFormat2
	}

	t, _ := time.ParseInLocation(timeFormat, s, tzLocation)
	return t
}

func (crimes Crimes) writeToCSV(hour *uint8) {
	os.Mkdir("results", 0755)
	filename := "crimes"
	if hour != nil {
		filename += fmt.Sprintf("_%02d", *hour)
	}

	if csv_file, err := os.Create("results/" + filename + ".csv"); err != nil {
		fmt.Println("Error:", err)
		return
	} else {
		defer csv_file.Close()

		writer := csv.NewWriter(csv_file)
		if err := writer.Write(crimes.GetHeaders()); err != nil {
			fmt.Println("Error:", err)
			return
		}

		for _, crime := range crimes {
			if hour != nil && crime.hour != *hour {
				continue

			}
			if err := writer.Write(crime.ToSlice()); err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
		writer.Flush()
	}

}


func isGangRelated(s string) bool {
	if strings.ToUpper(s) == "YES" || strings.ToUpper(s) == "TRUE" || strings.ToUpper(s) == "Y" {
		return true
	}
	return false
}
