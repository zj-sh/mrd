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
	DefaultKitDist     = "kit"
	DefaultKitMainFile = "kit.sh"
)

type buildKitOpts struct {
	*buildOpts
}

type buildKitCmd struct {
	*buildKitOpts
	cmd *cobra.Command
}

func newBuildKitCmd(opts *buildOpts) *buildKitCmd {
	c := &buildKitCmd{
		buildKitOpts: &buildKitOpts{buildOpts: opts},
	}
	c.cmd = &cobra.Command{
		Use:   "kit",
		Short: "build kits.",
		Long: `kit's directory must have the Mored.yaml and kit.sh files, example:
  - kit.sh
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

type kitInfo struct {
	Src   string
	Chart *Chart
}

func (c *buildKitCmd) kit(dir string) (*Chart, error) {
	chart, err := c.readChart(path.Join(dir, DefaultChartFile))
	if err != nil {
		return nil, fmt.Errorf("load Mored.yaml failed. %s", err.Error())
	}
	if err = c.verifyMust(chart); err != nil {
		return nil, err
	}
	return &Chart{
		Name:         chart.Name,
		FullName:     util.FirstTruthValue(chart.FullName, chart.Name),
		Version:      chart.Version,
		MoredVersion: util.FirstTruthValue(chart.MoredVersion, ">=0.0.0"),
		Os:           util.FirstTruthValue(chart.Os, []string{"linux", "darwin"}),
		Arch:         util.FirstTruthValue(chart.Arch, []string{"amd64", "arm64"}),
		DepKits:      chart.DepKits,
		Metadata:     chart.Metadata,
	}, nil
}
func (c *buildKitCmd) search(includes []string) []string {
	var paths []string
	c.do("scanning kits...", func() {
		err := filepath.Walk(".", func(dir string, fi fs.FileInfo, err error) error {
			if !fi.IsDir() {
				return nil
			}
			main := path.Join(dir, DefaultKitMainFile)
			chart := path.Join(dir, DefaultChartFile)
			if !util.IsExisted(main) || !util.IsExisted(chart) {
				return nil
			}
			if len(includes) > 0 {
				for _, include := range includes {
					if strings.Contains(fi.Name(), include) {
						c.info("found kit %s at %s", filepath.Base(dir), dir)
						paths = append(paths, dir)
					}
				}
			} else {
				c.info("found kit %s at %s", filepath.Base(dir), dir)
				paths = append(paths, dir)
			}
			return nil
		})
		c.hasErrExit("scan failed", err)
		if len(paths) == 0 {
			c.exit("no buildable kits found")
		}
	})
	return paths
}
func (c *buildKitCmd) parse(paths []string) map[string]*kitInfo {
	var charts = make(map[string]*kitInfo)
	c.do("parsing kits...", func() {
		for _, p := range paths {
			chart, err := c.kit(p)
			if err != nil {
				c.warn("%s: %s", p, err.Error())
				continue
			}
			if _, ok := charts[chart.Name]; ok {
				c.warn("kit %s is exist %s %s", p, chart.Name, chart.Version)
				continue
			}
			charts[chart.Name] = &kitInfo{Src: p, Chart: chart}
		}
	})
	return charts
}
func (c *buildKitCmd) build(charts map[string]*kitInfo) *Index {
	index := c.index()
	c.do("building kits...", func() {
		dist := path.Join(c.dist, DefaultKitDist)
		_ = os.RemoveAll(dist)
		_ = os.MkdirAll(dist, os.ModePerm)
		for _, cf := range charts {
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
				index.Kits[chart.Name] = append(index.Kits[chart.Name], chart)
			}
		}
		c.success("build kits success!")
	})
	return &index
}
func (c *buildKitCmd) push(index *Index) {
	dist := path.Join(c.dist, DefaultKitDist)
	c.do("pushing kits...", func() {
		c.oss.Load()
		for _, charts := range index.Kits {
			for _, chart := range charts {
				gzName := c.chartFileName(chart.Name, chart.Version)
				gzFile := path.Join(dist, fmt.Sprintf("%s.tar.gz", gzName))
				c.oss.Upload(DefaultKitDist, gzFile)
				c.info("%s success!", chart.Name)
			}
		}
		c.mergeIndex(index.Kits, nil)
		c.success("push success!")
	})
}
