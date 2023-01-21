package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type distance struct {
	StartName string
	EndName   string
	Distance  float64
}

type coordinate struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type apiResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Geometry string `json:"geometry"`
		Legs     []struct {
			Steps    []interface{} `json:"steps"`
			Summary  string        `json:"summary"`
			Weight   float64       `json:"weight"`
			Duration float64       `json:"duration"`
			Distance float64       `json:"distance"`
		} `json:"legs"`
		WeightName string  `json:"weight_name"`
		Weight     float64 `json:"weight"`
		Duration   float64 `json:"duration"`
		Distance   float64 `json:"distance"`
	} `json:"routes"`
	Waypoints []struct {
		Hint     string    `json:"hint"`
		Distance float64   `json:"distance"`
		Name     string    `json:"name"`
		Location []float64 `json:"location"`
	} `json:"waypoints"`
}

func main() {
	file, err := os.Create("distances.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	startFile, err := os.Open("./start.csv")
	if err != nil {
		log.Fatal(err)
	}
	startLocations, err := getLocations(startFile)
	if err != nil {
		log.Fatal(err)
	}
	endFile, err := os.Open("./end.csv")
	if err != nil {
		log.Fatal(err)
	}
	endLocations, err := getLocations(endFile)
	if err != nil {
		log.Fatal(err)
	}

	for idx, startLocation := range startLocations {
		fmt.Printf("Processing starting location number %d...\n", idx+1)
		distances := getDistances(startLocation, endLocations)
		writeResults(distances)
	}
	fmt.Print("Done!\n")
}

func getLocations(file *os.File) ([]coordinate, error) {
	var coordinates []coordinate
	reader := csv.NewReader(file)

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		lat, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			return nil, err
		}
		long, err := strconv.ParseFloat(line[2], 64)
		if err != nil {
			return nil, err
		}
		coordinate := coordinate{
			Name:      line[0],
			Latitude:  lat,
			Longitude: long,
		}
		coordinates = append(coordinates, coordinate)
	}
	return coordinates, nil
}

func getDistances(startLocation coordinate, endLocations []coordinate) []distance {
	var distances []distance
	for idx, location := range endLocations {

		fmt.Printf("Getting distance of location %d/%d...\n", idx+1, len(endLocations))

		url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f", startLocation.Longitude, startLocation.Latitude, location.Longitude, location.Latitude)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("resp : %s\n", body)
			log.Fatalln(err)
		}

		var apiResponse apiResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			fmt.Println(err)
		}
		var distance distance
		distance.StartName = startLocation.Name
		distance.EndName = location.Name
		distance.Distance = apiResponse.Routes[0].Distance

		distances = append(distances, distance)
	}
	return distances
}

func writeResults(distances []distance) {

	f, err := os.OpenFile("distances.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	w := csv.NewWriter(f)
	for _, j := range distances {
		w.Write([]string{j.StartName, j.EndName, fmt.Sprintf("%f", j.Distance)})
	}
	w.Flush()
}
