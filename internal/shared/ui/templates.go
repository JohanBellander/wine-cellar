package ui

import (
	"html/template"
	"net/url"
)

// FuncMap contains shared template functions
var FuncMap = template.FuncMap{
	"safeURL": func(s string) template.URL {
		return template.URL(s)
	},
	"sortURL": func(base url.Values, field, currentSort, currentDirection string) string {
		v := url.Values{}
		// Copy existing values
		for key, val := range base {
			v[key] = val
		}
		v.Set("sort", field)
		if field == currentSort && currentDirection == "asc" {
			v.Set("direction", "desc")
		} else {
			v.Set("direction", "asc")
		}
		v.Del("page") // Reset page when sorting changes
		return "?" + v.Encode()
	},
}
