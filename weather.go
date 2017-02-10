package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"os"

	"strings"

	"github.com/blevesearch/bleve"
)

const banner = `███╗   ███╗██╗   ██╗██████╗  ██████╗ ████████╗
████╗ ████║╚██╗ ██╔╝██╔══██╗██╔═══██╗╚══██╔══╝
██╔████╔██║ ╚████╔╝ ██████╔╝██║   ██║   ██║   
██║╚██╔╝██║  ╚██╔╝  ██╔══██╗██║   ██║   ██║   
██║ ╚═╝ ██║   ██║   ██████╔╝╚██████╔╝   ██║   
╚═╝     ╚═╝   ╚═╝   ╚═════╝  ╚═════╝    ╚═╝   
`

func init() {
	index, err := bleve.Open("./city.bleve")
	if err == nil {
		return
	}

	fmt.Println(banner)
	fmt.Println("Loading cities and creating text search index. Time to complete is 16 seconds")
	mapping := bleve.NewIndexMapping()
	index, err = bleve.New("city.bleve", mapping)
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

	loadCity(index, file)

	fmt.Println("Done")

	err = index.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// json file reader
func loadCity(index bleve.Index, file *os.File) {
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	for scanner.Scan() {
		data := scanner.Bytes()
		cityPtr := new(City)
		json.Unmarshal(data, &cityPtr)
		cityIndxer(index, cityPtr)
	}

}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func cityIndxer(index bleve.Index, city *City) {
	doc := struct {
		cityid int64
		Name   string
	}{
		cityid: city.ID,
		Name:   city.Name,
	}

	err := index.Index(randomString(5), doc)
	if err != nil {
		fmt.Println(err)
		return
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
	bashcmd([]string{"say", fmt.Sprintf(`"weather in %s is now %.2f degrees"`, w.Name, w.Main.Temp)})
	return fmt.Sprintf("name = %s\ntemperature = %.1f", w.Name, w.Main.Temp)
}

func WeatherAction(arg string) {

	args := strings.Split(arg, " ")

	//cityId := "5133268"
	city := strings.Join(args[0:], "%20")

	// fmt.Println("==>" + city)
	// return

	appId := "a12b2abebca2d75b74f6ebb800dc06c2"

	resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=imperial", city, appId))

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

func (c *City) String() string {
	return fmt.Sprintf("ID: %d\nName: %s\n", c.ID, c.Name)
}

func CitySearch(term string) []string {
	if *debug {
		fmt.Printf("\nsearching by [%s]\n", term)
	}

	cityIndex, _ := bleve.Open("city.bleve")

	if cityIndex == nil {
		fmt.Println("cityIndex is nil")
		return []string{}
	}

	// search for some text
	query := bleve.NewMatchQuery(term)
	search := bleve.NewSearchRequest(query)
	searchResults, err := cityIndex.Search(search)

	if err != nil {
		fmt.Println(err)
	}

	if *debug {
		// fmt.Printf("\n%s\n", searchResults)
	}

	var names = make([]string, 0, 0)

	for _, hit := range searchResults.Hits {
		doc, err := cityIndex.Document(hit.ID)
		if err != nil {
			fmt.Println(err)
		}
		for _, f := range doc.Fields {
			switch f.Name() {
			case "Name":
				name := string(f.Value())
				if strings.HasPrefix(name, term) {
					n := strings.Replace(name, term, "", -1)
					names = append(names, n)
				}
			}
		}
	}

	if *debug {
		fmt.Println()
		for _, n := range names {
			fmt.Printf("[%s]\n", n)
		}
	}

	return names
}
