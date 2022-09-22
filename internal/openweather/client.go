package openweather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

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

type ClientOption func(*Client)

type Client struct {
	apiKey  string
	baseUrl string
}

func ApiKeyOption(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

func BaseUrlOption(baseUrl string) ClientOption {
	return func(c *Client) {
		c.baseUrl = baseUrl
	}
}

func New(opts ...ClientOption) *Client {
	c := &Client{}

	for _, opt := range opts {
		opt(c)
	}

	if c.apiKey == "" {
		panic("Missing apikey in openweather client")
	}
	if c.baseUrl == "" {
		panic("Missing baseUrl in openweather client")
	}
	return c
}

func (c Client) GetHourlyWeather(lat float64, long float64) (*OneCallResponse, error) {
	req, err := url.Parse(c.baseUrl)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to parse baseUrl %s: %s", c.baseUrl, err.Error()))
		return nil, err
	}

	q := req.Query()
	q.Add("appid", c.apiKey)
	q.Add("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Add("lon", strconv.FormatFloat(long, 'f', -1, 64))
	q.Add("units", "metric")
	q.Add("exclude", "current,minutely,daily,alerts")
	req.RawQuery = q.Encode()

	resp, err := http.Get(req.String())
	if err != nil {
		err = errors.New(fmt.Sprintf("error on openweather api request: %s", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.New(fmt.Sprintf("error code %d returned from openweather", resp.StatusCode))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.New(fmt.Sprintf("error reading body of response: %s", err.Error()))
		return nil, err
	}
	//unmarshal into object
	var respObj OneCallResponse
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		err = errors.New(fmt.Sprintf("error unmarshalling response from openweather: %s", err.Error()))
		return nil, err
	}
	return &respObj, nil
}
