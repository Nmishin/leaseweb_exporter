package client

import (
	dedicatedserver "github.com/leaseweb/leaseweb-go-sdk/dedicatedserver/v2"
)

type Client struct {
	DedicatedserverAPI dedicatedserver.DedicatedserverAPI
}

var LeasewebClient Client

func Init(apiKey string) {
	cfg := dedicatedserver.NewConfiguration()
	cfg.AddDefaultHeader("X-LSW-Auth", apiKey)

	LeasewebClient = Client{
		DedicatedserverAPI: dedicatedserver.NewAPIClient(cfg).DedicatedserverAPI,
	}
}

