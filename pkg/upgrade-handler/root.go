package upgradehandler

import (
	"fmt"
	"os"

	"github.com/heathcliff26/kube-upgrade/pkg/version"

	"github.com/spf13/cobra"
)

const Name = "upgrade-handler"

func Execute() {
	cmd := NewUpgradeHandler()
	err := cmd.Execute()
	if err != nil {
		exitError(cmd, err)
	}
}

func NewUpgradeHandler() *cobra.Command {
	cobra.AddTemplateFunc(
		"ProgramName", func() string {
			return Name
		},
	)

	rootCmd := &cobra.Command{
		Use:   Name,
		Short: Name + " rebases the system and upgrades the node via kubeadm",
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

}

// Print the error information on stderr and exit with code 1
func exitError(cmd *cobra.Command, err error) {
	fmt.Fprintln(cmd.Root().ErrOrStderr(), "Fatal: "+err.Error())
	os.Exit(1)
}
