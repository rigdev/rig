package instance

import (
	"context"
	"errors"
	"io"
	"strings"

	"connectrpc.com/connect"
	"github.com/moby/term"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

var defaultEscapeKeys = []byte{16, 17}

func listen(
	stream *connect.BidiStreamForClient[capsule.ExecuteRequest, capsule.ExecuteResponse],
	stdout, stderr io.Writer,
) chan error {
	doneChan := make(chan error, 1)
	go func() {
		for {
			resp, err := stream.Receive()
			if err != nil {
				doneChan <- err
				return
			}
			switch v := resp.Response.(type) {
			case *capsule.ExecuteResponse_Stdout:
				stdout.Write(v.Stdout.GetData())
			case *capsule.ExecuteResponse_Stderr:
				stderr.Write(v.Stderr.GetData())
			}
		}
	}()
	return doneChan
}

func (c *Cmd) exec(ctx context.Context, cmd *cobra.Command, args []string) error {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	instance, err := c.provideInstanceID(ctx, capsule_cmd.CapsuleID, arg, cmd.ArgsLenAtDash())
	if err != nil {
		return err
	}

	command, arguments := parseArgs(cmd, args)

	var resize *capsule.ExecuteRequest_Resize
	var ttyStruct *Tty
	if tty {
		ttyStruct = NewTty()
		width, height := ttyStruct.GetTtySize()
		resize = &capsule.ExecuteRequest_Resize{
			Height: height,
			Width:  width,
		}
	}

	start := &capsule.ExecuteRequest{
		Request: &capsule.ExecuteRequest_Start_{
			Start: &capsule.ExecuteRequest_Start{
				CapsuleId:   capsule_cmd.CapsuleID,
				InstanceId:  instance,
				Command:     command,
				Arguments:   arguments,
				Tty:         resize,
				Interactive: interactive,
			},
		},
		ProjectId:     c.Cfg.GetProject(),
		EnvironmentId: base.Flags.Environment,
	}

	stream := c.Rig.Capsule().Execute(ctx)

	defer func() {
		if tty {
			err := ttyStruct.RestoreTerminal()
			if err != nil {
				cmd.Println(err)
			}
		}
		//nolint:errcheck
		stream.CloseRequest()
		//nolint:errcheck
		stream.CloseResponse()
	}()

	err = stream.Send(start)
	if err != nil {
		return err
	}

	stdinClose, stdout, stderr := term.StdStreams()
	var stdin io.Reader = stdinClose

	if interactive {
		outChan := listen(stream, stdout, stderr)
		inChan := make(chan error, 1)
		interruptChan := make(chan error, 1)
		if ttyStruct != nil {
			stdin = term.NewEscapeProxy(stdin, defaultEscapeKeys)
			err := ttyStruct.SetTtyTerminal()
			if err != nil {
				return err
			}
			err = ttyStruct.MonitorSize()
			if err != nil {
				return err
			}
			go func() {
				for size := range ttyStruct.resizeChan {
					err := stream.Send(&capsule.ExecuteRequest{
						Request: &capsule.ExecuteRequest_Resize_{
							Resize: &capsule.ExecuteRequest_Resize{
								Height: size.height,
								Width:  size.width,
							},
						},
					})
					if err != nil {
						inChan <- err
						return
					}
				}
			}()
			err = ttyStruct.MonitorInterrupt(interruptChan)
			if err != nil {
				return err
			}

		}
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := stdin.Read(buf)
				if n > 0 {
					err := stream.Send(&capsule.ExecuteRequest{
						Request: &capsule.ExecuteRequest_Stdin{
							Stdin: &capsule.StreamData{
								Data: buf[:n],
							},
						},
					})
					if err != nil {
						inChan <- err
						return
					}
				}
				if errors.Is(err, io.EOF) {
					err := stream.Send(&capsule.ExecuteRequest{
						Request: &capsule.ExecuteRequest_Stdin{
							Stdin: &capsule.StreamData{
								Data:   buf[:n],
								Closed: true,
							},
						},
					})
					if err != nil {
						inChan <- err
						return
					}
					return
				} else if err != nil {
					inChan <- err
					return
				}
			}
		}()
		select {
		case err := <-interruptChan:
			return err
		case err := <-outChan:
			if errors.Is(err, io.EOF) {
				return nil
			} else {
				return err
			}
		case err := <-inChan:
			if errors.Is(err, io.EOF) {
				// in is finished wait for out to finish.
				select {
				case err := <-outChan:
					return err
				case err := <-interruptChan:
					return err
				}
			} else {
				return err
			}
		}
	} else {
		outChan := listen(stream, stdout, stderr)
		err := <-outChan
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
}

func parseArgs(cmd *cobra.Command, args []string) (string, []string) {
	var command string
	var arguments []string
	var err error
	if cmd.ArgsLenAtDash() == -1 || len(args) <= 1 {
		command, err = common.PromptInput("command", common.ValidateNonEmptyOpt)
		if err != nil {
			return "", nil
		}

		argString, err := common.PromptInput("Arguments", common.ValidateAllOpt)
		if err != nil {
			return "", nil
		}
		if argString != "" {
			arguments = strings.Split(argString, " ")
		}
	} else {
		command = args[cmd.ArgsLenAtDash()]
		arguments = args[cmd.ArgsLenAtDash()+1:]
	}
	return command, arguments
}
