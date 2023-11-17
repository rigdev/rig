package instance

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func listen(stream *connect.BidiStreamForClient[capsule.ExecuteRequest, capsule.ExecuteResponse]) chan error {
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
				os.Stdout.Write(v.Stdout.GetData())
			case *capsule.ExecuteResponse_Stderr:
				os.Stderr.Write(v.Stderr.GetData())
			}
		}
	}()
	return doneChan
}

func (c Cmd) shell(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	instance, err := c.provideInstanceID(ctx, capsule_cmd.CapsuleID, arg, cmd.ArgsLenAtDash())
	if err != nil {
		return err
	}

	command, arguments := parseArgs(cmd, args)

	start := &capsule.ExecuteRequest{
		Request: &capsule.ExecuteRequest_Start_{
			Start: &capsule.ExecuteRequest_Start{
				CapsuleId:  capsule_cmd.CapsuleID,
				InstanceId: instance,
				Command:    command,
				Arguments:  arguments,
				// Tty:         tty,
				Interactive: interactive,
			},
		},
	}

	stream := c.Rig.Capsule().Execute(ctx)

	defer func() {
		stream.CloseRequest()
		stream.CloseResponse()
	}()

	err = stream.Send(start)
	if err != nil {
		return err
	}

	if interactive {
		outChan := listen(stream)
		inChan := make(chan error, 1)
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := os.Stdin.Read(buf)
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
		case err := <-outChan:
			if errors.Is(err, io.EOF) {
				return nil
			} else {
				return err
			}
		case err := <-inChan:
			if errors.Is(err, io.EOF) {
				// in is finished wait for out to finish.
				return <-outChan
			} else {
				return err
			}
		}
	} else {
		outChan := listen(stream)
		err := <-outChan
		if errors.Is(err, io.EOF) {
			return nil
		} else {
			return err
		}
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
