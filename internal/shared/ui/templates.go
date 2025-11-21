package ui

import (
	"html/template"
)

// FuncMap contains shared template functions
var FuncMap = template.FuncMap{
	"safeURL": func(s string) template.URL {
		return template.URL(s)
	},
}
