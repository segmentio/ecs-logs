package cloudwatchlogs

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

type document struct {
	Region string `json:"region"`
}

func getAwsRegion() (region string, err error) {
	var res *http.Response
	var doc document

	if region = __getAwsRegion(); len(region) != 0 {
		return
	}

	regmtx.Lock()
	defer regmtx.Unlock()

	if region = regvar; len(region) != 0 {
		return
	}

	if region = os.Getenv("AWS_REGION"); len(region) != 0 {
		goto saveRegion
	}

	if region = os.Getenv("AWS_DEFAULT_REGION"); len(region) != 0 {
		goto saveRegion
	}

	if res, err = http.Get("http://169.254.169.254/latest/dynamic/instance-identity/document"); err != nil {
		return
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return
	}

	region = doc.Region
saveRegion:
	regvar = region
	return
}

func __getAwsRegion() (region string) {
	regmtx.RLock()

	if len(regvar) != 0 {
		region = regvar
	}

	regmtx.RUnlock()
	return
}

var (
	regmtx sync.RWMutex
	regvar string
)
