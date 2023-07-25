package tcp

import (
	"fmt"
	"log"
	"net"
)

type Check struct {
	Address string
}

func (c Check) Check() error {
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}

	if err = conn.Close(); err != nil {
		log.Printf("%s connection close error: %s", c.Address, err.Error())
	}

	return nil
}
