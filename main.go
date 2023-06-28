package main

import (
	"fmt"
	"math"
	"sort"
)

const StatePlaneCoordinateZone string = "0405"

func main() {

	//c := Coordinate{6480464.803, 1830021.816}

	//c.ToLatLong("FEET")

	crimes, count := ProcessCrimes()

	fmt.Printf("Valid crimes processed: %d (out of %d)\n", len(crimes), count)
	crimes.writeToCSV(nil)
	for hour := uint8(0); hour < 24; hour++ {
		crimes.writeToCSV(&hour)
	}

	hours := make(map[uint8]int)
	weekdays := make(map[uint8]int)
	categories := make(map[string]int)
	gangRelated := make(map[bool]int)
	for _, crime := range crimes {
		hours[crime.hour]++
		weekdays[crime.weekday]++
		categories[crime.category]++
		if crime.gang {
			gangRelated[true]++
		} else {
			gangRelated[false]++
		}
	}
	fmt.Println("Hourly distribution:")
	histogramUint8(hours)

	fmt.Println("Weekday distribution:")
	histogramUint8(weekdays)

	fmt.Printf("Category counts (%d):\n", len(categories))
	var keys []string
	for key := range categories {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, categories[key])
	}

	fmt.Printf("\nGang related %d (out of %d)\n", gangRelated[true], gangRelated[true] + gangRelated[false])
}

func histogramUint8(data map[uint8]int) {
	largest := 0.0
	for _, v := range data {
		if float64(v) > largest {
			largest = float64(v)
		}
	}
	ratio := 10.0 / largest

	var keys []int
	for key := range data {
		keys = append(keys, int(key))
	}
	sort.Ints(keys)

	for _, key := range keys {
		padding := float64(data[uint8(key)]) * ratio
		fmt.Printf("%02d: %*v\n", key, Round(padding), "*")
	}
}

func Round(f float64) int {
	return int(math.Floor(f + .5))
}
