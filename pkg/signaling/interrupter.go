package signaling

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Interrupter func(err chan<- error)

// TerminateInterrupter - sends an error on termination signal, eg ctrl+c
func TerminateInterrupter(err chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	err <- fmt.Errorf("%s", <-c)
}
