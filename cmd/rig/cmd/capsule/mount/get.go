package mount

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func get(ctx context.Context, args []string, cmd *cobra.Command, rc rig.Client) error {
	r, err := capsule_cmd.GetCurrentRollout(ctx, rc)
	if err != nil {
		return err
	}

	if len(r.GetConfig().GetConfigFiles()) == 0 {
		cmd.Println("No config files mounted")
		return nil
	}

	files := r.GetConfig().GetConfigFiles()

	if len(args) > 0 {
		found := false
		for _, f := range files {
			if f.GetPath() == args[0] {
				files = []*capsule.ConfigFile{f}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("config file %v not found", args[0])
		}
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Config Files (%v)", len(files)), "Path", "size", "Age", "Created by"})
	for i, f := range files {

		t.AppendRow(table.Row{
			fmt.Sprintf("#%v", i),
			f.GetPath(),
			formatBytesSize(len(f.GetContent())),
			time.Since(f.GetUpdatedAt().AsTime()).Truncate(time.Second),
			f.GetUpdatedBy().GetPrintableName(),
		})

		if cmd.Flags().Changed("download") {
			dowloadFile(ctx, f, dstPath)
		}
	}

	cmd.Println(t.Render())

	return nil
}

func formatBytesSize(numBytes int) string {
	const unit = 1024
	if numBytes < unit {
		return fmt.Sprintf("%d B", numBytes)
	}
	div, exp := int64(unit), 0
	for n := numBytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(numBytes)/float64(div), "KMGTPE"[exp])
}

func dowloadFile(ctx context.Context, f *capsule.ConfigFile, dstPath string) error {
	path := path.Join(dstPath, path.Base(f.Path))
	if err := os.WriteFile(path, f.GetContent(), 0644); err != nil {
		return err
	}

	return nil
}
