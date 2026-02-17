package collector

import "sync"

var (
	collectors = make(map[string]*DedicatedServerCollector)
	mu         sync.Mutex
)

func GetCollector(target string) (*DedicatedServerCollector, error) {
	mu.Lock()
	c, ok := collectors[target]
	if !ok {
		c = NewDedicatedServerCollector(target)
		collectors[target] = c
	}
	mu.Unlock()

	c.collected.L.Lock()
	defer c.collected.L.Unlock()

	return c, nil
}
