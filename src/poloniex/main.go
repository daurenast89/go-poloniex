package main

import (
	"fmt"
	"os"
	"os/exec"
	"poloniex/polo"
	"sort"
	"strconv"
	"time"
)

const (
	API_KEY    = ""
	API_SECRET = ""
)

//Структура для хранения значений пар
//AskAmount-объем спроса
//AskPrice-цена спроса
//BidsAmount-объем предложений
//BidsPrice-цена предложений
type Start struct {
	AskAmount  float64
	AskPrice   float64
	BidsAmount float64
	BidsPrice  float64
}

func main() {
	poloniex := poloniex.New(API_KEY, API_SECRET)
	// Get orders book
	//Создаем карту StartOrderBook для сохранения начальных значений
	StartOrderBook := map[string]Start{}
	orderBook, err := poloniex.GetOrderBook("all", 100)
	if err != nil {
		fmt.Println("JSON error", err)
	}
	//Проходим  по всему orderBook, по каждой паре
	//считаем общее кол-во объема, цены спроса и предложений, считаем среднюю цену

	//
	for key, v := range orderBook {
		AskAvg0 := 0.0
		AskAmount0 := 0.0
		ak := 0.0
		for _, ask := range v.Asks {

			a, err := strconv.ParseFloat(ask[0].(string), 10) //convert interface string to float
			if err != nil {
				fmt.Println("don't convert format", err)
			}
			AskAmount0 = AskAmount0 + ask[1].(float64)
			AskAvg0 = AskAvg0 + a
			ak = ak + 1
		}
		AskAvg0 = AskAvg0 / ak
		BidsAvg0 := 0.0
		BidsAmount0 := 0.0
		bk := 0.0
		for _, bids := range v.Bids {
			b, _ := strconv.ParseFloat(bids[0].(string), 10) //convert interface string to float
			BidsAmount0 = BidsAmount0 + bids[1].(float64)
			BidsAvg0 = BidsAvg0 + b
			bk = bk + 1
		}
		BidsAvg0 = BidsAvg0 / bk
		//формирование карты начального значения StartOrderBook
		if _, ok := StartOrderBook[key]; !ok {
			StartOrderBook[key] = Start{AskAmount: AskAmount0,
				AskPrice:   AskAvg0,
				BidsAmount: BidsAmount0,
				BidsPrice:  BidsAvg0}
		}
	}
	//получаем текущие значения для сравнения
	CurrentOrderBook := map[string]Start{}
	//каждую секунду получаем orderBook.
	//Проходим  по всему orderBook, по каждой паре
	//считаем общее кол-во объема, цены спроса и предложений, считаем среднюю цену
	//сравниваем с начальными значениями, выводим процент изменения объема и цены
	for {
		orderBookCurrent, err := poloniex.GetOrderBook("all", 100)
		if err != nil {
			fmt.Println("JSON error", err)
		}
		for key, v := range orderBookCurrent {
			AskAvg := 0.0
			AskAmount := 0.0
			ak := 0.0
			for _, ask := range v.Asks {

				a, _ := strconv.ParseFloat(ask[0].(string), 10) //convert interface string to float
				AskAmount = AskAmount + ask[1].(float64)
				AskAvg = AskAvg + a
				ak = ak + 1
			}
			AskAvg = AskAvg / ak
			BidsAvg := 0.0
			BidsAmount := 0.0
			bk := 0.0
			for _, bids := range v.Bids {
				b, _ := strconv.ParseFloat(bids[0].(string), 10)
				BidsAmount = BidsAmount + bids[1].(float64)
				BidsAvg = BidsAvg + b
				bk = bk + 1
			}
			BidsAvg = BidsAvg / bk
			if _, ok := StartOrderBook[key]; ok {
				AskAmountPercent := (AskAmount/StartOrderBook[key].AskAmount - 1) * 100
				AskPricePercent := (AskAvg/StartOrderBook[key].AskPrice - 1) * 100
				BidAmountPercent := (BidsAmount/StartOrderBook[key].BidsAmount - 1) * 100
				BidsPricePercent := (BidsAvg/StartOrderBook[key].BidsPrice - 1) * 100

				CurrentOrderBook[key] = Start{AskAmount: AskAmountPercent,
					AskPrice:   AskPricePercent,
					BidsAmount: BidAmountPercent,
					BidsPrice:  BidsPricePercent}
			}
		}
		//fmt.Println(CurrentOrderBook)
		//сортировка карты по ключам
		var keys []string
		for k := range CurrentOrderBook {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			AskAmountS := strconv.FormatFloat(CurrentOrderBook[k].AskAmount, 'f', 2, 64)   //конвертация из float64 to string
			AskPriceS := strconv.FormatFloat(CurrentOrderBook[k].AskPrice, 'f', 2, 64)     //конвертация из float64 to string
			BidsAmountS := strconv.FormatFloat(CurrentOrderBook[k].BidsAmount, 'f', 2, 64) //конвертация из float64 to string
			BidsPriceS := strconv.FormatFloat(CurrentOrderBook[k].BidsPrice, 'f', 2, 64)   //конвертация из float64 to string
			fmt.Println(k, "\t\t\t", "dVA:", AskAmountS, "\t\t\t", "dPA",
				AskPriceS, "\t\t\t", "dVB", BidsAmountS, "\t\t\t", "dVB", BidsPriceS)
		}
		time.Sleep(1 * time.Second) //Ждем секунду до следующего цикла
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
