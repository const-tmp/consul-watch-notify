package config

import (
	"github.com/spf13/viper"
	"time"
)

type (
	Check struct {
		Address string
		Period  time.Duration
		Timeout time.Duration
	}

	Config struct {
		Period time.Duration
		Checks map[string]Check
	}
)

func NewFromViper[T any](v *viper.Viper, f func(*viper.Viper, *T) error) (T, error) {
	obj := new(T)
	err := f(v, obj)
	return *obj, err
}
