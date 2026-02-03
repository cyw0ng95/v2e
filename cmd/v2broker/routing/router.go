package routing

import (
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// Router abstracts message routing and correlation handling.
type Router interface {
	Route(msg *proc.Message, sourceProcess string) error
	ProcessBrokerMessage(msg *proc.Message) error
}
