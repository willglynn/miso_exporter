package miso

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Fuel struct {
	StartAt   time.Time // inclusive
	EndAt     time.Time // exclusive
	Name      string
	Megawatts float64
}

func (c Client) Fuel(ctx context.Context) ([]Fuel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.misoenergy.org/MISORTWDDataBroker/DataBrokerServices.asmx?messageType=getfuelmix&returnType=json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		RefId   string `json:"RefId"` // "08-Jul-2023 - Interval 16:40 EST"
		TotalMW string `json:"TotalMW"`
		Fuel    struct {
			Type []struct {
				INTERVALEST  string `json:"INTERVALEST"` // "2023-07-08 4:40:00 PM"
				CATEGORY     string `json:"CATEGORY"`
				ACT          string `json:"ACT"`
				FUELCATEGORY string `json:"FUEL_CATEGORY"`
			} `json:"Type"`
		} `json:"Fuel"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var out []Fuel
	for _, fuel := range data.Fuel.Type {
		endAt, err := time.ParseInLocation("2006-01-02 3:04:05 PM", fuel.INTERVALEST, tz)
		if err != nil {
			return nil, err
		}

		startAt := endAt.Add(-5 * time.Minute)

		megawatts, err := strconv.ParseFloat(fuel.ACT, 64)
		if err != nil {
			return nil, err
		}

		out = append(out, Fuel{
			StartAt:   startAt,
			EndAt:     endAt,
			Name:      fuel.CATEGORY,
			Megawatts: megawatts,
		})
	}

	return out, nil
}
