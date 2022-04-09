package loader

import (
	"github.com/evanhutnik/wipercheck-loader/pkg/openweather"
	"go.uber.org/zap"
	"math"
	"os"
	"strconv"
	"time"
)

type Coordinate struct {
	lat float64
	lon float64
}

type Loader struct {
	coordinate   Coordinate
	stepDistance float64
	duration     time.Duration

	ow     *openweather.Client
	logger *zap.SugaredLogger
}

func New() *Loader {
	l := &Loader{}

	distance, err := strconv.ParseFloat(os.Getenv("loader_stepdistance"), 64)
	if err != nil {
		panic("Invalid loader_stepdistance env variable")
	}
	l.stepDistance = distance

	duration, err := strconv.ParseInt(os.Getenv("loader_duration"), 10, 0)
	if err != nil {
		panic("Invalid loader_duration env variable")
	}
	l.duration = time.Duration(duration) * time.Minute

	lat, err := strconv.ParseFloat(os.Getenv("loader_start_lat"), 64)
	if err != nil {
		panic("Invalid loader_start_lat variable")
	}
	lon, err := strconv.ParseFloat(os.Getenv("loader_start_lon"), 64)
	if err != nil {
		panic("Invalid loader_start_lon variable")
	}
	l.coordinate = Coordinate{
		lat: lat,
		lon: lon,
	}

	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync()
	logger := baseLogger.Sugar()
	l.logger = logger

	l.ow = openweather.New(
		openweather.ApiKeyOption(os.Getenv("openweather_apikey")),
		openweather.BaseUrlOption(os.Getenv("openweather_baseurl")),
	)

	return l
}

func (l Loader) Load() {
	start := l.coordinate
	//starting column and row values for loading
	column, row := float64(1), float64(1)
	//because we are loading 1 coordinate/second, square root of load duration (in seconds) is max column/row value.
	//ex. if duration = 1 hour (3600 seconds), sqrt(3600) = 60, therefore population square is 60x60
	maxColumn := math.Sqrt(float64(l.duration / time.Second))
	maxRow := maxColumn
	l.logger.Info("Loading hourly weather data...")
	for {
		weather, err := l.ow.GetHourlyWeather(l.coordinate.lat, l.coordinate.lon)
		if err != nil {
			l.logger.Errorw(err.Error(),
				"lat", l.coordinate.lat,
				"long", l.coordinate.lon)
		}
		l.logger.Infof("Weather returned for coordinates (%v,%v)", weather.Lat, weather.Lon)
		//insert into db

		//calculate next step
		column++
		l.moveRight()
		if column > maxColumn {
			column = 1
			if row < maxRow {
				row++
				l.moveDown()
				l.coordinate.lon = start.lon
			} else {
				row = 1
				l.coordinate = start
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (l *Loader) moveRight() {
	//depending on your distance from the equator, there is a different distance between each degree of longitude.
	//the closer you are to the poles, the closer the distance between degrees of longitude
	kmBetweenDegrees := 111.2 * math.Cos(l.coordinate.lat*math.Pi/180)
	//temporarily adding 180 to longitude so we can work with all positive numbers
	tempLong := l.coordinate.lon + 180
	//moving coordinate to the right (increasing longitude) by amount of stepDistance
	tempLong = tempLong + l.stepDistance/kmBetweenDegrees
	//handling 180 -> -180 longitudinal wraparound
	tempLong = math.Mod(tempLong, 360)

	l.coordinate.lon = tempLong - 180
}

func (l *Loader) moveDown() {
	//assuming earth is a perfect sphere (for simplicity's sake), the distance between degrees of latitude is constant (111.2km)
	l.coordinate.lat = l.coordinate.lat - l.stepDistance/111.2
}
