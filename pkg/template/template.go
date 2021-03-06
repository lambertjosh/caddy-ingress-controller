/*	This file is a derivative of https://github.com/kubernetes/ingress/blob/master/controllers/nginx/pkg/template/template.go
	Licensed under the Apache License.  http://www.apache.org/licenses/LICENSE-2.0
*/

package template

import (
	"bytes"
	"log"
	text_template "text/template"

	"k8s.io/ingress/controllers/caddy/pkg/config"

	"k8s.io/ingress/core/pkg/watch"
)

const (
	slash         = "/"
	defBufferSize = 65535
	errNoChild    = "wait: no child processes"
)

// Template
type Template struct {
	tmpl    *text_template.Template
	fw      watch.FileWatcher
	s       int
	tmplBuf *bytes.Buffer
}

// NewTemplate returns a new Template instance or an
// error if the specified template contains errors
func NewTemplate(file string, onChange func()) (*Template, error) {
	tmpl := text_template.Must(text_template.New("Caddyfile.tmpl").Funcs(funcMap).ParseFiles(file))
	fw, err := watch.NewFileWatcher(file, onChange)
	if err != nil {
		return nil, err
	}

	return &Template{
		tmpl:    tmpl,
		fw:      fw,
		s:       defBufferSize,
		tmplBuf: bytes.NewBuffer(make([]byte, 0, defBufferSize)),
	}, nil
}

// Close removes the file watcher
func (t *Template) Close() {
	t.fw.Close()
}

// Write populates a buffer using the template with the Caddy configuration
// and the servers and upstreams created by the Ingress rules
func (t *Template) Write(conf config.TemplateConfig) ([]byte, error) {
	defer t.tmplBuf.Reset()

	defer func() {
		if t.s < t.tmplBuf.Cap() {
			log.Printf("adjusting template buffer size from %v to %v", t.s, t.tmplBuf.Cap())
			t.s = t.tmplBuf.Cap()
			t.tmplBuf = bytes.NewBuffer(make([]byte, 0, t.tmplBuf.Cap()))
		}
	}()

	err := t.tmpl.Execute(t.tmplBuf, conf)
	if err != nil && err.Error() != errNoChild {
		return nil, err
	}

	return t.tmplBuf.Bytes(), nil
}

var (
	funcMap = text_template.FuncMap{
		"cleanHostname": cleanHostname,
	}
)

// cleanHostname will replace the "_" hostname with ""
func cleanHostname(input interface{}) string {
	if hostname, ok := input.(string); ok {
		if hostname == "_" {
			return ""
		}
		return hostname
	}
	return ""
}
