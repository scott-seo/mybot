package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
