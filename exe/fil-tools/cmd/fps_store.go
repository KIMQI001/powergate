package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	fpsStoreCmd.Flags().StringP("cid", "c", "", "cid of the data to pin")
	fpsStoreCmd.Flags().StringP("token", "t", "", "wallet address used to store the data")

	fpsCmd.AddCommand(fpsStoreCmd)
}

var fpsStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Store data in fps",
	Long:  `Store data in fps`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		token := viper.GetString("token")
		cidString := viper.GetString("cid")

		if token == "" {
			Fatal(errors.New("get requires token"))
		}
		ctx = context.WithValue(ctx, authKey("fpstoken"), token)

		if cidString == "" {
			Fatal(errors.New("store command needs a cid"))
		}

		c, err := cid.Parse(cidString)
		checkErr(err)

		s := spin.New("%s Pinning specified cid...")
		s.Start()
		err = fcClient.FPS.Store(ctx, c)
		s.Stop()
		checkErr(err)
	},
}
