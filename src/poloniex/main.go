package main

import (
	"fmt"
	"poloniex/polo"
	"strconv"
	"os"
)

const (
	API_KEY    = ""
	API_SECRET = ""
)
type Start struct {
	AskAmount float64
	AskPrice float64


}

func main() {
	// Poloniex client
	poloniex := poloniex.New(API_KEY, API_SECRET)

	// Get Ticker (BTC-VTC)
	/*
		ticker, err := poloniex.GetTicker("BTC-DRK")
		fmt.Println(err, ticker)
	*/

	// Get Tickers
	/*
		tickers, err := poloniex.GetTickers()
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			for key, ticker := range tickers {
				fmt.Printf("Ticker: %s, Last: %.8f\n", key, ticker.Last)
			}
		}
		tickerName := "BTC_FLO"
		ticker, ok := tickers[tickerName]
		if ok {
			fmt.Printf("BTC_FLO Last: %.8f\n", ticker.Last)
		} else {
			fmt.Println("ticker not found - ", tickerName)
		}
	*/

	// Get Volumes
	/*
		volumes, err := poloniex.GetVolumes()
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			for key, volume := range volumes.Volumes {
				fmt.Printf("Ticker: %s Value: %#v\n", key, volume["BTC"])
			}
		}
	*/

	// Get CandleStick chart data ( OHLCV )
	/*
		candles, err := client.ChartData("BTC_XMR", 300, time.Now().Add(-time.Hour), time.Now())
		if err != nil {
			panic(err)
		}
		for _, candle := range candles {
			fmt.Printf("BTC_XMR %s\tOpened at: %f\tClosed at: %f\n", candle.Date, candle.Open, candle.Close)
		}
	*/

	// Get markets
	/*
		markets, err := poloniex.GetMarkets()
		fmt.Println(err, markets)
	*/

	// Get orders book
	//Создание файла
	file, err := os.Create("poloniex.txt")
	if err != nil {
		fmt.Println("Не удалось создать файл poloniex.txt")
		return
	}
	defer file.Close()
		orderBook, err := poloniex.GetOrderBook("all",10)
		if err != nil {
			fmt.Println("JSON error", err)
		}
	for key,v := range orderBook {
		fmt.Println(key, v)
		AskAvg:= 0.0
		AskAmount:= 0.0
		for _, ask := range v.Asks{

			a,_:=strconv.ParseFloat(ask[0].(string),10)//convert interface string to float
			AskAmount = AskAmount + ask[1].(float64)
			AskAvg = AskAvg + a
		}
		//fmt.Println("AskAmount=", AskAmount)
		AskAmount1:= strconv.FormatFloat(AskAmount, 'f', 8, 64)
		AskPrice:=strconv.FormatFloat(AskAvg/10, 'f', 8, 64)
		//fmt.Println("AskPrice=", AskAvg/10)
		BidsAvg:= 0.0
		BidsAmount:= 0.0
		for _, bids := range v.Bids{
			b,_:=strconv.ParseFloat(bids[0].(string), 10)
			BidsAmount = BidsAmount + bids[1].(float64)
			BidsAvg = BidsAvg + b
		}
		//fmt.Println("BidsAmount=", BidsAmount)
		BidsAmount1:= strconv.FormatFloat(BidsAmount, 'f', 8, 64)//conv float to str after "." - 6 symbol
		BidsPrice:= strconv.FormatFloat(BidsAvg/10, 'f', 8, 64)
		//fmt.Println("BidsPrice=", BidsAvg/10)
		//открытие файла и доваление записи
		f, err := os.OpenFile("poloniex.txt", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil{
			fmt.Println("Не удалось открыть файл poloniex.txt")
		}
		if _, err = f.WriteString(key + "\t" +"\t"+ "AskAmount="+ AskAmount1+"\t"+"\t"+"AskPrice="+AskPrice+"\t"+"\t"+"BidsAmount="+BidsAmount1+"\t"+"\t"+"BidsPrice="+BidsPrice+"\n"); err != nil {
			fmt.Println("Не удалось записать строку в файл", err)
		}
		f.Close()
	}




	// Market history
	/*
		marketHistory, err := poloniex.GetMarketHistory("BTC-DRK", 100)
		for _, trade := range marketHistory {
			fmt.Println(err, trade.Timestamp.String(), trade.Quantity, trade.Price)
		}
	*/

	// Market

	// BuyLimit
	/*
		uuid, err := poloniex.BuyLimit("BTC-DOGE", 1000, 0.00000102)
		fmt.Println(err, uuid)
	*/

	// BuyMarket
	/*
		uuid, err := poloniex.BuyLimit("BTC-DOGE", 1000)
		fmt.Println(err, uuid)
	*/

	// Sell limit
	/*
		uuid, err := poloniex.SellLimit("BTC-DOGE", 1000, 0.00000115)
		fmt.Println(err, uuid)
	*/

	// Cancel Order
	/*
		err := poloniex.CancelOrder("e3b4b704-2aca-4b8c-8272-50fada7de474")
		fmt.Println(err)
	*/

	// Get open orders
	/*
		orders, err := poloniex.GetOpenOrders("BTC-DOGE")
		fmt.Println(err, orders)
	*/

	// Account
	// Get balances
	/*
		balances, err := poloniex.GetBalances()
		fmt.Println(err, balances)
	*/

	// Get balance
	/*
		balance, err := poloniex.GetBalance("DOGE")
		fmt.Println(err, balance)
	*/

	// Get address
	/*
		address, err := poloniex.GetDepositAddress("QBC")
		fmt.Println(err, address)
	*/

	// WithDraw
	/*
		whitdrawUuid, err := poloniex.Withdraw("QYQeWgSnxwtTuW744z7Bs1xsgszWaFueQc", "QBC", 1.1)
		fmt.Println(err, whitdrawUuid)
	*/

	// Get order history
	/*
		orderHistory, err := poloniex.GetOrderHistory("BTC-DOGE", 10)
		fmt.Println(err, orderHistory)
	*/

	// Get getwithdrawal history
	/*
		withdrawalHistory, err := poloniex.GetWithdrawalHistory("all", 0)
		fmt.Println(err, withdrawalHistory)
	*/

	// Get deposit history
	/*
		deposits, err := poloniex.GetDepositHistory("all", 0)
		fmt.Println(err, deposits)
	*/

}
