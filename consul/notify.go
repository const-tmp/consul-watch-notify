package consul

import (
	"bytes"
	"github.com/const-tmp/consul-watch-notify/notify"
	"html/template"
	"log"
)

type NotifyConfig struct {
	Name               string
	RegisterTemplate   *template.Template
	ChangeTemplate     *template.Template
	DeregisterTemplate *template.Template
	Notifier           notify.Interface
}

func (w Watcher[K, V]) notify(k K, v V, getTmpl func(config NotifyConfig) *template.Template, name string) {
	dot := struct {
		K K
		V V
	}{K: k, V: v}

	for _, notifier := range w.Notifiers {
		tmpl := getTmpl(notifier)
		if tmpl == nil {
			continue
		}

		buf := new(bytes.Buffer)

		if err := tmpl.Execute(buf, dot); err != nil {
			log.Printf("render %s template %s error: %s", name, notifier.Name, err.Error())
			continue
		}

		if err := notifier.Notifier.Notify(buf.String()); err != nil {
			log.Printf("notify %s %s error: %s", name, notifier.Name, err.Error())
		}
	}
}

func (w Watcher[K, V]) notifyRegister(k K, v V) {
	w.notify(
		k, v,
		func(config NotifyConfig) *template.Template { return config.RegisterTemplate },
		"register",
	)
}

func (w Watcher[K, V]) notifyDeregister(k K, v V) {
	w.notify(
		k, v,
		func(config NotifyConfig) *template.Template { return config.DeregisterTemplate },
		"deregister",
	)
}

func (w Watcher[K, V]) notifyChange(k K, v V) {
	w.notify(
		k, v,
		func(config NotifyConfig) *template.Template { return config.ChangeTemplate },
		"change",
	)
}
