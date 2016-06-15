package cwl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type document struct {
	Region string `json:"region"`
}

func getAwsRegion() (region string, err error) {
	if region = __getAwsRegion(); len(region) != 0 {
		return
	}

	regmtx.Lock()
	defer regmtx.Unlock()

	if region = regvar; len(region) != 0 {
		return
	}

	fmt.Println("fetching AWS region from EC2 instance metadata...")

	var res *http.Response
	var doc document

	if res, err = http.Get("http://169.254.169.254/latest/dynamic/instance-identity/document"); err != nil {
		return
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return
	}

	region = doc.Region
	regvar = region
	fmt.Println("the AWS region is", region)
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
