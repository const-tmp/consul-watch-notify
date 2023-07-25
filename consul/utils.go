package consul

import (
	"github.com/const-tmp/consul-watch-notify/utils"
	"github.com/spf13/viper"
	"html/template"
)

func TemplatesFromViper(name string, v *viper.Viper) (*template.Template, *template.Template, *template.Template) {
	var registerTmpl, changeTmpl, deregisterTmpl *template.Template

	if text := v.GetString("register"); text != "" {
		registerTmpl = template.Must(
			template.New(name + ".register").
				Funcs(utils.FuncMap).
				Parse(text),
		)
	}

	if text := v.GetString("change"); text != "" {
		changeTmpl = template.Must(
			template.New(name + ".change").
				Funcs(utils.FuncMap).
				Parse(text),
		)
	}

	if text := v.GetString("deregister"); text != "" {
		deregisterTmpl = template.Must(
			template.New(name + ".deregister").
				Funcs(utils.FuncMap).
				Parse(text),
		)
	}

	return registerTmpl, changeTmpl, deregisterTmpl
}
