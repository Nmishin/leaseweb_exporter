package leaseweb

import (
	dedicatedserver "github.com/leaseweb/leaseweb-go-sdk/dedicatedserver/v2"
)

var (
	leasewebClient Client
)

type Client struct {
	DedicatedserverAPI dedicatedserver.DedicatedserverAPI
}

func InitLeasewebClient() {
	cfg := dedicatedserver.NewConfiguration()
        apiKey := os.Getenv("LEASEWEB_API_KEY")
	if apiKey == "" {
		log.Fatal("LEASEWEB_API_KEY is not set")
	}

	cfg.AddDefaultHeader("X-LSW-Auth", apiKey)

	leasewebClient = Client{
		DedicatedserverAPI: dedicatedserver.NewAPIClient(cfg).DedicatedserverAPI,
	}
}
