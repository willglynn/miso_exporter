package miso

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type LoadAndForecastDatapoint struct {
	At        time.Time
	Megawatts int
}

type LoadAndForecast struct {
	// every 5 minutes
	FiveMinuteLoad []LoadAndForecastDatapoint
	// every 1 hour
	HourlyForecast []LoadAndForecastDatapoint
}

func (c Client) LoadAndForecast(ctx context.Context) (*LoadAndForecast, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.misoenergy.org/MISORTWDDataBroker/DataBrokerServices.asmx?messageType=gettotalload&returnType=json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		LoadInfo struct {
			RefID string `json:"RefId"` // "07-Jul-2023 - Interval 22:30 EST"

			ClearedMW []struct {
				ClearedMWHourly struct {
					Hour  string `json:"Hour"`
					Value string `json:"Value"`
				} `json:"ClearedMWHourly"`
			} `json:"ClearedMW"`

			MediumTermLoadForecast []struct {
				Forecast struct {
					HourEnding   string `json:"HourEnding"`
					LoadForecast string `json:"LoadForecast"`
				} `json:"Forecast"`
			} `json:"MediumTermLoadForecast"`

			FiveMinTotalLoad []struct {
				Load struct {
					Time  string `json:"Time"` // "22:10"
					Value string `json:"Value"`
				} `json:"Load"`
			} `json:"FiveMinTotalLoad"`
		} `json:"LoadInfo"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	dateStr := data.LoadInfo.RefID[:11]
	date, err := time.ParseInLocation("02-Jan-2006", dateStr, tz)
	if err != nil {
		return nil, fmt.Errorf("error parsing RefID %q: %v", data.LoadInfo.RefID, err)
	}

	var hourlyForecast []LoadAndForecastDatapoint
	for _, forecast := range data.LoadInfo.MediumTermLoadForecast {
		hour, err := strconv.Atoi(forecast.Forecast.HourEnding)
		if err != nil {
			return nil, fmt.Errorf("error parsing MediumTermLoadForecast hour %q: %v", forecast.Forecast.HourEnding, err)
		}
		if hour < 1 || hour > 24 {
			return nil, fmt.Errorf("MediumTermLoadForecast hour out of range: %v", hour)
		}

		t := date.Add(time.Duration(hour-1) * time.Hour)

		value, err := strconv.Atoi(forecast.Forecast.LoadForecast)
		if err != nil {
			return nil, fmt.Errorf("error parsing MediumTermLoadForecast value %q: %v", forecast.Forecast.LoadForecast, err)
		}

		hourlyForecast = append(hourlyForecast, LoadAndForecastDatapoint{
			At:        t,
			Megawatts: value,
		})
	}

	var fiveMinuteLoad []LoadAndForecastDatapoint
	for _, load := range data.LoadInfo.FiveMinTotalLoad {
		if load.Load.Time == "" {
			b, _ := json.Marshal(load)
			log.Printf("skipping: %s", string(b))
			continue
		}

		t, err := time.ParseInLocation("02-Jan-2006 15:04", dateStr+" "+load.Load.Time, tz)
		if err != nil {
			return nil, fmt.Errorf("error parsing FiveMinTotalLoad time %q: %v", load.Load.Time, err)
		}

		value, err := strconv.Atoi(load.Load.Value)
		if err != nil {
			return nil, fmt.Errorf("error parsing FiveMinTotalLoad value %q: %v", load.Load.Value, err)
		}

		fiveMinuteLoad = append(fiveMinuteLoad, LoadAndForecastDatapoint{
			At:        t,
			Megawatts: value,
		})
	}

	return &LoadAndForecast{
		FiveMinuteLoad: fiveMinuteLoad,
		HourlyForecast: hourlyForecast,
	}, nil
}
