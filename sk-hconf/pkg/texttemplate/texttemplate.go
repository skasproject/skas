package texttemplate

import (
	"bytes"
	"gopkg.in/yaml.v3"
	"strings"
	"text/template"
)

type TextTemplate interface {
	RenderToText(model map[string]interface{}) (string, error)
	RenderToMap(model map[string]interface{}) (map[string]interface{}, error)
	SetDelimiters(d1, d2 string)
}

var _ TextTemplate = &textTemplate{}

type textTemplate struct {
	template *template.Template
}

func New(templateName string, tempText string) (TextTemplate, error) {
	var err error
	tt := &textTemplate{}
	tt.template = template.New(templateName).Option("missingkey=zero").Funcs(funcMap())
	tt.template, err = tt.template.Parse(tempText)
	if err != nil {
		return nil, err
	}
	return tt, nil
}

func (tt *textTemplate) SetDelimiters(d1, d2 string) {
	tt.template.Delims(d1, d2)
}

func (tt *textTemplate) RenderToText(model map[string]interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := tt.template.Execute(buf, model)
	if err != nil {
		return "", err
	}
	// Work around the issue where Go will emit "<no value>" even if Options(missing=zero)
	// is set. Since missing=error will never get here, we do not need to handle
	// the Strict case.
	return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
}

// Helper functions

func (tt *textTemplate) RenderToMap(model map[string]interface{}) (map[string]interface{}, error) {
	txt, err := tt.RenderToText(model)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(txt), m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewAndRenderToMap(templateName string, tmplText string, model map[string]interface{}) (map[string]interface{}, error) {
	tt, err := New(templateName, tmplText)
	if err != nil {
		return nil, err
	}
	m, err := tt.RenderToMap(model)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewAndRenderToText(templateName string, tmplText string, model map[string]interface{}) (string, error) {
	tt, err := New(templateName, tmplText)
	if err != nil {
		return "", err
	}
	txt, err := tt.RenderToText(model)
	if err != nil {
		return "", err
	}
	return txt, nil
}
