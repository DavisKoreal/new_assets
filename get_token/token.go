package main

import (
	"fmt"
	"net/http"
	"strings"
)

func get_token(url string) string {
	// Post data to url
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Println("Error posting to url: ", err)
	}
	// fmt.Println("Response body: ", resp.Body)
	// fmt.Println("Response status: ", resp.Status)
	// fmt.Println("Response status code: ", resp.StatusCode)
	// //fmt.Println("Response header: ", resp.Header)
	// fmt.Println("Response header type ", reflect.TypeOf(resp.Header))
	// fmt.Println("Response token ", reflect.TypeOf(resp.Header.Values("Set-Cookie")))
	// //fmt.Println("Response token ", resp.Header.Get("Set-Cookie")
	// // for index, value := range resp.Header.Values("Set-Cookie") {
	// // 	fmt.Println("Index: ", index, "Value: ", value)
	// // }
	// fmt.Println("Response token ", reflect.TypeOf(resp.Header.Values("Set-Cookie")[2]))
	// fmt.Println("Response parts ", strings.Split(resp.Header.Values("Set-Cookie")[2], ";"))
	// for index, value := range strings.Split(resp.Header.Values("Set-Cookie")[2], ";") {
	// 	fmt.Println("Index: ", index, "Value: ", value)
	// }
	// fmt.Println("Response token ", strings.Split(resp.Header.Values("Set-Cookie")[2], ";")[0])
	res := (strings.Split((strings.Split((resp.Header.Values("Set-Cookie")[2]), ";")[0]), "ken="))[1]
	return res
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

func main() {
	// Post data to url
	token := get_the_token()
	fmt.Println("Token: ", token)
}
