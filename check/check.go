package check

import (
	"fmt"
	"github.com/const-tmp/consul-watch-notify/check/http"
	"github.com/const-tmp/consul-watch-notify/check/ping"
	"github.com/const-tmp/consul-watch-notify/check/tcp"
	"github.com/const-tmp/consul-watch-notify/config"
	"github.com/spf13/viper"
	"time"
)

type Interface interface {
	Check() error
}

func NewFromViper(v *viper.Viper) (Interface, error) {
	v.SetDefault("type", "http")
	typ := v.GetString("type")
	switch typ {
	case "http":
		return http.NewFromViper(v)
	case "tcp":
		return config.NewFromViper[tcp.Check](v, func(v *viper.Viper, p *tcp.Check) error {
			addr := v.GetString("address")
			if addr == "" {
				return fmt.Errorf("address required")
			}
			return v.Unmarshal(p)
		})
	case "ping":
		return config.NewFromViper[ping.Check](v, func(v *viper.Viper, p *ping.Check) error {
			v.SetDefault("count", 5)
			v.SetDefault("timeout", time.Second)
			addr := v.GetString("address")
			if addr == "" {
				return fmt.Errorf("address required")
			}
			return v.Unmarshal(p)
		})
	default:
		return nil, fmt.Errorf("%s is not implemented", typ)
	}
}
