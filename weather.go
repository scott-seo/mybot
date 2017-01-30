package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"os"

	"github.com/blevesearch/bleve"
)

var cityIndex bleve.Index

func IndexCity() {
	// open a new index

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("city.bleve", mapping)
	if err != nil {
		fmt.Println(err)
		return
	}

	// index some data
	// file, err := os.Open("/Users/seos/src/github.com/scott-seo/mybot/city.list.us.json")
	file, err := os.Open("./city.list.us.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	fmt.Println("opening file" + file.Name())

	indexerCount := 25

	cityChan := make(chan City, 50)
	done := make(chan bool)

	for i := 0; i < indexerCount; i++ {
		// indexer go routines
		go cityIndxer(cityChan, index)
	}

	// json file reader
	go func() {
		var wg sync.WaitGroup

		for scanner.Scan() {
			data := scanner.Bytes()
			wg.Add(1)
			if err != nil {
				fmt.Println(err)
			}
			// go routine the json unmarshalling
			go func() {
				cityPtr := new(City)
				json.Unmarshal(data, &cityPtr)

				// fmt.Print(".")
				cityChan <- *cityPtr
				wg.Done()
			}()
		}

		wg.Wait()
		done <- true
		close(cityChan)
	}()

	<-done

	cityIndex = index

	fmt.Println("json file processing completed")
}

func cityIndxer(cityChan chan City, index bleve.Index) {
	for {
		city, more := <-cityChan
		if more {
			go func() {
				err := index.Index(string(city.ID), city)
				fmt.Print(">")
				if err != nil {
					fmt.Println(err)
					return
				}
			}()
		} else {
			return
		}
	}
}

type WeatherData struct {
	Base   string `json:"base"`
	Clouds struct {
		All int64 `json:"all"`
	} `json:"clouds"`
	Cod   int64 `json:"cod"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
	Dt   int64 `json:"dt"`
	ID   int64 `json:"id"`
	Main struct {
		GrndLevel float64 `json:"grnd_level"`
		Humidity  int64   `json:"humidity"`
		Pressure  float64 `json:"pressure"`
		SeaLevel  float64 `json:"sea_level"`
		Temp      float64 `json:"temp"`
		TempMax   float64 `json:"temp_max"`
		TempMin   float64 `json:"temp_min"`
	} `json:"main"`
	Name string `json:"name"`
	Sys  struct {
		Country string  `json:"country"`
		Message float64 `json:"message"`
		Sunrise int64   `json:"sunrise"`
		Sunset  int64   `json:"sunset"`
	} `json:"sys"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ID          int64  `json:"id"`
		Main        string `json:"main"`
	} `json:"weather"`
	Wind struct {
		Deg   float64 `json:"deg"`
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

func (w WeatherData) String() string {
	return fmt.Sprintf("name = %s\ntemperature = %.1f", w.Name, w.Main.Temp)
}

func WeatherAction(args []string) {
	cityId := args[0]
	appId := args[1]

	resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?id=%s&appid=%s&units=imperial", cityId, appId))

	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := ioutil.ReadAll(bufio.NewReader(resp.Body))
	if err != nil {
		fmt.Println(err)
		return
	}

	weather := new(WeatherData)

	json.Unmarshal(data, &weather)

	fmt.Println(weather)
}

type City struct {
	ID    int64 `json:"_id"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
	Country string `json:"country"`
	Name    string `json:"name"`
}

func CitySearch(partialWord string) []string {
	fmt.Println("searching by " + partialWord)

	// search for some text
	query := bleve.NewMatchQuery(partialWord)
	search := bleve.NewSearchRequest(query)
	searchResults, err := cityIndex.Search(search)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(searchResults)

	return []string{}
}
