package main

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"encoding/json"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	//"net/http"
	"os/exec"
	"time"

	"github.com/Kucoin/kucoin-go-sdk"

	"log"
	"net/http"
	"os"
	"strings"
	//"github.com/rs/zerolog/log"
	//"github.com/go-gota/gota/dataframe"
	//"github.com/go-gota/gota/series"
)

const (
	ChannelTicker   string = "ticker"
	TypeSubscribe   string = "subscribe"
	TypeUnsubscribe string = "unsubscribe"
)

type Message struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Topic    string `json:"topic"`
	Response bool   `json:"responseonse"`
}

type Trade struct {
	Type   string    `json:"type"`
	Time   time.Time `json:"time"`
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
}

//{"topic":"/market/ticker:all","type":"message","data":{"bestAsk":"0.17257","bestAskSize":"2311.8817","bestBid":"0.17151",
//"bestBidSize":"2150.1183","price":"0.17157","sequence":"533524567","size":"43.4673","time":1714317416582},"subject":"UOS-USDT"}

type data struct {
	BestAsk     string `json:"bestAsk"`
	BestAskSize string `json:"bestAskSize"`
	BestBid     string `json:"bestBid"`
	BestBidSize string `json:"bestBidSize"`
	Price       string `json:"price"`
	Sequence    string `json:"sequence"`
	Size        string `json:"size"`
	Time        int    `json:"time"`
}

type response_struct struct {
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Data    data   `json:"data"`
	Subject string `json:"subject"`
}

type state_variables struct {
	Subject string
	Price   float64
	Time    int
}

type bought_coins struct {
	Subject       string
	number_bought float64
}

type investment_track struct {
	Subject     string
	Percentgain float64
	start_price float64
}

func this_coin_is_usdt(subject string) bool {
	return subject[len(subject)-4:] == "USDT"
}

// type row struct {
// 	Subject string
// 	Price float64
// 	Volume float64
// 	Pricechange float64
// 	Volumechange float64

// }

// KUCOINAPIKEY = "661bd98603e77600013bfd3d"
// KUCOINSECRET = "7a0c92b9-69dc-40c0-8a01-30d73f660560"
// #KUCOINPASSWORD = "47H6PJsb-5WRidM"
// KUCOINPASSWORD = "*9Sd49G.!rt4RC$"
var KUCOINAPIKEY = "661bd98603e77600013bfd3d"
var KUCOINSECRET = "7a0c92b9-69dc-40c0-8a01-30d73f660560"
var KUCOINPASSWORD = "*9Sd49G.!rt4RC$"

// var tradingdollars string = "1"
var tradedollarsfloat float64 = 5
var population_size int = 0
var target_percentage float64 = 50
var stop_loss_percentage float64 = -15
var connection_websocket = false

var Adress string = "wss://ws-api.kucoin.com/?token=2neAiuYvAU61ZDXANAGAsiL4-iAExhsBXZxftpOeh_55i3Ysy2q2LEsEWU64mdzUOPusi34M_wGoSf7iNyEWJ751xiGSPyebzyUy2oXl06UwcMPlW-6PiNiYB9J6i9GjsxUuhPw3BlrzazF6ghq4L93hHTSO1paWJCFQSppLjAw=.tfpBWYok2QDijX13JYQ7NQ==&[connectId=Dave2024]"

// var number_of_top_to_buy int = 5
var filter_price_change float64 = 10

// var negativepercentage float64 = -1
var buytrials = 0
var dollarsused float64 = 0.0

//var maxnumberoftrades int = 5

// maps used to store data
var data_map = make(map[string]state_variables)
var purchase_map = make(map[string]bought_coins)
var tradestracking = make(map[string]investment_track)

// initiating variable s to be used in the runtime
var s *kucoin.ApiService = kucoin.NewApiService(
	kucoin.ApiKeyOption(KUCOINAPIKEY),
	kucoin.ApiSecretOption(KUCOINSECRET),
	kucoin.ApiPassPhraseOption(KUCOINPASSWORD),
)

//functions used in the runtime

func serverTime(s *kucoin.ApiService) {
	rsp, err := s.ServerTime()
	if err != nil {
		log.Printf("Error: %s", err.Error())
		// Handle error
		return
	}

	var ts int64
	if err := rsp.ReadData(&ts); err != nil {
		// Handle error
		return
	}
	log.Printf("The server time: %d", ts)
}

func round_to_one_decimal(number float64) float64 {
	return float64(int(number*10)) / 10
}

func coin_has_been_purchased(coin_to_buy string) bool {
	_, coin_is_in_map := purchase_map[coin_to_buy]
	return coin_is_in_map
}

func coin_has_been_tracked(coin_to_track string) bool {
	_, coin_is_in_map := tradestracking[coin_to_track]
	return coin_is_in_map
}

