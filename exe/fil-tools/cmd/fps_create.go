package cmd

import (
	"context"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	fpsCmd.AddCommand(fpsCreateCmd)
}

var fpsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create fps instance",
	Long:  `Create fps instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Creating fps instance...")
		s.Start()
		id, token, err := fcClient.FPS.Create(ctx)
		checkErr(err)
		s.Stop()
		Message("Instance created with id %s and token %s.", id, token)

	},
}
