package plugin

import (
	"fmt"
	"sort"
	"sync"
)

var (
	registry = make(map[string]Plugin)
	mu       sync.Mutex
)

func Register(p Plugin) {
	mu.Lock()
	defer mu.Unlock()

	name := p.Name()
	if name == "" {
		panic("plugin: Register called with empty name")
	}
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("plugin: Register called twice for %q", name))
	}
	registry[name] = p
}

func All() []Plugin {
	mu.Lock()
	defer mu.Unlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)

	plugins := make([]Plugin, 0, len(names))
	for _, name := range names {
		plugins = append(plugins, registry[name])
	}
	return plugins
}

func Get(name string) (Plugin, bool) {
	mu.Lock()
	defer mu.Unlock()
	p, ok := registry[name]
	return p, ok
}
