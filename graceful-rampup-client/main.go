package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	timeDelay     = "time-delay"
	errorEndpoint = "error"
)

var client = http.Client{
	Timeout: time.Second,
}

var circuitConfigs = map[string]hystrix.CommandConfig{
	timeDelay: {
		Timeout:                10000, //10 seconds
		MaxConcurrentRequests:  50,
		RequestVolumeThreshold: 2,
		SleepWindow:            2000,
		ErrorPercentThreshold:  5,
	},
	errorEndpoint: {
		Timeout:                10000, //10 seconds
		MaxConcurrentRequests:  2,
		RequestVolumeThreshold: 2,
		SleepWindow:            3000,
		ErrorPercentThreshold:  10,
	},
}

func main() {
	hystrix.Configure(circuitConfigs)

	for i := 0; i < 300; i++ {
		go run(i)
		time.Sleep(200 * time.Millisecond)
	}

	<-time.After(5 * time.Minute)
}

func run(i int) {
	ch := make(chan string)
	initiateRequest(timeDelay, ch)
	fmt.Println("Resp[", i, "]: ", <-ch)
}
