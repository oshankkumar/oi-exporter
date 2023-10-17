package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type OptionChainIndex struct {
	Records Record `json:"records"`
}

type Record struct {
	Data            []OptionRecord `json:"data"`
	UnderlyingValue float64        `json:"underlyingValue"`
}

type OptionRecord struct {
	StrikePrice int        `json:"strikePrice"`
	ExpiryDate  string     `json:"expiryDate"`
	PE          OptionData `json:"PE"`
	CE          OptionData `json:"CE"`
}

type OptionData struct {
	StrikePrice           int     `json:"strikePrice"`
	ExpiryDate            string  `json:"expiryDate"`
	Underlying            string  `json:"underlying"`
	Identifier            string  `json:"identifier"`
	OpenInterest          float64 `json:"openInterest"`
	ChangeinOpenInterest  float64 `json:"changeinOpenInterest"`
	PchangeinOpenInterest float64 `json:"pchangeinOpenInterest"`
	TotalTradedVolume     int     `json:"totalTradedVolume"`
	ImpliedVolatility     float64 `json:"impliedVolatility"`
	LastPrice             float64 `json:"lastPrice"`
	Change                float64 `json:"change"`
	PChange               float64 `json:"pChange"`
	TotalBuyQuantity      int     `json:"totalBuyQuantity"`
	TotalSellQuantity     int     `json:"totalSellQuantity"`
	BidQty                int     `json:"bidQty"`
	Bidprice              float64 `json:"bidprice"`
	AskQty                int     `json:"askQty"`
	AskPrice              float64 `json:"askPrice"`
	UnderlyingValue       float64 `json:"underlyingValue"`
}

type NSEClient struct {
	BaseURL string
	Doer    Doer
}

func (m *NSEClient) ListOptionChain(ctx context.Context, symbol string) (*OptionChainIndex, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/option-chain-indices?symbol=%s", m.BaseURL, symbol), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("User-Agent", "Chrome/117.0.0.0")

	resp, err := m.Doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Doer.Do: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid http status: %d", resp.StatusCode)
	}

	var idx OptionChainIndex
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, fmt.Errorf("json.NewDecoder.Decode: %w", err)
	}

	return &idx, nil
}
