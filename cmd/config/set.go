package config

import (
	"github.com/chengyumeng/khadijah/pkg/config"
	"github.com/spf13/cobra"
)

var APIkey string

var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Used to set some user configurations,such as `APIKey`.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(APIkey) > 0 {
			config.SetAPIKey(APIkey)
		}
	},
}

func init() {
	SetCmd.Flags().StringVarP(&APIkey, "apikey", "a", "", "In Wayne, APIKEY is primarily used for calls to OpenAPI.")
}
