package miso

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type RenewableDatapoint struct {
	StartAt   time.Time // inclusive
	EndAt     time.Time // exclusive
	Megawatts float64
}

type RenewableProduction struct {
	Actual   []RenewableDatapoint
	Forecast []RenewableDatapoint
}

type Renewable int

func (r Renewable) String() string {
	switch r {
	case RenewableSolar:
		return "solar"
	case RenewableWind:
		return "wind"
	default:
		return ""
	}
}

const (
	RenewableSolar Renewable = iota + 1
	RenewableWind
)

func (c Client) RenewableProduction(ctx context.Context, kind Renewable) (*RenewableProduction, error) {
	var url string
	switch kind {
	case RenewableSolar:
		url = "https://api.misoenergy.org/MISORTWDDataBroker/DataBrokerServices.asmx?messageType=getSolar&returnType=json"
	case RenewableWind:
		url = "https://api.misoenergy.org/MISORTWDDataBroker/DataBrokerServices.asmx?messageType=getWind&returnType=json"
	default:
		return nil, errors.New("invalid kind")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		MktDay   string `json:"MktDay"`
		RefId    string `json:"RefId"`
		Instance []struct {
			ForecastDateTimeEST   *string `json:"ForecastDateTimeEST"`
			ForecastHourEndingEST *string `json:"ForecastHourEndingEST"`
			ForecastValue         *string `json:"ForecastValue"`
			ActualDateTimeEST     *string `json:"ActualDateTimeEST"`
			ActualHourEndingEST   *string `json:"ActualHourEndingEST"`
			ActualValue           *string `json:"ActualValue"`
		} `json:"instance"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var forecast, actual []RenewableDatapoint
	for _, hour := range data.Instance {
		if hour.ForecastHourEndingEST != nil && hour.ForecastValue != nil {
			startAt, err := time.ParseInLocation("2006-01-02 3:04:05 PM", *hour.ForecastDateTimeEST, tz)
			if err != nil {
				return nil, err
			}

			megawatts, err := strconv.ParseFloat(*hour.ForecastValue, 64)
			if err != nil {
				return nil, err
			}

			forecast = append(forecast, RenewableDatapoint{
				StartAt:   startAt,
				EndAt:     startAt.Add(time.Hour),
				Megawatts: megawatts,
			})
		}

		if hour.ActualDateTimeEST != nil && hour.ActualValue != nil {
			startAt, err := time.ParseInLocation("2006-01-02 3:04:05 PM", *hour.ActualDateTimeEST, tz)
			if err != nil {
				return nil, err
			}

			megawatts, err := strconv.ParseFloat(*hour.ActualValue, 64)
			if err != nil {
				return nil, err
			}

			actual = append(actual, RenewableDatapoint{
				StartAt:   startAt,
				EndAt:     startAt.Add(time.Hour),
				Megawatts: megawatts,
			})
		}
	}

	return &RenewableProduction{
		Actual:   actual,
		Forecast: forecast,
	}, nil
}
