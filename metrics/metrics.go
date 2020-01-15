package metrics

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Metrics contains required information about
type Metrics struct {
	name    string
	tags    map[string]string
	taglist []string
	timesrc func() int64
}

func New(name string) *Metrics {
	metrics := &Metrics{
		name:    name,
		tags:    map[string]string{},
		taglist: []string{},
	}
	metrics.timesrc = metrics.now

	return metrics
}

func (m *Metrics) now() int64 {
	return time.Now().Unix()
}

func (m *Metrics) generateList() {
	keys := make([]string, 0, len(m.tags))
	for key := range m.tags {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	m.taglist = make([]string, len(m.tags))
	for i, key := range keys {
		m.taglist[i] = fmt.Sprintf("%s=%s", key, m.tags[key])
	}
}

func (m *Metrics) Log(name string, value interface{}) {
	var taglist string

	if len(m.taglist) > 0 {
		taglist = " " + strings.Join(m.taglist, " ")
	}

	fmt.Printf(
		"%s.%s %d %v%s\n",
		m.name,
		name,
		m.timesrc(),
		value,
		taglist,
	)
}

func (m *Metrics) With(newTags map[string]string) *Metrics {
	newMetrics := &Metrics{name: m.name, timesrc: m.timesrc, tags: map[string]string{}}

	for key, val := range m.tags {
		newMetrics.tags[key] = val
	}

	for key, val := range newTags {
		newMetrics.tags[sanitize(key)] = sanitize(val)
	}

	newMetrics.generateList()

	return newMetrics
}

func sanitize(input string) string {
	output := make([]byte, 0, len(input))

	for i := 0; i < len(input); i++ {
		char := input[i]
		if ('a' <= char && char <= 'z') ||
			('A' <= char && char <= 'Z') ||
			('0' <= char && char <= '9') ||
			char == '-' ||
			char == '_' ||
			char == '.' ||
			char == '/' {
			output = append(output, char)
		} else if char == ' ' {
			output = append(output, '_')
		}
	}

	return string(output)
}
