package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
)

type rootOpts struct {
	dry     bool
	version string
	conf    string
}
type rootCmd struct {
	*rootOpts
	cmd *cobra.Command
}

func newRootCmd(version string) *rootCmd {
	c := &rootCmd{
		rootOpts: &rootOpts{version: version},
	}
	c.cmd = &cobra.Command{
		Use:           "mrd",
		Short:         "mored command line tool",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	c.cmd.PersistentFlags().BoolVarP(&c.dry, "dry-run", "", false, "dry run mode.")
	c.cmd.PersistentFlags().StringVarP(&c.conf, "config", "c", "", "config file (default is $HOME/.config/mored/config.toml)")

	c.cmd.AddCommand(
		newVersionCmd(c.rootOpts).cmd,
		newOssCmd(c.rootOpts).cmd,
		newBuildCmd(c.rootOpts).cmd,
	)

	return c
}

func Execute(version string) {
	r := newRootCmd(version)
	cobra.OnInitialize(r.initConfig)
	if err := r.cmd.Execute(); err != nil {
		color.Red("🔴 %s", err)
		os.Exit(0)
	}
}

func (r *rootOpts) tips(format string, a ...interface{}) {
	color.Blue("🚀 "+format, a...)
}
func (r *rootOpts) info(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("=> "+format, a...))
}
func (r *rootOpts) warn(format string, a ...interface{}) {
	color.Yellow(fmt.Sprintf(">> "+format, a...))
}
func (r *rootOpts) example(format string, a ...interface{}) {
	color.Magenta("👉 "+format, a...)
}
func (r *rootOpts) success(format string, a ...interface{}) {
	color.Green("✅ "+format, a...)
}
func (r *rootOpts) error(format string, a ...interface{}) {
	color.Red("🔴 "+format, a...)
}
func (r *rootOpts) exit(format string, a ...interface{}) {
	r.error(format, a...)
	os.Exit(0)
}
func (r *rootOpts) hasErrExit(tip string, err error) {
	if err != nil {
		if tip != "" {
			r.exit("%s: %s", tip, err.Error())
		} else {
			r.exit(err.Error())
		}
	}
}
func (r *rootOpts) do(tip string, fc func()) {
	if r.dry {
		color.Magenta("🌈 %s", tip)
	} else {
		r.tips(tip)
		fc()
	}
}
func (r *rootOpts) initConfig() {
	if r.conf != "" {
		viper.SetConfigFile(r.conf)
	} else {
		if home, err := homedir.Dir(); err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName(".config/mored/config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		r.warn("load config failed: %s", err.Error())
	}
}
func (r *rootOpts) saveConfig() {
	file := viper.ConfigFileUsed()
	if file == "" {
		home, err := homedir.Dir()
		r.hasErrExit("access user dir failed", err)
		file = path.Join(home, "/.config/mored/config")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			dir := filepath.Dir(file)
			r.hasErrExit("create config file failed", os.MkdirAll(dir, os.ModePerm))
			f, err := os.Create(file)
			r.hasErrExit("create config file failed", err)
			_ = f.Close()
		}
	}
	r.hasErrExit("save config failed", viper.WriteConfigAs(file))
}
