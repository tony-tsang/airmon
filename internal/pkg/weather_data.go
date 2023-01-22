package pkg

type GeneralData struct {
	Temperature float64
	Humidity    float64
	Pressure    float64
}

type WeatherData struct {
	PM100            float64
	PM25             float64
	PM10             float64
	IndoorData       GeneralData
	OutdoorData      GeneralData
	GeneralSituation string
	Warnings         string
	Rainfall         float64
}
