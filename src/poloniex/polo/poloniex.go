// Package Poloniex is an implementation of the Poloniex API in Golang.
package poloniex

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"crypto/hmac"
	"crypto/sha512"
	"strconv"
)

const (
	API_BASE                   = "https://poloniex.com/" // Poloniex API endpoint
	DEFAULT_HTTPCLIENT_TIMEOUT = 30                      // HTTP client timeout
)

// New return a instantiate poloniex struct
func New(apiKey, apiSecret string) *Poloniex {
	client := NewClient(apiKey, apiSecret)
	return &Poloniex{client}
}

// poloniex represent a poloniex client
type Poloniex struct {
	client *client
}
type OrderBook struct {
	Asks     [][]interface{} `json:"asks"`
	Bids     [][]interface{} `json:"bids"`
	IsFrozen int             `json:"isFrozen,string"`
	Error    string          `json:"error"`
}
type OrderBookAll struct {
	Pair map[string]OrderBook
	Error string
}

// GetTickers is used to get the ticker for all markets
func (b *Poloniex) GetTickers() (tickers map[string]Ticker, err error) {
	r, err := b.client.do("GET", "public?command=returnTicker", "", false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &tickers); err != nil {
		return
	}
	return
}

// GetVolumes is used to get the volume for all markets
func (b *Poloniex) GetVolumes() (vc VolumeCollection, err error) {
	r, err := b.client.do("GET", "public?command=return24hVolume", "", false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &vc); err != nil {
		return
	}
	return
}

func (b *Poloniex) GetCurrencies() (currencies Currencies, err error) {
	r, err := b.client.do("GET", "public?command=returnCurrencies", "", false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &currencies.Pair); err != nil {
		return
	}
	return
}

// GetOrderBook is used to get retrieve the orderbook for a given market
// market: a string literal for the market (ex: BTC_NXT). 'all' not implemented.
// cat: bid, ask or both to identify the type of orderbook to return.
// depth: how deep of an order book to retrieve
func (b *Poloniex) GetOrderBook(market, cat string, depth int) (orderBook map[string]OrderBook, err error) {
	// not implemented
	if cat != "bid" && cat != "ask" && cat != "both" {
		cat = "both"
	}
	if depth > 100 {
		depth = 100
	}
	if depth < 1 {
		depth = 1
	}

	r, err := b.client.do("GET", fmt.Sprintf("public?command=returnOrderBook&currencyPair=%s&depth=%d", strings.ToUpper(market), depth), "", false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &orderBook); err != nil {
		return
	}
	/*if orderBook.Error != "" {
		err = errors.New(orderBook.Error)
		return
	}*/
	return
}

// Returns candlestick chart data. Required GET parameters are "currencyPair",
// "period" (candlestick period in seconds; valid values are 300, 900, 1800,
// 7200, 14400, and 86400), "start", and "end". "Start" and "end" are given in
// UNIX timestamp format and used to specify the date range for the data
// returned.
func (b *Poloniex) ChartData(currencyPair string, period int, start, end time.Time) (candles []*CandleStick, err error) {
	r, err := b.client.do("GET", fmt.Sprintf(
		"/public?command=returnChartData&currencyPair=%s&period=%d&start=%d&end=%d",
		strings.ToUpper(currencyPair),
		period,
		start.Unix(),
		end.Unix(),
	), "", false)
	if err != nil {
		return
	}

	if err = json.Unmarshal(r, &candles); err != nil {
		return
	}

	return
}
type CandleStick struct {
	Date            PoloniexDate `json:"date"`
	High            float64      `json:"high"`
	Low             float64      `json:"low"`
	Open            float64      `json:"open"`
	Close           float64      `json:"close"`
	Volume          float64      `json:"volume"`
	QuoteVolume     float64      `json:"quoteVolume"`
	WeightedAverage float64      `json:"weightedAverage"`
}
type client struct {
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	throttle   <-chan time.Time
}

var (
	// Technically 6 req/s allowed, but we're being nice / playing it safe.
	reqInterval = 200 * time.Millisecond
)

// NewClient return a new Poloniex HTTP client
func NewClient(apiKey, apiSecret string) (c *client) {
	return &client{apiKey, apiSecret, &http.Client{}, time.Tick(reqInterval)}
}

// doTimeoutRequest do a HTTP request with timeout
func (c *client) doTimeoutRequest(timer *time.Timer, req *http.Request) (*http.Response, error) {
	// Do the request in the background so we can check the timeout
	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := c.httpClient.Do(req)
		done <- result{resp, err}
	}()
	// Wait for the read or the timeout
	select {
	case r := <-done:
		return r.resp, r.err
	case <-timer.C:
		return nil, errors.New("timeout on reading data from Poloniex API")
	}
}

