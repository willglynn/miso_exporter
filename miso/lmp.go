package miso

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LMP struct {
	FiveMinuteLMP       LMPTime
	HourlyIntegratedLMP LMPTime
	DayAheadExAnteLMP   LMPTime
	DayAheadExPostLMP   LMPTime
}

type LMPTime struct {
	StartAt time.Time // inclusive
	EndAt   time.Time // exclusive
	Nodes   []LMPPriceNode
}

type LMPPriceNode struct {
	Name   string  `json:"name"`
	Region string  `json:"region"`
	LMP    float64 `json:"LMP"` // location marginal price, i.e. total
	MCC    float64 `json:"MCC"` // congestions
	MLC    float64 `json:"MLC"` // loss
	// energy price = LMP - MCC - MLC
}

type lmpPriceNode struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	LMP    string `json:"LMP"`
	MCC    string `json:"MCC"`
	MLC    string `json:"MLC"`
}

func (c Client) LMP(ctx context.Context) (*LMP, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.misoenergy.org/MISORTWDDataBroker/DataBrokerServices.asmx?messageType=getLMPConsolidatedTable&returnType=json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		LMPData struct {
			// "08-Jul-2023 - Interval 13:50 EST"
			RefID string `json:"RefId"`

			// "13:50"
			FiveMinLMP lmpBlob `json:"FiveMinLMP"`

			// "HE 13"
			HourlyIntegratedLMP lmpBlob `json:"HourlyIntegratedLMP"`

			// "HE 14"
			DayAheadExAnteLMP lmpBlob `json:"DayAheadExAnteLMP"`

			// "HE 14"
			DayAheadExPostLMP lmpBlob `json:"DayAheadExPostLMP"`
		} `json:"LMPData"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	date, err := time.ParseInLocation("02-Jan-2006", data.LMPData.RefID[:11], tz)
	if err != nil {
		return nil, fmt.Errorf("error parsing RefID %q: %v", data.LMPData.RefID, err)
	}

	var out LMP
	if err == nil {
		out.FiveMinuteLMP, err = data.LMPData.FiveMinLMP.toLMPTime(date, 5*time.Minute)
	}
	if err == nil {
		out.HourlyIntegratedLMP, err = data.LMPData.HourlyIntegratedLMP.toLMPTime(date, time.Hour)
	}
	if err == nil {
		out.DayAheadExAnteLMP, err = data.LMPData.DayAheadExAnteLMP.toLMPTime(date, time.Hour)
	}
	if err == nil {
		out.DayAheadExPostLMP, err = data.LMPData.DayAheadExPostLMP.toLMPTime(date, time.Hour)
	}

	if err == nil {
		return &out, nil
	} else {
		return nil, err
	}
}

type lmpBlob struct {
	HourAndMin  string         `json:"HourAndMin"`
	PricingNode []lmpPriceNode `json:"PricingNode"`
}

func (b lmpBlob) toLMPTime(ref time.Time, duration time.Duration) (LMPTime, error) {
	var out LMPTime

	if strings.HasPrefix(b.HourAndMin, "HE ") {
		// "hour ending 1" means 00:00-00:59
		// "hour ending 24" means 23:00-23:59
		he, err := strconv.Atoi(b.HourAndMin[3:])
		if err == nil && he >= 1 && he <= 24 {
			out.StartAt = ref.Add(time.Duration((he-1)*3600) * time.Second)
		}
	} else {
		// By the logic above:
		// "11:05" means 11:00-11:04:59
		// "00:00" means 23:55-23:59 the day before
		t, err := time.ParseInLocation("2006-01-02 15:04", ref.Format("2006-01-02")+" "+b.HourAndMin, tz)
		if err == nil {
			out.StartAt = t.Add(-5 * time.Minute)
		}
	}

	if out.StartAt.IsZero() {
		return out, fmt.Errorf("invalid LMP time: %q", b.HourAndMin)
	}

	out.EndAt = out.StartAt.Add(duration)
	for _, node := range b.PricingNode {
		lmp, err := strconv.ParseFloat(node.LMP, 64)
		if err != nil {
			return out, fmt.Errorf("invalid LMP price: %v", err)
		}
		mcc, err := strconv.ParseFloat(node.MCC, 64)
		if err != nil {
			return out, fmt.Errorf("invalid MCC price: %v", err)
		}
		mlc, err := strconv.ParseFloat(node.MLC, 64)
		if err != nil {
			return out, fmt.Errorf("invalid MLC price: %v", err)
		}
		out.Nodes = append(out.Nodes, LMPPriceNode{
			Name:   node.Name,
			Region: node.Region,
			LMP:    lmp,
			MCC:    mcc,
			MLC:    mlc,
		})
	}

	return out, nil
}
