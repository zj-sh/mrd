package cmd

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
)

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		IfErrExit(err)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName(".config/mored/config")
	}
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

func SaveConfig() {
	file := viper.ConfigFileUsed()
	if file == "" {
		home, err := homedir.Dir()
		IfErrExit(err)
		file = path.Join(home, "/.config/mored/config")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			dir := filepath.Dir(file)
			IfErrExit(os.MkdirAll(dir, os.ModePerm))
			f, err := os.Create(file)
			IfErrExit(err)
			_ = f.Close()
		}
	}
	IfErrExit(viper.WriteConfigAs(file))
}
