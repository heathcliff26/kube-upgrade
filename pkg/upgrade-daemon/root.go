package upgradedaemon

import (
	"fmt"
	"os"
	"syscall"

	"github.com/heathcliff26/kube-upgrade/pkg/upgrade-daemon/daemon"
	"github.com/heathcliff26/kube-upgrade/pkg/utils"
	"github.com/heathcliff26/kube-upgrade/pkg/version"

	"github.com/spf13/cobra"
)

const Name = "upgrade-daemon"

func Execute() {
	cmd := NewUpgradeDaemon()
	err := cmd.Execute()
	if err != nil {
		exitError(cmd, err)
	}
}

func NewUpgradeDaemon() *cobra.Command {
	cobra.AddTemplateFunc(
		"ProgramName", func() string {
			return Name
		},
	)

	rootCmd := &cobra.Command{
		Use:   Name,
		Short: Name + " runs as a daemon on the node and periodically checks for upgrades via rpm-ostree.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			run(cmd)
			return nil
		},
	}

	rootCmd.AddCommand(
		version.NewCommand(Name),
	)

	return rootCmd
}

func run(cmd *cobra.Command) {
	ns, err := utils.GetNamespace()
	if err != nil {
		exitError(cmd, err)
	}
	group, ok := os.LookupEnv("UPGRADE_GROUP")
	if !ok {
		exitError(cmd, fmt.Errorf("could not find the upgrade group"))
	}
	node, ok := os.LookupEnv("POD_NODE")
	if !ok {
		exitError(cmd, fmt.Errorf("could not find the node name"))
	}

	chroot, ok := os.LookupEnv("POD_CHROOT")
	if ok {
		err := syscall.Chroot(chroot)
		if err != nil {
			exitError(cmd, err)
		}
	}

	d, err := daemon.NewDaemon(Name, ns, group, node)
	if err != nil {
		exitError(cmd, err)
	}

	err = d.Run()
	if err != nil {
		exitError(cmd, err)
	}
}

// Print the error information on stderr and exit with code 1
func exitError(cmd *cobra.Command, err error) {
	fmt.Fprintln(cmd.Root().ErrOrStderr(), "Fatal: "+err.Error())
	os.Exit(1)
}
