package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zj-sh/mrd/util"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultSuiteDist     = "suite"
	DefaultSuiteMainFile = "suite"
	DefaultSuitePattern  = `^suite(\.\w+)?$`
)

type suiteOpts struct {
	*buildOpts
}
type suiteCmd struct {
	*suiteOpts
	cmd *cobra.Command
}

func newSuiteCmd(opts *buildOpts) *suiteCmd {
	c := &suiteCmd{
		suiteOpts: &suiteOpts{buildOpts: opts},
	}
	c.cmd = &cobra.Command{
		Use:   "suite",
		Short: "build suites.",
		Long: `suite's directory must have the Mored.yaml and suite(.*) files, example:
  - suite [suite.jar] [suite.py] [suite.sh] [suite.exe] [suite.js]
  - Mored.yaml
  - ...`,
		Run: func(cmd *cobra.Command, includes []string) {
			paths := c.search(includes)
			charts := c.parse(paths)
			index := c.build(charts)
			if c.buildOpts.push {
				c.push(index)
			}
		},
	}
	return c
}

type suiteInfo struct {
	Src   string
	Chart *Chart
}

func (c *suiteCmd) suite(dir string) (*Chart, error) {
	chart, err := c.readChart(path.Join(dir, DefaultChartFile))
	if err != nil {
		return nil, fmt.Errorf("load Mored.yaml failed. %s", err.Error())
	}
	if err = c.verifyMust(chart); err != nil {
		return nil, err
	}
	if !strings.Contains(chart.Command, "{file}") {
		return nil, fmt.Errorf("command must contain the {file} placeholder")
	}
	if len(chart.Effects) == 0 {
		return nil, fmt.Errorf("effects is required")
	}
	return &Chart{
		Name:         chart.Name,
		FullName:     util.FirstTruthValue(chart.FullName, chart.Name),
		Version:      chart.Version,
		MoredVersion: util.FirstTruthValue(chart.MoredVersion, ">=0.0.0"),
		Command:      util.FirstTruthValue(chart.Command, "{file}"),
		Effects:      chart.Effects,
		Os:           util.FirstTruthValue(chart.Os, []string{"linux", "darwin"}),
		Arch:         util.FirstTruthValue(chart.Arch, []string{"amd64", "arm64"}),
		DepKits:      chart.DepKits,
		DepSuites:    chart.DepSuites,
		Metadata:     chart.Metadata,
	}, nil
}
func (c *suiteCmd) search(includes []string) []string {
	var paths []string
	c.do("scanning suite...", func() {
		err := filepath.Walk(".", func(dir string, fi fs.FileInfo, err error) error {
			if !fi.IsDir() {
				return nil
			}
			if !util.IsExisted(path.Join(dir, DefaultChartFile)) {
				return nil
			}
			if !util.HasPatternFile(dir, DefaultSuitePattern) {
				return nil
			}
			if len(includes) > 0 {
				for _, include := range includes {
					if strings.Contains(fi.Name(), include) {
						c.info("found suite %s at %s", filepath.Base(dir), dir)
						paths = append(paths, dir)
					}
				}
			} else {
				c.info("found suite %s at %s", filepath.Base(dir), dir)
				paths = append(paths, dir)
			}
			return nil
		})
		c.hasErrExit("scan failed", err)
		if len(paths) == 0 {
			c.exit("no buildable suite found")
		}
	})
	return paths
}
func (c *suiteCmd) parse(paths []string) map[string]*suiteInfo {
	suites := make(map[string]*suiteInfo)
	c.do("parsing suites...", func() {
		for _, p := range paths {
			chart, err := c.suite(p)
			if err != nil {
				c.warn("%s: %s", p, err.Error())
				continue
			}
			if _, ok := suites[chart.Name]; ok {
				c.warn("suite %s is exist %s %s", p, chart.Name, chart.Version)
				continue
			}
			suites[p] = &suiteInfo{Src: p, Chart: chart}
		}
	})
	return suites
}
func (c *suiteCmd) build(suites map[string]*suiteInfo) *Index {
	index := c.index()
	c.do("building suites...", func() {
		dist := path.Join(c.dist, DefaultSuiteDist)
		_ = os.RemoveAll(dist)
		_ = os.MkdirAll(dist, os.ModePerm)
		for _, cf := range suites {
			chart := cf.Chart
			gzName := c.chartFileName(chart.Name, chart.Version)
			gzFile := path.Join(dist, fmt.Sprintf("%s.tar.gz", gzName))
			if err := util.Compress(cf.Src, gzFile); err != nil {
				c.warn("%s build failed: %s", cf.Src, err.Error())
			} else {
				if chart.Metadata == nil {
					chart.Metadata = &Metadata{}
				}
				chart.Metadata.Digest = util.FileDigest(gzFile)
				chart.Metadata.Generated = time.Now()
				c.success("%s build success %s digest:%s", cf.Src, gzFile, chart.Metadata.Digest)
				index.Suites[chart.Name] = append(index.Suites[chart.Name], chart)
			}
		}
		c.success("build suites success")
	})
	return &index
}
func (c *suiteCmd) push(index *Index) {
	dist := path.Join(c.dist, DefaultSuiteDist)
	c.do("pushing suites...", func() {
		c.oss.Load()
		for _, charts := range index.Suites {
			for _, chart := range charts {
				gzName := c.chartFileName(chart.Name, chart.Version)
				gzFile := path.Join(dist, fmt.Sprintf("%s.tar.gz", gzName))
				c.oss.Upload(DefaultSuiteDist, gzFile)
				c.info("%s success!", chart.Name)
			}
		}
		c.mergeIndex(nil, index.Suites)
		c.success("push success!")
	})
}
