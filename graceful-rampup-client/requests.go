package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

type thRequest struct {
	respChan chan string
	req      *http.Request
	endpoint string
}

var reqChannel chan thRequest
var ticker *time.Ticker

func init() {
	reqChannel = make(chan thRequest, 10000)
	ticker = time.NewTicker(Throttlers[timeDelay].tickerTime)

	go requestThrottler()
}

func requestThrottler() {
	for {

		t := <-ticker.C
		fmt.Println("ticker ticked at: ", t)
		newReqStruct := <-reqChannel
		respChan := newReqStruct.respChan
		sendRequest(newReqStruct.endpoint, newReqStruct.req, respChan)

	}
}

func createDelayRequest(delay string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8084/"+timeDelay+"?delay="+delay, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return req
}

func createErrorRequest() *http.Request {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8084/"+errorEndpoint, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return req
}

func getRequest(endpoint string) *http.Request {
	switch endpoint {
	case timeDelay:
		return createDelayRequest("3s")
	default:
		return createErrorRequest()
	}
}

func initiateRequest(endpoint string, resp chan string) {
	req := getRequest(endpoint)

	if Throttlers.IsThrottling(endpoint) {
		newThRequest := thRequest{
			respChan: resp,
			endpoint: endpoint,
			req:      req,
		}
		reqChannel <- newThRequest
	} else {
		go sendRequest(endpoint, req, resp)
	}
}

func sendRequest(endpoint string, request *http.Request, respChan chan string) {
	output := make(chan *http.Response)
	startTime := time.Now()
	errCh := hystrix.Go(endpoint, func() error {
		resp, err := client.Do(request)
		if err != nil {
			return err
		} else {
			if resp.StatusCode/100 == http.StatusInternalServerError/100 { // check for 5xx error code
				circuit, _, _ := hystrix.GetCircuit(endpoint)
				_ = circuit.ReportEvent([]string{"failure"}, startTime, time.Since(startTime)) // circuit breaker doesn't works as hystrix doesn't considers an actual server response for circuit breaker functionality
			}
			output <- resp
			return nil
		}
	}, nil)

	select {
	case resp := <-output:
		respChan <- fmt.Sprintln("response is: ", resp.Status, " time: ", time.Now())
		func() {
			if resp.StatusCode/100 != http.StatusInternalServerError/100 {
				func() {
					if Throttlers.IsThrottling(endpoint) {
						Throttlers.GracefulRampup(endpoint)
					}
				}()
			}

		}()

	case err := <-errCh:
		respChan <- fmt.Sprintln("Error is: ", err.Error(), " time: ", time.Now())
		//fmt.Println(err.Error(), Throttlers.IsThrottling(endpoint))
		if err.Error() == hystrix.ErrCircuitOpen.Error() {
			func() {
				if Throttlers.IsThrottling(endpoint) {
					Throttlers.UpdateTicker(endpoint)
				} else {
					Throttlers.EnableThrottling(endpoint)
				}
			}()
		}

	}
}
