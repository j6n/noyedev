package main

import (
	"bytes"
	"log"
	"text/template"
)

type option struct {
	Name, Val, Help string
}

func (o *option) Set(val string) {
	o.Val = val
}

func (o *option) String() string {
	return o.Val
}

const optMessage = `available options:{{ range $i, $element := . }}
  {{$i}}: {{$element.Help}}{{ end }}
current:{{ range $i, $element := . }}
  {{$i}}: {{$element.Val}}{{ end }}`

var optTmpl = template.Must(template.New("set").Parse(optMessage))

func getOps() string {
	out := &bytes.Buffer{}

	if err := optTmpl.Execute(out, opts); err != nil {
		log.Println("err execute", err)
		return ""
	}

	return out.String()
}
