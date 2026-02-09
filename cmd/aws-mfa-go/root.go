package main

import (
	"os"

	"github.com/jlis/aws-mfa-go/internal/app"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Execute runs the root command.
func Execute() {
	rootCmd := newRootCmd()
	cobra.CheckErr(rootCmd.Execute())
}

func newRootCmd() *cobra.Command {
	var (
		profile         string
		device          string
		durationSeconds int
		token           string
		force           bool
		longTermSuffix  string
		shortTermSuffix string
		credentialsFile string
	)

	cmd := &cobra.Command{
		Use:          "aws-mfa-go",
		Short:        "Refresh AWS credentials using MFA (writes to ~/.aws/credentials)",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			deps := app.DefaultDeps()
			deps.Stdout = cmd.OutOrStdout()
			deps.Stderr = cmd.ErrOrStderr()
			deps.Stdin = os.Stdin

			return app.Run(cmd.Context(), app.RunInputs{
				Inputs: app.Inputs{
					Profile:                profile,
					ProfileChanged:         flagChanged(flags, "profile"),
					Device:                 device,
					DeviceChanged:          flagChanged(flags, "device"),
					DurationSeconds:        durationSeconds,
					DurationSecondsChanged: flagChanged(flags, "duration"),
					Token:                  token,
					TokenChanged:           flagChanged(flags, "token"),
					Force:                  force,
					LongTermSuffix:         longTermSuffix,
					ShortTermSuffix:        shortTermSuffix,
					CredentialsFile:        credentialsFile,
				},
			}, deps)
		},
	}

	cmd.Flags().StringVar(&profile, "profile", "", "AWS profile name (env: AWS_PROFILE, default: default)")
	cmd.Flags().StringVar(&device, "device", "", "MFA device ARN/serial (env: MFA_DEVICE, or aws_mfa_device in long-term section)")
	cmd.Flags().IntVar(&durationSeconds, "duration", 0, "STS session duration seconds (env: MFA_STS_DURATION, default: 43200)")
	cmd.Flags().StringVar(&token, "token", "", "MFA token code (6 digits). If omitted, prompts on stdin")
	cmd.Flags().BoolVar(&force, "force", false, "Refresh credentials even if still valid")
	cmd.Flags().StringVar(&longTermSuffix, "long-term-suffix", "long-term", "Suffix for long-term section (<profile>-<suffix>). Use 'none' for <profile>")
	cmd.Flags().StringVar(&shortTermSuffix, "short-term-suffix", "none", "Suffix for short-term section (<profile>-<suffix>). Use 'none' for <profile>")
	cmd.Flags().StringVar(&credentialsFile, "credentials-file", "~/.aws/credentials", "Path to shared credentials file")

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	return cmd
}

func flagChanged(flags *pflag.FlagSet, name string) bool {
	f := flags.Lookup(name)
	return f != nil && f.Changed
}
