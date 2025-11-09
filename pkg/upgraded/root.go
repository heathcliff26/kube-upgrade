package upgraded

import (
	"log/slog"
	"os"
	"os/user"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/daemon"
	"github.com/heathcliff26/kube-upgrade/pkg/version"

	"github.com/spf13/cobra"
)

const Name = "upgraded"

func Execute() {
	cmd := NewUpgraded()
	err := cmd.Execute()
	if err != nil {
		slog.Error("Command exited with error", "err", err)
		os.Exit(1)
	}
}

func NewUpgraded() *cobra.Command {
	cobra.AddTemplateFunc(
		"ProgramName", func() string {
			return Name
		},
	)

	rootCmd := &cobra.Command{
		Use:   Name,
		Short: Name + " daemon for keeping the system up-to-date",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg, err := cmd.Flags().GetString("config")
			if err != nil {
				slog.Error("Failed to parse config file flag", "err", err)
				os.Exit(1)
			}

			run(cfg)
		},
	}

	rootCmd.Flags().StringP("config", "c", "", "Path to config file")
	rootCmd.AddCommand(
		version.NewCommand(Name),
	)

	return rootCmd
}

func run(cfgPath string) {
	u, err := user.Current()
	if err != nil {
		slog.Error("Failed to check if running as root", "err", err)
		os.Exit(1)
	}
	if u.Username != "root" {
		slog.Error("Need to be root")
		os.Exit(1)
	}

	d, err := daemon.NewDaemon(cfgPath)
	if err != nil {
		slog.Error("Failed to create a new daemon", "err", err)
		os.Exit(1)
	}

	err = d.Run()
	if err != nil {
		slog.Error("Daemon exited with error", "err", err)
		os.Exit(1)
	}
}
