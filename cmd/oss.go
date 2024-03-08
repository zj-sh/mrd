package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var endpoint string
var key string
var secret string

var ossCmd = &cobra.Command{
	Use:   "oss",
	Short: "oss config for repository.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("oss.endpoint", endpoint)
		viper.Set("oss.key", key)
		viper.Set("oss.secret", secret)
		SaveConfig()
	},
}

func init() {
	ossCmd.Flags().StringVarP(&endpoint, "endpoint", "p", "", "oss endpoint.")
	ossCmd.Flags().StringVarP(&key, "key", "k", "", "oss key.")
	ossCmd.Flags().StringVarP(&secret, "secret", "s", "", "oss secret.")

	_ = ossCmd.MarkFlagRequired("endpoint")
	_ = ossCmd.MarkFlagRequired("key")
	_ = ossCmd.MarkFlagRequired("secret")

	_ = viper.BindPFlag("oss.endpoint", ossCmd.Flags().Lookup("endpoint"))
	_ = viper.BindPFlag("oss.key", ossCmd.Flags().Lookup("key"))
	_ = viper.BindPFlag("oss.secret", ossCmd.Flags().Lookup("secret"))

	rootCmd.AddCommand(ossCmd)
}
