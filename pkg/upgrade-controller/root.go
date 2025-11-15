package upgradecontroller

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/heathcliff26/kube-upgrade/pkg/upgrade-controller/controller"
	"github.com/heathcliff26/kube-upgrade/pkg/version"

	"github.com/spf13/cobra"
)

const Name = "upgrade-controller"

func Execute() {
	cmd := NewUpgradeController()
	err := cmd.Execute()
	if err != nil {
		fatalf("Failed to execute command: %v", err)
	}
}

func NewUpgradeController() *cobra.Command {
	cobra.AddTemplateFunc(
		"ProgramName", func() string {
			return Name
		},
	)

	rootCmd := &cobra.Command{
		Use:   Name,
		Short: Name + " runs the controller to orchestrate cluster wide kubernetes upgrades.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			run()
			return nil
		},
	}

	rootCmd.AddCommand(
		version.NewCommand(Name),
	)

	return rootCmd
}

func run() {
	ctrl, err := controller.NewController(Name)
	if err != nil {
		fatalf("Failed to create controller: %v", err)
	}
	err = ctrl.Run()
	if err != nil {
		fatalf("Controller exited with error: %v", err)
	}
}

func fatalf(format string, args ...interface{}) {
	slog.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}
