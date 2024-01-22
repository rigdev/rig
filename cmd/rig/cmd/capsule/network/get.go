package network

import (
	"context"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	n, err := capsule_cmd.GetCurrentNetwork(ctx, c.Rig, c.Cfg)
	if errors.IsNotFound(err) {
		cmd.Println("No interfaces are set")
		return nil
	} else if err != nil {
		return err
	}

	interfaces := n.GetInterfaces()

	if len(args) > 0 {
		found := false
		for _, i := range interfaces {
			if i.Name == args[0] {
				interfaces = []*capsule.Interface{i}
				break
			}
		}
		if !found {
			return fmt.Errorf("interface %v not found", args[0])
		}
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(n, flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Interfaces (%v)", len(n.GetInterfaces())), "Name", "Port", "Public"})
	for n, i := range interfaces {
		if !i.GetPublic().GetEnabled() {
			t.AppendRow(table.Row{fmt.Sprintf("#%v", n), i.GetName(), i.GetPort(), "-"})
			t.AppendSeparator()
			continue
		}

		switch v := i.GetPublic().GetMethod().GetKind().(type) {
		case *capsule.RoutingMethod_Ingress_:
			t.AppendRow(table.Row{fmt.Sprintf("#%v", n), i.GetName(), i.GetPort(), "Ingress"})
			t.AppendRow(table.Row{"Host", "", "", v.Ingress.GetHost()})
			t.AppendRow(table.Row{"Paths", "", "", strings.Join(v.Ingress.GetPaths(), " ")})
			t.AppendRow(table.Row{"TLS", "", "", v.Ingress.GetTls()})
		case *capsule.RoutingMethod_LoadBalancer_:
			t.AppendRow(table.Row{fmt.Sprintf("#%v", n), i.GetName(), i.GetPort(), "LoadBalancer"})
			t.AppendRow(table.Row{"Public Port", "", "", v.LoadBalancer.GetPort()})
		default:
			t.AppendRow(table.Row{"Public", "Unknown"})
		}
		t.AppendSeparator()
	}
	cmd.Println(t.Render())
	return nil
}
