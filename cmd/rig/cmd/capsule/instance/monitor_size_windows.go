package instance

import (
	"time"
)

func (tty *Tty) MonitorSize() error {
	go func() {
		prevH, prevW := tty.GetTtySize()
		for {
			time.Sleep(time.Millisecond * 250)
			h, w := tty.GetTtySize()

			if prevW != w || prevH != h {
				tty.resizeChan <- sizeMsg{
					width:  w,
					height: h,
				}
			}
			prevH = h
			prevW = w
		}
	}()

	return nil
}
