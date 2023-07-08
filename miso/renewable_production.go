package miso

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type RenewableProduction struct {
	StartAt   time.Time // inclusive
	EndAt     time.Time // exclusive
	Megawatts float64
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

func (c Client) RenewableProduction(ctx context.Context, kind Renewable) ([]RenewableProduction, error) {
	var url string
	switch kind {
	case RenewableSolar:
		url = "https://api.misoenergy.org/MISORTWDDataBroker/DataBrokerServices.asmx?messageType=getSolarActual&returnType=json"
	case RenewableWind:
		url = "https://api.misoenergy.org/MISORTWDDataBroker/DataBrokerServices.asmx?messageType=getWindActual&returnType=json"
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
			DateTimeEST   string `json:"DateTimeEST"` // "2023-07-08 1:00:00 AM"
			HourEndingEST string `json:"HourEndingEST"`
			Value         string `json:"Value"`
		} `json:"instance"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var out []RenewableProduction
	for _, hour := range data.Instance {
		startAt, err := time.ParseInLocation("2006-01-02 3:04:05 PM", hour.DateTimeEST, tz)
		if err != nil {
			return nil, err
		}

		megawatts, err := strconv.ParseFloat(hour.Value, 64)
		if err != nil {
			return nil, err
		}

		out = append(out, RenewableProduction{
			StartAt:   startAt,
			EndAt:     startAt.Add(time.Hour),
			Megawatts: megawatts,
		})
	}

	return out, nil
}
