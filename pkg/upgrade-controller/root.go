package upgradecontroller

import (
	"github.com/heathcliff26/kube-upgrade/pkg/upgrade-controller/controller"
	"github.com/heathcliff26/kube-upgrade/pkg/version"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

const Name = "upgrade-controller"

func Execute() {
	cmd := NewUpgradeController()
	err := cmd.Execute()
	if err != nil {
		klog.Fatalf("Failed to execute command: %v", err)
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
		klog.Fatalf("Failed to create controller: %v", err)
	}
	err = ctrl.Run()
	if err != nil {
		klog.Fatalf("Controller exited with error: %v", err)
	}
}
