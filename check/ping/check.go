package ping

import (
	"fmt"
	"github.com/go-ping/ping"
	"time"
)

type Check struct {
	Address string
	Count   int
	Timeout time.Duration
}

func (c Check) Check() error {
	pinger, err := ping.NewPinger(c.Address)
	if err != nil {
		return fmt.Errorf("create pinger error: %w", err)
	}

	pinger.Count = c.Count
	pinger.Timeout = c.Timeout

	if err = pinger.Run(); err != nil {
		return fmt.Errorf("run pinger error: %w", err)
	}

	stats := pinger.Statistics()
	if stats.PacketLoss > 0 {
		return fmt.Errorf("ping %s %.0f%% packet loss", c.Address, stats.PacketLoss)
	}

	return nil
}
