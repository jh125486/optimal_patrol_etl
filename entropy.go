package main

import (
	"encoding/csv"
	// "encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"math"
	"strconv"
)

type Centroid struct {
	id        uint64
	latitude  float64
	longitude float64
	weight    uint64
	patrol 	  bool
	entropy   float64
}
type Centroids map[uint64]*Centroid

type Edge struct {
	time     uint64
	distance uint64
}
type Graph map[uint64]map[uint64]*Edge

var systemWeight, quickestPath uint64

const tResponse = float64(120)

var graph Graph
var centroids Centroids

func main() {

	centroids = loadCentroids("../results/centroids.csv")
	graph = loadGraph("../results/graph.csv")
	
	for _, c := range centroids {
		systemWeight += c.weight
	}
	quickestPath = centroids.quickestPathInSystem()

	var totalEntopy float64
	count := 0

	for _, c1 := range centroids {
		c1.entropy = c1.totalEntropy()
		for _, c2 := range centroids {
			if c1 == c2 {
				continue
			}
			totalEntopy += entropy(c1, c2)
			count++
		}
	}

	for _, c1 := range centroids {
		fmt.Printf("%v, %v, %v, %v\n", c1.id, c1.latitude, c1.longitude, c1.entropy)
	}
}

func getCentroid(id uint64) *Centroid {
	return centroids[id]
}

func (c1 *Centroid) totalEntropy() float64 {
	var cEntropy float64
	for n2 := range graph[c1.id] {
		cEntropy += entropy(c1, getCentroid(n2))
	}
	return cEntropy
}

func (c1 Centroid) quickestPath() uint64 {
	paths := make(Uint64Slice, 0)
	for _, edge := range graph[c1.id] {
		paths = append(paths, edge.time)
	}
	paths.Sort()
	
	return paths[0]
} 

func (c Centroids) quickestPathInSystem() uint64 {
	paths := make(Uint64Slice, 0)
	for _, c2 := range graph {
		for _, edge := range c2 {
			paths = append(paths, edge.time)
		}
	}
	paths.Sort()
	
	return paths[0]
}

func entropy(c1, c2 *Centroid) float64 {
	pathHyper := float64(c1.quickestPath()) / float64(quickestPath)
	weightHyper := float64(c1.weight) / float64(systemWeight)
	
	info := pathHyper + weightHyper

	entropy := -(info/2) * math.Log(info/2)
	return entropy
}

func (c Centroids) sortedIDs() Uint64Slice {
	ids := make(Uint64Slice, 0)
	for id := range c {
		ids = append(ids, id)
	}
	ids.Sort()
	return ids
}

type Uint64Slice []uint64

func (s Uint64Slice) Len() int           { return len(s) }
func (s Uint64Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s Uint64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Uint64Slice) Sort()              { sort.Sort(s) }

func loadCentroids(filename string) Centroids {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	defer file.Close()

	centroids := make(Centroids)

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 4
	//remove header line
	reader.Read()
	var id, weight uint64
	var latitude, longitude float64
	for {
		if record, err := reader.Read(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("fatal error: %s", err)
		} else {
			id, _ = strconv.ParseUint(record[0], 10, 64)
			longitude, _ = strconv.ParseFloat(record[1], 64)
			latitude, _ = strconv.ParseFloat(record[2], 64)
			weight, _ = strconv.ParseUint(record[3], 10, 64)
			centroids[id] = &Centroid{
				id:        id,
				latitude:  latitude,
				longitude: longitude,
				weight:    weight,
			}
		}
	}
	return centroids
}

func loadGraph(filename string) Graph {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	defer file.Close()

	graph := make(Graph)

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 4
	//remove header line
	reader.Read()
	var n1, n2, time, distance uint64
	for {
		if record, err := reader.Read(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("fatal error: %s", err)
		} else {
			n1, _ = strconv.ParseUint(record[0], 10, 64)
			n2, _ = strconv.ParseUint(record[1], 10, 64)
			time, _ = strconv.ParseUint(record[2], 10, 64)
			distance, _ = strconv.ParseUint(record[3], 10, 64)
			edge := &Edge{
				time:     time,
				distance: distance,
			}
			if graph[n1] == nil {
				graph[n1] = make(map[uint64]*Edge)
			}
			graph[n1][n2] = edge
			if graph[n2] == nil {
				graph[n2] = make(map[uint64]*Edge)
			}
			graph[n2][n1] = edge
		}
	}
	return graph
}

func (c *Centroid) location() string {
	return fmt.Sprintf("%f, %f", c.latitude, c.longitude)
}
