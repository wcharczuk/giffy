package external

import (
	"fmt"
	"time"

	"github.com/stathat/go"
	"github.com/wcharczuk/giffy/server/core"
)

// StatHatRequestTiming posts the request timing to stat hat.
func StatHatRequestTiming(timing time.Duration) {
	if core.ConfigEnvironment() != "local" {
		statHatToken := core.ConfigStathatToken()
		if len(statHatToken) != 0 {
			requestTimingBucket := fmt.Sprintf("request_timing_%s", core.ConfigEnvironment())
			stathat.PostEZValue(requestTimingBucket, statHatToken, float64(timing/time.Millisecond))
		}
	}
}

// StatHatError posts the request timing to stat hat.
func StatHatError() {
	if core.ConfigEnvironment() != "local" {
		statHatToken := core.ConfigStathatToken()
		if len(statHatToken) != 0 {
			errorCountBucket := fmt.Sprintf("error_count_%s", core.ConfigEnvironment())
			stathat.PostEZCount(errorCountBucket, statHatToken, 1)
		}
	}
}

//StatHatSearch logs a search.
func StatHatSearch() {
	if core.ConfigEnvironment() != "local" {
		statHatToken := core.ConfigStathatToken()
		if len(statHatToken) != 0 {
			seachCountBucket := fmt.Sprintf("search_count_%s", core.ConfigEnvironment())
			stathat.PostEZCount(seachCountBucket, statHatToken, 1)
		}
	}
}

//StatHatUserSignup logs a search.
func StatHatUserSignup() {
	if core.ConfigEnvironment() != "local" {
		statHatToken := core.ConfigStathatToken()
		if len(statHatToken) != 0 {
			seachCountBucket := fmt.Sprintf("user_signup_%s", core.ConfigEnvironment())
			stathat.PostEZCount(seachCountBucket, statHatToken, 1)
		}
	}
}
