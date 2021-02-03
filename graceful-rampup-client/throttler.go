package main

import (
	"fmt"
	"sync"
	"time"
)

var rwLock sync.RWMutex

type ThrottleState struct {
	enable               bool
	tickerTime           time.Duration
	lastTickerUpdateTime time.Time
}

type ThrottleStates map[string]ThrottleState

var Throttlers = ThrottleStates{
	timeDelay: {
		enable:               false,
		tickerTime:           10 * time.Millisecond,
		lastTickerUpdateTime: time.Time{},
	},
	errorEndpoint: {
		enable:               false,
		tickerTime:           10 * time.Millisecond,
		lastTickerUpdateTime: time.Time{},
	},
}

func (t ThrottleStates) EnableThrottling(key string) {
	rwLock.Lock()

	newState := t[key]
	newState.enable = true

	t[key] = newState
	rwLock.Unlock()
	fmt.Println("EnabledThrottling for endpoint:", key, "at time:", time.Now())
}

func (t ThrottleStates) UpdateTicker(key string) {
	rwLock.Lock()
	defer rwLock.Unlock()
	if time.Since(t[key].lastTickerUpdateTime) > time.Duration(circuitConfigs[key].SleepWindow)*time.Millisecond {
		newState := t[key]
		newState.lastTickerUpdateTime = time.Now()
		newState.tickerTime = minTime(10*time.Second, 2*t[key].tickerTime)

		t[key] = newState
		ticker.Stop()
		ticker = time.NewTicker(newState.tickerTime)
		fmt.Println("UpdatedTicker for endpoint:", key, "at time:", time.Now(), "newValue:", t[key].tickerTime)
	}
}

func (t ThrottleStates) GracefulRampup(key string) {
	rwLock.Lock()
	defer rwLock.Unlock()
	if t[key].enable && time.Since(t[key].lastTickerUpdateTime) > time.Duration(circuitConfigs[key].SleepWindow)*time.Millisecond {

		newState := t[key]
		newState.lastTickerUpdateTime = time.Now()
		if newState.tickerTime == 10*time.Millisecond {
			newState.enable = false
			fmt.Println("GracefulRampup for endpoint:", key, "at time:", time.Now(), "disabled throttling")
		} else {
			newState.tickerTime = maxTime(10*time.Millisecond, t[key].tickerTime/2)
			fmt.Println("GracefulRampup for endpoint:", key, "at time:", time.Now(), "newValue:", newState.tickerTime)
		}

		t[key] = newState
		ticker.Stop()
		ticker = time.NewTicker(newState.tickerTime)

	}
}

func (t ThrottleStates) IsThrottling(key string) bool {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return t[key].enable
}

func maxTime(t1, t2 time.Duration) time.Duration {
	if t1 > t2 {
		return t1
	} else {
		return t2
	}
}
func minTime(t1, t2 time.Duration) time.Duration {
	if t1 < t2 {
		return t1
	} else {
		return t2
	}
}
