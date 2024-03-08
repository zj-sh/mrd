package cmd

import (
	"github.com/fatih/color"
	"os"
	"reflect"
)

func IfDryRun() {
	if dry {
		color.Yellow("[WARNING] in dry-run state.")
	}
}

func IfErrExit(err error) {
	if err != nil {
		color.Red(err.Error())
		os.Exit(0)
	}
}

func FirstTruthValue[T any](args ...T) T {
	for _, item := range args {
		if !reflect.ValueOf(item).IsZero() {
			return item
		}
	}
	return args[0]
}