func buy(coin_to_buy string) {
	file, errs := os.OpenFile("buy.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errs != nil {
		fmt.Println("Failed to create file:", errs)
		return
	}
	defer file.Close()
	size_of_trade := round_to_one_decimal(tradedollarsfloat / data_map[coin_to_buy].Price)
	floatString := strconv.FormatFloat(float64(size_of_trade), 'f', -1, 64)
	p := &kucoin.CreateOrderModel{
		ClientOid: kucoin.IntToString(time.Now().UnixNano()),
		Side:      "buy",
		Symbol:    coin_to_buy,
		Type:      "market",
		Size:      floatString,
	}
	// Open the file for writing
	_, errs = file.WriteString("Buying " + floatString + " of " + coin_to_buy + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}

	rsp, err := s.CreateOrder(p)
	if err != nil {
		fmt.Println(err)
		fmt.Println("The coin ", coin_to_buy, " has not been bought")
		return
	}
	_, errs = file.WriteString(rsp.Message + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
	//fmt.Println("response: ", rsp)
	dollarsused += tradedollarsfloat
	fmt.Println(coin_to_buy, " has been bought")
	action := "Bought " + floatString + " of " + coin_to_buy + "at price " + strconv.FormatFloat(data_map[coin_to_buy].Price, 'f', 6, 32) + " at time " + time.Now().String()
	log_actions(action)
}

func sell(coin_to_sell string, number_to_sell string) {

	file, errs := os.OpenFile("sell.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errs != nil {
		fmt.Println("Failed to create file:", errs)
		return
	}
	defer file.Close()

	p := &kucoin.CreateOrderModel{
		ClientOid: kucoin.IntToString(time.Now().UnixNano()),
		Side:      "sell",
		Symbol:    coin_to_sell,
		Type:      "market",
		Size:      number_to_sell,
	}
	_, errs = file.WriteString("Selling " + number_to_sell + " of " + coin_to_sell + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
	rsp, err := s.CreateOrder(p)
	if err != nil {
		fmt.Println(err)
		fmt.Println("The coin ", coin_to_sell, " has not been sold")
		return
	}
	_, errs = file.WriteString(rsp.Message + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
	fmt.Println("response: ", rsp)
	action := "Sold " + number_to_sell + " of " + coin_to_sell + "at price " + strconv.FormatFloat(data_map[coin_to_sell].Price, 'f', 6, 32) + " at time " + time.Now().String()
	log_actions(action)
}

func track_investment(coin_to_track string, current_price float64, start_price float64, number_bought float64) {

	if coin_has_been_tracked(coin_to_track) {

		percentgain := (current_price - start_price) / start_price * 100
		tradestracking[coin_to_track] = investment_track{coin_to_track, percentgain, start_price}

		if percentgain > target_percentage {
			sell(coin_to_track, strconv.FormatFloat(number_bought, 'f', 6, 64))
			log_actions("Target percentage reached. Sold " + strconv.FormatFloat(float64(purchase_map[coin_to_track].number_bought), 'f', 6, 32) + " of " + coin_to_track + " at time " + time.Now().String())
			delete(purchase_map, coin_to_track)
			delete(tradestracking, coin_to_track)
		}
		if percentgain < stop_loss_percentage {
			sell(coin_to_track, strconv.FormatFloat(number_bought, 'f', 6, 64))
			log_actions("Stop loss triggered. Sold " + strconv.FormatFloat(float64(purchase_map[coin_to_track].number_bought), 'f', 6, 32) + " of " + coin_to_track + " at time " + time.Now().String())
			delete(purchase_map, coin_to_track)
			delete(tradestracking, coin_to_track)
		}

	}
}

func add_new_coin(new_coin string) {
	file, errs := os.OpenFile("COINS.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errs != nil {
		fmt.Println("Failed to create file:", errs)
		return
	}
	defer file.Close()
	_, errs = file.WriteString(new_coin + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
}

// func coin_is_in_file(coin_to_check string) bool {
// 	file, errs := os.Open("COINS.txt")
// 	if errs != nil {
// 		fmt.Println("Failed to open file:", errs)
// 		coin_is_in_file(coin_to_check)
// 	}

// 	defer file.Close()
// 	var coin string
// 	for {
// 		_, errs := fmt.Fscanf(file, "%s\n", &coin)
// 		if errs != nil {
// 			break
// 		}
// 		fmt.Println(coin)
// 		if coin == coin_to_check {
// 			return true
// 		}
// 	}
// 	return false
// }

func log_actions(action string) {
	file, errs := os.OpenFile("LOGS.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errs != nil {
		fmt.Println("Failed to create file:", errs)
		return
	}
	defer file.Close()
	_, errs = file.WriteString(action + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}

}

func populate_data_map() {
	file, errs := os.Open("COINS.txt")
	if errs != nil {
		fmt.Println("Failed to open file:", errs)
		populate_data_map()
	}

	defer file.Close()
	var coin string
	for {
		_, errs := fmt.Fscanf(file, "%s\n", &coin)
		if errs != nil {
			break
		}
		fmt.Println(coin)
		state := state_variables{coin, 0.0, 0}
		data_map[coin] = state
	}
	population_size += 1
}

// func closealltrades() {
// 	for key := range purchase_map {
// 		sell(purchase_map[key].Subject, strconv.FormatFloat(float64(purchase_map[key].number_bought), 'f', 6, 32))
// 	}
// 	tradingisallowed = false
// }

func printvariables() {
	fmt.Println("The number of records seen is: \t", records_seen)
	fmt.Println("Buy trials is: \t", buytrials)
	fmt.Println("Connection is: ", connection_websocket)
	fmt.Println("Dollars used is: ", dollarsused)
	fmt.Println("Filter price change is: ", filter_price_change)
	fmt.Println("Target percentage is: ", target_percentage)
	fmt.Println("Stop loss percentage is: ", stop_loss_percentage)
	fmt.Println("Tradedollarsfloat is: ", tradedollarsfloat)
	fmt.Println("Population trials is: ", population_size)
	fmt.Println("THe lenght of the data map is: ", len(data_map))
	if len(tradestracking) > 0 {
		for key := range tradestracking {
			fmt.Println("Coin: ", key, " StartPrice", tradestracking[key].start_price, " CurrentPrice ", data_map[key].Price, " Gain : ", tradestracking[key].Percentgain, "percent")
		}
	}
	fmt.Println("\n............")
}

func clear_terminal() {
	cmd := exec.Command("clear")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
	fmt.Println("Output: ", string(out))
}

func get_the_token() string {
	// Post data to url
	var url string = "https://api.kucoin.com/api/v1/bullet-public"
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Println("Error posting to url: ", err)
	}
	res := (strings.Split((strings.Split((resp.Header.Values("Set-Cookie")[2]), ";")[0]), "ken="))[1]
	return res
}

var records_seen int = 0

//var nrow row =
//var df dataframe.DataFrame= dataframe.LoadStructs([]row{ {"firstrand", 0.987, 2.3, 0.0, 0.0}})

func main() {
	serverTime(s)
	populate_data_map()

	c, _, err := websocket.Dial(context.Background(), Adress, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("I am connected")
	connection_websocket = true

	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	_, message, err := c.Read(context.Background())
	fmt.Println("The type of message is:	", reflect.TypeOf(message))
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("The message is: ", message)
	fmt.Println("Message: ", message)
	//fmt.Println("The message is " + string(message))

	//subscribing to a channel in the websocket by defining a message first annd connecting to channel
	sub := Message{
		Id:       "Dave2024",
		Type:     "subscribe",
		Topic:    "/market/ticker:all",
		Response: true,
	}
	//fmt.Println("sending the message to the channel")
	//fmt.Println("The message is: ", sub)
	//fmt.Println("The type of message is:	", reflect.TypeOf(sub))
	err = wsjson.Write(context.Background(), c, sub)
	if err != nil {
		fmt.Println(err)
		return
	}

	//reading the responseonse from the channel
	for i := 2; i > 1; i++ {

		clear_terminal()
		printvariables()

		_, message, err := c.Read(context.Background())
		if err != nil {
			fmt.Println(err)
			Adress = "wss://ws-api.kucoin.com/?token=" + get_the_token() + "&[connectId=Dave2024]"
			main()
		}

		stringmessage := string(message)
		records_seen += 1
		var response response_struct
		err = json.Unmarshal([]byte(stringmessage), &response)

		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("The message is: ", stringmessage)
			fmt.Println("The routine for reading a json has been triggered ")
			main()
			return
		}

		if !this_coin_is_usdt(response.Subject) {
			continue
		}

		stateprice, _ := strconv.ParseFloat(response.Data.Price, 64)
		statetime := time.Now().Nanosecond()
		state := state_variables{response.Subject, stateprice, statetime}

		_, coin_is_in_map := data_map[response.Subject]
		if !coin_is_in_map {

			add_new_coin(response.Subject)
			log_actions("Added " + response.Subject + " to the list of coins at time " + time.Now().String())
			data_map[response.Subject] = state
			buy(response.Subject)
			ammount := round_to_one_decimal(tradedollarsfloat / data_map[response.Subject].Price)
			purchase_map[response.Subject] = bought_coins{response.Subject, ammount}
			buytrials += 1
			tradestracking[response.Subject] = investment_track{response.Subject, 0.0, stateprice}

			continue
		}

		if coin_is_in_map {
			data_map[response.Subject] = state
			fmt.Println("The coin is in the map. Checking if it is purchased and tracking investment")
			if coin_has_been_purchased(response.Subject) {
				number_bought := purchase_map[response.Subject].number_bought
				start_price := tradestracking[response.Subject].start_price
				track_investment(response.Subject, stateprice, start_price, number_bought)
			}
		}

	}

}
