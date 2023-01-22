package hko

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type HKOData struct {
	GeneralSituation   string
	Warnings           string
	CurrentTemperature int
	CurrentHumidity    int
	Rainfall           int
}

type HKOCurrentWeather struct {
	Rainfall struct {
		Data []struct {
			Unit  string `json:"unit"`
			Place string `json:"place"`
			Max   int    `json:"max"`
			Main  string `json:"main"`
		} `json:"data"`
		//StartTime string `json:"startTime"`
		//EndTime   string `json:"endTime"`
	} `json:"rainfall"`
	WarningMessage []string `json:"warningMessage"`
	Icon           []int    `json:"icon"`
	UVIndex        string   `json:"uvindex"`
	//UpdateTime     string   `json:"updateTime"`
	Temperature struct {
		Data []struct {
			Place string `json:"place"`
			Value int    `json:"value"`
			Unit  string `json:"unit"`
		} `json:"data"`
		RecordTime string `json:"recordTime"`
	} `json:"temperature"`
	Humidity struct {
		Data []struct {
			Unit  string `json:"unit"`
			Value int    `json:"value"`
			Place string `json:"place"`
		} `json:"data"`
		RecordTime string `json:"recordTime"`
	}
	//TCMessage         string `json:"tcmessage"`
	//MinTEmpFrom00To09 string `json`
}

type HKOLocalWeatherForecast struct {
	GeneralSituation  string `json:"generalSituation"`
	TCInfo            string `json:"tcInfo"`
	FireDangerWanring string `json:"fireDangerWanring"`
	ForecastPeriod    string `json:"forecastPeriod"`
	ForecastDesc      string `json:"forecastDesc"`
	Outlook           string `json:"outlook"`
	UpdateTime        string `json:"updateTime"`
}

const (
	CURRENT_WEATHER        = "https://data.weather.gov.hk/weatherAPI/opendata/weather.php?dataType=rhrread&lang=tc"
	LOCAL_WEATHER_FORECAST = "https://data.weather.gov.hk/weatherAPI/opendata/weather.php?dataType=flw&lang=tc"
)

type GenericChan[T any] chan T

func sendRequest[T any](url string, channel chan T) {
	response, err := http.Get(url)

	if err != nil {
		log.Printf("Error making request to %s: %v", url, err)
	}

	var responseData T
	err = json.NewDecoder(response.Body).Decode(&responseData)

	if err != nil {
		log.Printf("Error decoding json data: %v", err)
	}

	response.Body.Close()

	channel <- responseData

	close(channel)
}

func FetchWeather() HKOData {

	currentWeatherChan := make(chan HKOCurrentWeather)
	go sendRequest(CURRENT_WEATHER, currentWeatherChan)
	currentWeather := <-currentWeatherChan

	localWeatherForecastChan := make(chan HKOLocalWeatherForecast)
	go sendRequest(LOCAL_WEATHER_FORECAST, localWeatherForecastChan)
	localWeatherForecast := <-localWeatherForecastChan

	var currentTemperature int

	for _, weather := range currentWeather.Temperature.Data {
		if weather.Place == "荃灣城門谷" {
			currentTemperature = weather.Value
		}
	}

	return HKOData{
		Warnings:           strings.Join(currentWeather.WarningMessage, "\n"),
		GeneralSituation:   localWeatherForecast.GeneralSituation,
		CurrentTemperature: currentTemperature,
		CurrentHumidity:    currentWeather.Humidity.Data[0].Value,
	}
}

func DoLoop(weatherInfo chan HKOData, sleep time.Duration) {

	for {

		hkoData := FetchWeather()
		weatherInfo <- hkoData

		time.Sleep(sleep)
	}
}
