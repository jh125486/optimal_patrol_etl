package main

import (
	"encoding/csv"
	"flag"
	// "encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

const elementsPerDay = 2500

var apiKey string
var inFile string
var outFile string

type centroid struct {
	latitude  float64
	longitude float64
	weight    uint64
}
type centroids map[uint64]*centroid
type graph map[uint64]map[uint64]distanceDuration

type distanceDuration struct {
	distance maps.Distance
	duration time.Duration
}

func init() {
	flag.StringVar(&apiKey, "apiKey", "", "Google API Key")
	flag.StringVar(&inFile, "input", "in.csv", "Input (centroids) CSV filename")
	flag.StringVar(&outFile, "output", "out.csv", "Output (graph nodes) CSV filename")
	flag.Parse()
}

func main() {
	// read from os.Stdin
	// create array of node centroids indexed from one
	// calculate total number of API hits -> bail if over limit and 'free' flag not set
	// roll through each permutation -> get distance, save to graph
	// ouput graph to os.Stdout
	
	centroids := make(centroids)
	graph := make(graph)
	
	centroids.load(inFile)
	ids := centroids.sortedIDs()

	var n1, n2 uint64
	for i := range ids {
		n1 = ids[i]
		graph[n1] = make(map[uint64]distanceDuration)
		for j := i + 1; j < len(ids); j++ {
			n2 = ids[j]
			if dd, err := getDistance(centroids[n1], centroids[n2]); err == nil {
				graph[n1][n2] = dd
			} else {
				fmt.Println("Duration failed")
			}
			if dd, err := getDistance(centroids[n2], centroids[n1]); err == nil {
				graph[n2][n1] = dd
			} else {
				fmt.Println("Duration failed")
			}
		}
	}

	graph.write(outFile)
}

func (c centroids) sortedIDs() uint64Slice {
	ids := make(uint64Slice, 0)
	for id := range c {
		ids = append(ids, id)
	}
	ids.Sort()
	return ids
}

type uint64Slice []uint64

func (s uint64Slice) Len() int           { return len(s) }
func (s uint64Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s uint64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s uint64Slice) Sort()              { sort.Sort(s) }

func (c centroids) load(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 4
	//remove header line
	reader.Read()
	var latitude, longitude float64
	var weight uint64
	id := uint64(1)
	for {
		if record, err := reader.Read(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("fatal error: %s", err)
		} else {
			latitude, _ = strconv.ParseFloat(record[2], 64)
			longitude, _ = strconv.ParseFloat(record[1], 64)
			weight, _ = strconv.ParseUint(record[3], 10, 64)
			c[id] = &centroid{
				latitude:  latitude,
				longitude: longitude,
				weight:    weight,
			}
		}
		id++
	}
}

func (g graph) write(filename string) {
	// records := make([][]string, 0)
	var records [][]string
	records = append(records, []string{"N1", "N2", "Time (s)", "Distance (m)"})
	for n1, nodes := range g {
		for n2, dd := range nodes {
			// record := make([]string, 0)
			var record []string
			record = append(record, fmt.Sprint(n1))
			record = append(record, fmt.Sprint(n2))
			record = append(record, fmt.Sprint(dd.duration.Seconds()))
			record = append(record, fmt.Sprint(dd.distance.Meters))
			records = append(records, record)
		}
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
} 

func (c *centroid) location() string {
	return fmt.Sprintf("%f, %f", c.latitude, c.longitude)
}

func getDistance(c1, c2 *centroid) (dd distanceDuration, err error) {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	r := &maps.DistanceMatrixRequest{
		Origins:      []string{c1.location()},
		Destinations: []string{c2.location()},
	}
	resp, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	if resp.Rows[0].Elements[0].Status == "OK" {
		dd.distance = resp.Rows[0].Elements[0].Distance
		dd.duration = resp.Rows[0].Elements[0].Duration
		return dd, err
	}

	error := fmt.Sprintln("Distance Matrix returned:", resp.Rows[0].Elements[0].Status)
	return dd, errors.New(error)
}
