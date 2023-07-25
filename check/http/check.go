package http

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"time"
)

type Check struct {
	healthyCode int
	client      http.Client
	req         *http.Request
}

func New(addr, method string, timeout time.Duration, healthyCode int) (Check, error) {
	var (
		req *http.Request
		err error
	)

	switch method {
	case http.MethodGet:
		req, err = http.NewRequest(method, addr, nil)
	default:
		err = fmt.Errorf("method %s is not implemented", method)
	}
	if err != nil {
		return Check{}, err
	}

	return Check{
		client:      http.Client{Timeout: timeout},
		req:         req,
		healthyCode: healthyCode,
	}, nil
}

func NewFromViper(v *viper.Viper) (Check, error) {
	addr := v.GetString("address")
	if addr == "" {
		return Check{}, fmt.Errorf("address required")
	}

	v.SetDefault("method", http.MethodGet)
	method := v.GetString("method")

	v.SetDefault("timeout", 5*time.Second)
	timeout := v.GetDuration("timeout")

	v.SetDefault("code", http.StatusOK)
	code := v.GetInt("code")

	return New(addr, method, timeout, code)
}

type BadCodeError struct {
	Want, Got int
	Body      string
}

func (e BadCodeError) Error() string {
	return fmt.Sprintf("wanted code %d but got %d", e.Want, e.Got)
}

func (h Check) Check() error {
	resp, err := h.client.Do(h.req)
	if err != nil {
		return err
	}
	if resp.StatusCode != h.healthyCode {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read body error: %w", err)
		}
		return BadCodeError{
			Want: h.healthyCode,
			Got:  resp.StatusCode,
			Body: string(body),
		}
	}
	return nil
}
