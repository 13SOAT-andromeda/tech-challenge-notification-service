package domain

import (
	"errors"
	"strings"
)

var ErrTemplateNotFound = errors.New("template not found")

type Template struct {
	Type    string
	Subject string
	HTML    string
}

func (t *Template) Render(data map[string]string) string {
	result := t.HTML
	for key, value := range data {
		result = strings.ReplaceAll(result, "$"+key, value)
	}
	return result
}
