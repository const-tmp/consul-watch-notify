package notify

import (
	"errors"
	"fmt"
	"github.com/const-tmp/consul-watch-notify/notify/telegram"
	"github.com/spf13/viper"
)

type Interface interface {
	Notify(message string) error
}

var TypeRequired = errors.New("type field is required")

func NewFromViper(v *viper.Viper) (Interface, error) {
	typ := v.GetString("type")
	if typ == "" {
		return nil, TypeRequired
	}

	switch typ {
	case "telegram":
		return telegram.NewFromViper(v)
	default:
		return nil, fmt.Errorf("type %s is not implemented", typ)
	}
}
