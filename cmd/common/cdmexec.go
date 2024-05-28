package common

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/fatih/color"
)

type bytesWithSource struct {
	bytes    []byte
	isStdOut bool
}

type buffer struct {
	lock  sync.Mutex
	bytes []bytesWithSource
	idx   int
}

func (b *buffer) next() ([]byte, bool, bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.idx >= len(b.bytes) {
		return nil, false, false
	}
	bb := b.bytes[b.idx]
	b.idx++
	return bb.bytes, bb.isStdOut, true
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

	verbose bool
	cancel  context.CancelFunc
	ctx     context.Context
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
	for {
		bytes, isStdOut, ok := c.buffer.next()
		if !ok {
			break
		}
		if isStdOut {
			os.Stdout.WriteString(string(bytes))
		} else {
			os.Stderr.WriteString(string(bytes))
		}
	}
}

func (c *cmdBuffer) end() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.verbose {
		c.print()
	}
}

func (c *cmdBuffer) start() {
	if !c.verbose {
		return
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(250 * time.Millisecond):
			}
			c.print()
		}
	}()
}

func (c *cmdBuffer) clean() {
	c.buffer.lock.Lock()
	defer c.buffer.lock.Unlock()
	c.buffer.bytes = nil
	c.buffer.idx = 0
}

func newCmdBuffer(verbose bool) *cmdBuffer {
	b := &cmdBuffer{buffer: buffer{}, verbose: verbose}
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

type option struct {
	verbose bool
}

type Option func(*option)

func Verbose(v bool) Option {
	return func(o *option) {
		o.verbose = v
	}
}

type DeferredOutputCommand struct {
	cmdBuffer      *cmdBuffer
	cmd            *exec.Cmd
	displayMessage string
	hasStarted     bool
}

func NewDeferredOutputCommand(displayMessage string, opts ...Option) *DeferredOutputCommand {
	o := option{}
	for _, opt := range opts {
		opt(&o)
	}
	d := &DeferredOutputCommand{
		cmdBuffer:      newCmdBuffer(o.verbose),
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
		d.cmdBuffer.start()
	}
}

func (d *DeferredOutputCommand) End(successful bool) {
	if successful {
		color.Green(" ✓")
	} else {
		fmt.Println()
		d.cmdBuffer.print()
	}
	d.cmdBuffer.end()
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
