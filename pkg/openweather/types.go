package openweather

type Conditions struct {
	Id          int
	Main        string
	Description string
	Icon        string
}

type HourlyWeather struct {
	Time    int64 `json:"dt"`
	Weather []Conditions
	Pop     float64
}

type OneCallResponse struct {
	Lat    float64
	Lon    float64
	Hourly []HourlyWeather
}
