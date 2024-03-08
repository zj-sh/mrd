package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var toolCmd = &cobra.Command{
	Use:   "toolkit",
	Short: "Generate toolkit repository based from shell scripts.",
	Long: `directory structure:
<NAME>:
- main.sh
- other.sh
- Mored.yaml`,
	Example: "mrd toolkit <NAME...>",
	Run: func(cmd *cobra.Command, includes []string) {
		IfDryRun()
		toolDist := path.Join(dist, "toolkits")
		_ = os.RemoveAll(toolDist)
		_ = os.MkdirAll(toolDist, os.ModePerm)
		fmt.Println(includes)
		var src []string
		IfErrExit(filepath.Walk(".", func(filename string, fi os.FileInfo, err error) error {
			if fi.IsDir() {
				main := path.Join(filename, "main.sh")
				chart := path.Join(filename, "Mored.yaml")
				if IsExisted(main) && IsExisted(chart) {
					if len(includes) > 0 {
						for _, include := range includes {
							if strings.HasPrefix(fi.Name(), include) {
								src = append(src, filename)
							}
						}
					} else {
						// default include all
						src = append(src, filename)
					}
				}
			}
			return nil
		}))
		fmt.Println(src)
		for _, s := range src {
			name := path.Base(s)
			dest := path.Join(toolDist, fmt.Sprintf("%s.tar.gz", name))
			if err := Compress(s, dest); err != nil {
				color.Red("build toolkit %s failed: %s", name, err.Error())
			}
		}
	},
}

func init() {
	toolCmd.Flags().StringVarP(&dist, "dist", "d", "./dist", "destination directory.")
	toolCmd.Flags().BoolVarP(&push, "push", "p", false, "auto push remote.")
	toolCmd.Flags().BoolVarP(&reset, "reset", "r", false, "reset remote index.")

	rootCmd.AddCommand(toolCmd)
}
