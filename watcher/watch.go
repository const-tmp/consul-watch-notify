package watcher

import (
	"bytes"
	"context"
	"github.com/const-tmp/consul-watch-notify/check"
	"github.com/const-tmp/consul-watch-notify/notify"
	"html/template"
	"log"
	"time"
)

type (
	Watcher struct {
		Name      string
		Period    time.Duration
		Check     check.Interface
		Notifiers []NotifyConfig
	}

	NotifyConfig struct {
		Name     string
		Notifier notify.Interface
		Template *template.Template
	}

	NotifyTemplate struct{ Name, Message string }
)

func (w Watcher) Start(ctx context.Context) {
	if err := w.Check.Check(); err != nil {
		w.notify(err)
	}

	tick := time.Tick(w.Period)
	for {
		select {
		case <-ctx.Done():
			log.Println(w.Name + "context done")
			return
		case <-tick:
			if err := w.Check.Check(); err != nil {
				w.notify(err)
			}
		}
	}
}

func (w Watcher) notify(err error) {
	for _, notifyConfig := range w.Notifiers {
		buf := new(bytes.Buffer)
		dot := NotifyTemplate{Name: w.Name, Message: err.Error()}
		if tmplErr := notifyConfig.Template.Execute(buf, dot); tmplErr != nil {
			log.Printf("[%s]\t%s notify render template error: %s", w.Name, notifyConfig.Name, tmplErr.Error())
			continue
		}
		if notifyErr := notifyConfig.Notifier.Notify(buf.String()); notifyErr != nil {
			log.Printf("[%s]\t%s notify error: %s", w.Name, notifyConfig.Name, notifyErr.Error())
		}
	}
}
