package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// The Leaseweb Go SDK has no Floating IPs package, so this is a minimal
// hand-written client for the read-only endpoints the exporter needs.

const floatingIPPageLimit = 50

type FloatingIPRange struct {
	ID       string `json:"id"`
	Range    string `json:"range"`
	Location string `json:"location"`
	Type     string `json:"type"`
}

type FloatingIPDefinition struct {
	ID              string `json:"id"`
	RangeID         string `json:"rangeId"`
	Location        string `json:"location"`
	Type            string `json:"type"`
	FloatingIP      string `json:"floatingIp"`
	AnchorIP        string `json:"anchorIp"`
	PassiveAnchorIP string `json:"passiveAnchorIp"`
	Status          string `json:"status"`
	UpdatedAt       string `json:"updatedAt"`
}

type pageMetadata struct {
	TotalCount int `json:"totalCount"`
}

type FloatingIPClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewFloatingIPClient(apiKey string) *FloatingIPClient {
	return &FloatingIPClient{
		baseURL: "https://api.leaseweb.com/floatingIps/v2",
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *FloatingIPClient) ListRanges(ctx context.Context) ([]FloatingIPRange, error) {
	return listAll[FloatingIPRange](ctx, c, "/ranges", "ranges")
}

func (c *FloatingIPClient) ListDefinitions(ctx context.Context, rangeID string) ([]FloatingIPDefinition, error) {
	path := "/ranges/" + url.PathEscape(rangeID) + "/floatingIpDefinitions"
	return listAll[FloatingIPDefinition](ctx, c, path, "floatingIpDefinitions")
}

func listAll[T any](ctx context.Context, c *FloatingIPClient, path, key string) ([]T, error) {
	var all []T
	for offset := 0; ; offset += floatingIPPageLimit {
		var page map[string]json.RawMessage
		q := url.Values{
			"limit":  {strconv.Itoa(floatingIPPageLimit)},
			"offset": {strconv.Itoa(offset)},
		}
		if err := c.get(ctx, path+"?"+q.Encode(), &page); err != nil {
			return nil, err
		}

		var items []T
		if err := json.Unmarshal(page[key], &items); err != nil {
			return nil, fmt.Errorf("decoding %s: %w", key, err)
		}
		all = append(all, items...)

		var meta pageMetadata
		if raw, ok := page["_metadata"]; ok {
			_ = json.Unmarshal(raw, &meta)
		}
		if len(items) < floatingIPPageLimit || len(all) >= meta.TotalCount {
			return all, nil
		}
	}
}

func (c *FloatingIPClient) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-LSW-Auth", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: unexpected status %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
