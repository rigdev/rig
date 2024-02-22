package common

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/fatih/color"
)

type bytesWithSource struct {
	bytes    []byte
	isStdOut bool
}

type buffer struct {
	lock  sync.Mutex
	bytes []bytesWithSource
}

func (b *buffer) write(bytes []byte, isStdOut bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.bytes = append(b.bytes, bytesWithSource{
		bytes:    bytes,
		isStdOut: isStdOut,
	})
}

type bufferWithSource struct {
	buffer   *buffer
	isStdOut bool
}

func (c *bufferWithSource) Write(p []byte) (int, error) {
	c.buffer.write(p, c.isStdOut)
	return len(p), nil
}

type cmdBuffer struct {
	buffer buffer
	stdOut bufferWithSource
	stdErr bufferWithSource
}

func (c *cmdBuffer) makeStdOut() []byte {
	c.buffer.lock.Lock()
	defer c.buffer.lock.Unlock()
	var bytes []byte
	for _, b := range c.buffer.bytes {
		if b.isStdOut {
			bytes = append(bytes, b.bytes...)
		}
	}
	return bytes
}

func (c *cmdBuffer) print() {
	for _, b := range c.buffer.bytes {
		if b.isStdOut {
			os.Stdout.WriteString(string(b.bytes))
		} else {
			os.Stderr.WriteString(string(b.bytes))
		}
	}
}

func (c *cmdBuffer) clean() {
	c.buffer.lock.Lock()
	defer c.buffer.lock.Unlock()
	c.buffer.bytes = nil
}

func newCmdBuffer() *cmdBuffer {
	b := &cmdBuffer{buffer: buffer{}}
	b.stdOut = bufferWithSource{
		buffer:   &b.buffer,
		isStdOut: true,
	}
	b.stdErr = bufferWithSource{
		buffer:   &b.buffer,
		isStdOut: false,
	}
	return b
}

type DeferredOutputCommand struct {
	cmdBuffer      *cmdBuffer
	cmd            *exec.Cmd
	displayMessage string
	hasStarted     bool
}

func NewDeferredOutputCommand(displayMessage string) *DeferredOutputCommand {
	d := &DeferredOutputCommand{
		cmdBuffer:      newCmdBuffer(),
		displayMessage: displayMessage,
	}
	return d
}

func (d *DeferredOutputCommand) Command(cmd string, args ...string) *DeferredOutputCommand {
	d.cmdBuffer.clean()
	d.cmd = exec.Command(cmd, args...)
	d.cmd.Stdout = &d.cmdBuffer.stdOut
	d.cmd.Stderr = &d.cmdBuffer.stdErr
	return d
}

func (d *DeferredOutputCommand) RunNew(cmd string, args ...string) error {
	d.Command(cmd, args...)
	return d.Run()
}

func (d *DeferredOutputCommand) start() {
	if !d.hasStarted {
		d.hasStarted = true
		fmt.Print(d.displayMessage)
	}
}

func (d *DeferredOutputCommand) End(successful bool) {
	if successful {
		color.Green(" ✓")
	} else {
		fmt.Println()
		d.cmdBuffer.print()
	}
}

func (d *DeferredOutputCommand) Run() error {
	d.start()
	if err := d.cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (d *DeferredOutputCommand) Start() error {
	d.start()
	if err := d.cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d *DeferredOutputCommand) Wait() error {
	if err := d.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (d *DeferredOutputCommand) Output() ([]byte, error) {
	err := d.cmd.Run()
	return d.cmdBuffer.makeStdOut(), err
}

func (d *DeferredOutputCommand) StdinPipe() (io.WriteCloser, error) {
	return d.cmd.StdinPipe()
}
