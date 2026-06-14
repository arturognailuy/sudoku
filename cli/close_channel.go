package cli

import (
	"os"
	"os/signal"
)

// CloseChannel provides a signal-based shutdown mechanism for the game loop.
type CloseChannel chan struct{}

// handleInterruptSignal listens for OS interrupt signals and closes the channel.
func (closeChannel CloseChannel) handleInterruptSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			closeChannel.Close()
		}
	}()
}

// NewCloseChannel creates a new close channel that responds to OS interrupt signals.
func NewCloseChannel() CloseChannel {
	var closeChannel CloseChannel = make(chan struct{})
	closeChannel.handleInterruptSignal()

	return closeChannel
}

// IsClosed checks if the close channel has been closed (non-blocking).
func (closeChannel CloseChannel) IsClosed() bool {
	select {
	case <-closeChannel:
		return true
	default:
	}

	return false
}

// Close closes the channel, signaling shutdown.
func (closeChannel CloseChannel) Close() {
	close(closeChannel)
}
