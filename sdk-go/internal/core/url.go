package core

import (
	"net/url"
	"sort"
	"strings"
)

// BuildURL concatenates baseURL + apiPath + path and appends a query string.
// Pure function; no I/O. Trailing slashes on baseURL are trimmed; missing
// leading slashes on path are added.
//
// Query values are stable-sorted by key for deterministic output.
func BuildURL(baseURL, apiPath, path string, query map[string]string) string {
	base := strings.TrimRight(baseURL, "/")
	p := path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	out := base + apiPath + p

	if len(query) == 0 {
		return out
	}

	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := url.Values{}
	for _, k := range keys {
		v := query[k]
		if v == "" {
			continue
		}
		values.Add(k, v)
	}
	if encoded := values.Encode(); encoded != "" {
		out += "?" + encoded
	}
	return out
}