func (c *client) makeReq(method, resource, payload string, authNeeded bool, respCh chan<- []byte, errCh chan<- error) {
	body := []byte{}
	connectTimer := time.NewTimer(DEFAULT_HTTPCLIENT_TIMEOUT * time.Second)

	var rawurl string
	if strings.HasPrefix(resource, "http") {
		rawurl = resource
	} else {
		rawurl = fmt.Sprintf("%s/%s", API_BASE, resource)
	}

	req, err := http.NewRequest(method, rawurl, strings.NewReader(payload))
	if err != nil {
		respCh <- body
		errCh <- errors.New("You need to set API Key and API Secret to call this method")
		return
	}
	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Add("Accept", "application/json")

	// Auth
	if authNeeded {
		if len(c.apiKey) == 0 || len(c.apiSecret) == 0 {
			respCh <- body
			errCh <- errors.New("You need to set API Key and API Secret to call this method")
			return
		}
		nonce := time.Now().UnixNano()
		q := req.URL.Query()
		q.Set("apikey", c.apiKey)
		q.Set("nonce", fmt.Sprintf("%d", nonce))
		req.URL.RawQuery = q.Encode()
		mac := hmac.New(sha512.New, []byte(c.apiSecret))
		_, err = mac.Write([]byte(req.URL.String()))
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Add("apisign", sig)
	}

	resp, err := c.doTimeoutRequest(connectTimer, req)
	if err != nil {
		respCh <- body
		errCh <- err
		return
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		respCh <- body
		errCh <- err
		return
	}
	if resp.StatusCode != 200 {
		respCh <- body
		errCh <- errors.New(resp.Status)
		return
	}

	respCh <- body
	errCh <- nil
	close(respCh)
	close(errCh)
}

// do prepare and process HTTP request to Poloniex API
func (c *client) do(method, resource, payload string, authNeeded bool) (response []byte, err error) {
	respCh := make(chan []byte)
	errCh := make(chan error)
	<-c.throttle
	go c.makeReq(method, resource, payload, authNeeded, respCh, errCh)
	response = <-respCh
	err = <-errCh
	return
}
type Currency struct {
	Name               string  `json:"name"`
	MaxDailyWithdrawal string  `json:"maxDailyWithdrawal"`
	TxFee              float64 `json:"txFee,string"`
	MinConf            int     `json:"minConf"`
	Disabled           int     `json:"disabled"`
	Frozen             int     `json:"frozen"`
	Delisted           int     `json:"delisted"`
}

type Currencies struct {
	Pair map[string]Currency
}
type PoloniexDate struct {
	time.Time
}

func (pd *PoloniexDate) UnmarshalJSON(data []byte) error {
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return errors.New("Timestamp invalid (can't parse int64)")
	}
	pd.Time = time.Unix(i, 0)
	return nil
}

type Tickers struct {
	Pair map[string]Ticker
}

type Ticker struct {
	Last          float64 `json:"last,string"`
	LowestAsk     float64 `json:"lowestAsk,string"`
	HighestBid    float64 `json:"highestBid,string"`
	PercentChange float64 `json:"percentChange,string"`
	BaseVolume    float64 `json:"baseVolume,string"`
	QuoteVolume   float64 `json:"quoteVolume,string"`
	IsFrozen      int     `json:"isFrozen,string"`
	High24Hr      float64 `json:"high24hr,string"`
	Low24Hr       float64 `json:"low24hr,string"`
}
type Volume map[string]float64

type VolumeCollection struct {
	TotalBTC  float64 `json:"totalBTC,string"`
	TotalUSDT float64 `json:"totalUSDT,string"`
	TotalXMR  float64 `json:"totalXMR,string"`
	TotalXUSD float64 `json:"totalXUSD,string"`
	Volumes   map[string]Volume
}

func (tc *VolumeCollection) UnmarshalJSON(b []byte) error {
	m := make(map[string]json.RawMessage)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	tc.Volumes = make(map[string]Volume)
	for k, v := range m {
		switch k {
		case "totalBTC":
			f, err := parseJSONFloatString(v)
			if err != nil {
				return err
			}
			tc.TotalBTC = f
		case "totalUSDT":
			f, err := parseJSONFloatString(v)
			if err != nil {
				return err
			}
			tc.TotalUSDT = f
		case "totalXMR":
			f, err := parseJSONFloatString(v)
			if err != nil {
				return err
			}
			tc.TotalXMR = f
		case "totalXUSD":
			f, err := parseJSONFloatString(v)
			if err != nil {
				return err
			}
			tc.TotalXUSD = f
		default:
			t := make(Volume)
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			tc.Volumes[k] = t
		}
	}
	return nil
}

func (t *Volume) UnmarshalJSON(b []byte) error {
	m := make(map[string]json.RawMessage)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		f, err := parseJSONFloatString(v)
		if err != nil {
			return err
		}
		(*t)[k] = f
	}
	return nil
}

func parseJSONFloatString(b json.RawMessage) (float64, error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}