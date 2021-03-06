package config

import (
	"fmt"

	"github.com/gin-gonic/gin/json"
	"github.com/spf13/cobra"

	"github.com/chengyumeng/khadijah/pkg/config"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Used to show all user configurations.",
	Example: `khadijah config show`,
	Run: func(cmd *cobra.Command, args []string) {
		data, _ := json.MarshalIndent(config.GlobalOption, " ", " ")
		fmt.Println(string(data))
	},
}
