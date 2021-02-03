package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var count = 0

func delayFunc(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("[", count, "]request received at :", time.Now())
	//params := r.URL.Query()
	//delay := params["delay"][0]
	var delay string
	if count < 20 {
		delay = "3s"
	} else {
		delay = "10ms"
	}
	count++

	delayDuration, err := time.ParseDuration(delay)

	if err != nil {
		<-time.After(1 * time.Second)
	} else {
		<-time.After(delayDuration)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "Success")
	//fmt.Println("response sent after ", delay, " at ", time.Now())
}

func errorFunc(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("request received at :", time.Now())
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprint(w, "Error")
	fmt.Println("response sent at ", time.Now())
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/time-delay", delayFunc).Methods(http.MethodGet)

	r.HandleFunc("/error", errorFunc).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe("localhost:8084", r))
}
