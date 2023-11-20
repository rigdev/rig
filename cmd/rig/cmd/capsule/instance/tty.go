package instance

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/moby/term"
)

var terminationSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}

type Tty struct {
	fdout         uintptr
	fdin          uintptr
	outPrevState  *term.State
	inPrevState   *term.State
	resizeChan    chan sizeMsg
	interruptChan chan struct{}
}

type sizeMsg struct {
	width  uint32
	height uint32
}

func NewTty() *Tty {
	fdout := os.Stdout.Fd()
	fdin := os.Stdin.Fd()
	resizeChan := make(chan sizeMsg, 1)
	interruptChan := make(chan struct{}, 1)

	return &Tty{
		fdout:         fdout,
		fdin:          fdin,
		resizeChan:    resizeChan,
		interruptChan: interruptChan,
	}
}

func (tty *Tty) GetTtySize() (uint32, uint32) {
	ws, err := term.GetWinsize(tty.fdout)
	if err != nil {
		return 0, 0
	}
	return uint32(ws.Width), uint32(ws.Height)
}

func (tty *Tty) SetTtyTerminal() error {
	var err error
	tty.outPrevState, err = term.SetRawTerminalOutput(tty.fdout)
	if err != nil {
		return err
	}

	tty.inPrevState, err = term.SetRawTerminal(tty.fdin)
	if err != nil {
		return err
	}
	return err
}

func (tty *Tty) RestoreTerminal() error {
	if tty.outPrevState != nil {
		err := term.RestoreTerminal(tty.fdout, tty.outPrevState)
		if err != nil {
			return err
		}
	}

	if tty.inPrevState != nil {
		err := term.RestoreTerminal(tty.fdin, tty.inPrevState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tty *Tty) MonitorInterrupt(interruptChan chan error) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, terminationSignals...)
	defer func() {
		signal.Stop(sigChan)
		close(sigChan)
	}()

	go func() {
		for range sigChan {
			interruptChan <- errors.New("interrupted")
		}
	}()

	return nil
}
