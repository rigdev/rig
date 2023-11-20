//go:build !windows
// +build !windows

package instance

import (
	"os"
	"os/signal"
	"syscall"
)

func (tty *Tty) MonitorSize() error {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGWINCH)
		defer func() {
			signal.Stop(sigChan)
			close(sigChan)
		}()

		for range sigChan {
			w, h := tty.GetTtySize()
			tty.resizeChan <- sizeMsg{
				width:  w,
				height: h,
			}
		}
	}()

	return nil
}
