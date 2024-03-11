package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zj-sh/mrd/util"
	"github.com/zohu/reg"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/user"
	"path"
	"runtime"
	"slices"
	"strings"
	"time"
)

const (
	DefaultChartFile    = "Mored.yaml"
	DefaultIndexVersion = "v1"
)

type buildOpts struct {
	*rootOpts
	oss  *ossOpts
	dist string
	push bool
}

type buildCmd struct {
	*buildOpts
	cmd *cobra.Command
}

func newBuildCmd(opts *rootOpts) *buildCmd {
	c := &buildCmd{
		buildOpts: &buildOpts{rootOpts: opts, oss: &ossOpts{rootOpts: opts}},
	}
	c.cmd = &cobra.Command{
		Use:              "build",
		Short:            "build kit or suite",
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			c.error("missing <kit|suite>")
			c.example("mrd build <kit|suite> [includes...] [flags...]")
		},
	}

	c.cmd.PersistentFlags().StringVarP(&c.dist, "dist", "d", "dist", "release directory.")
	c.cmd.PersistentFlags().BoolVarP(&c.push, "push", "p", false, "enable auto push to remote.")

	c.cmd.AddCommand(
		newBuildKitCmd(c.buildOpts).cmd,
		newSuiteCmd(c.buildOpts).cmd,
	)
	return c
}

type Dependency struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Remote  string `json:"remote,omitempty" yaml:"remote,omitempty"`
}
type Maintainer struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
	Home  string `json:"home,omitempty" yaml:"home,omitempty"`
}
type Metadata struct {
	Icon        string        `json:"icon,omitempty" yaml:"icon,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Digest      string        `json:"digest,omitempty" yaml:"digest,omitempty"`
	Keywords    []string      `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	Maintainers []*Maintainer `json:"maintainers,omitempty" yaml:"maintainers,omitempty"`
	Generated   time.Time     `json:"generated,omitempty" yaml:"generated,omitempty"`
}
type Chart struct {
	Name         string        `json:"name,omitempty" yaml:"name,omitempty"`
	FullName     string        `json:"fullName,omitempty" yaml:"fullName,omitempty"`
	Version      string        `json:"version,omitempty" yaml:"version,omitempty"`
	Command      string        `json:"command,omitempty" yaml:"command,omitempty"`
	MoredVersion string        `json:"moredVersion,omitempty" yaml:"moredVersion,omitempty"`
	Os           []string      `json:"os,omitempty,omitempty" yaml:"os,omitempty,omitempty"`
	Arch         []string      `json:"arch,omitempty" yaml:"arch,omitempty"`
	Effects      []int64       `json:"effects,omitempty" yaml:"effects,omitempty"`
	DepKits      []*Dependency `json:"depKits,omitempty" yaml:"depKits,omitempty"`
	DepSuites    []*Dependency `json:"depSuites,omitempty" yaml:"depSuites,omitempty"`
	Metadata     *Metadata     `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
type Author struct {
	Name      string    `json:"name,omitempty" yaml:"name,omitempty"`
	Platform  string    `json:"platform,omitempty" yaml:"platform,omitempty"`
	Ip        string    `json:"ip,omitempty" yaml:"ip,omitempty"`
	Generated time.Time `json:"generated,omitempty" yaml:"generated,omitempty"`
}
type Index struct {
	Version string              `json:"version,omitempty" yaml:"version,omitempty"`
	Suites  map[string][]*Chart `json:"suites,omitempty" yaml:"suites,omitempty"`
	Kits    map[string][]*Chart `json:"kits,omitempty" yaml:"kits,omitempty"`
	Authors []*Author           `json:"authors,omitempty" yaml:"authors,omitempty"`
}

func (c *buildOpts) index() Index {
	return Index{
		Version: DefaultIndexVersion,
		Kits:    make(map[string][]*Chart),
		Suites:  make(map[string][]*Chart),
		Authors: []*Author{},
	}
}
func (c *buildOpts) readRemoteIndex() *Index {
	index := c.index()
	bkt := c.oss.Bucket()
	ossName := fmt.Sprintf("%s/index.yaml", c.oss.prefix)
	r, err := bkt.GetObject(ossName)
	if err != nil && strings.Contains(err.Error(), "StatusCode=404") {
		color.Yellow("remote index does not exist, a new index will be created")
		return &index
	}
	defer r.Close()
	buf, err := io.ReadAll(r)
	c.hasErrExit("failed to read remote index", err)
	c.hasErrExit("failed to parse remote index", yaml.Unmarshal(buf, &index))
	return &index
}
func (c *buildOpts) mergeCharts(remote, local []*Chart) []*Chart {
	for _, l := range local {
		var exist bool
		for i, r := range remote {
			if l.Version == r.Version {
				remote[i] = l
				exist = true
				break
			}
		}
		if !exist {
			remote = append(remote, l)
		}
	}
	return remote
}
func (c *buildOpts) sortCharts(charts []*Chart) {
	slices.SortFunc(charts, func(a, b *Chart) int {
		if reg.Version(a.Version).HighThan(b.Version).B() {
			return -1
		}
		if reg.Version(a.Version).LowThan(b.Version).B() {
			return 1
		}
		return 0
	})
}
func (c *buildOpts) mergeAuthor(index *Index) {
	now := Author{
		Name:      "unknown",
		Ip:        util.NetIp(),
		Platform:  runtime.GOOS,
		Generated: time.Now(),
	}
	if u, err := user.Current(); err == nil {
		now.Name = util.FirstTruthValue(u.Username, u.Name)
	}
	var authors []*Author
	authors = append(authors, &now)
	for i := 0; i < 4; i++ {
		if i < len(index.Authors) {
			authors = append(authors, index.Authors[i])
		}
	}
	index.Authors = authors
}
func (c *buildOpts) mergeIndex(kits, suites map[string][]*Chart) {
	c.tips("loading remote index...")
	remote := c.readRemoteIndex()
	c.tips("merge index...")
	if kits != nil {
		for name, cts := range kits {
			if _, ok := remote.Kits[name]; ok {
				remote.Kits[name] = c.mergeCharts(remote.Kits[name], cts)
			} else {
				remote.Kits[name] = cts
			}
			c.sortCharts(remote.Kits[name])
		}
	}
	if suites != nil {
		for name, cts := range suites {
			if _, ok := remote.Suites[name]; ok {
				remote.Suites[name] = c.mergeCharts(remote.Suites[name], cts)
			} else {
				remote.Suites[name] = cts
			}
			c.sortCharts(remote.Suites[name])
		}
	}
	c.mergeAuthor(remote)
	d, err := yaml.Marshal(remote)
	c.hasErrExit("failed to parse index", err)
	filename := path.Join(c.dist, "index.yaml")
	c.hasErrExit("failed to create index", util.WriteFile(filename, d))

	c.tips("push index...")
	c.oss.Upload("", filename)
	c.success("push index success!")
}
func (c *buildOpts) readChart(filename string) (*Chart, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}
	d, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var chart Chart
	if err = yaml.Unmarshal(d, &chart); err != nil {
		return nil, err
	}
	return &chart, nil
}
func (c *buildOpts) chartFileName(args ...string) string {
	var name string
	for _, s := range args {
		s = strings.TrimSpace(s)
		s = util.CamelCaseToUnderscore(s)
		name = fmt.Sprintf("%s_%s", name, s)
	}
	return strings.TrimPrefix(name, "_")
}
func (c *buildOpts) verifyMust(ct *Chart) error {
	if reg.String(ct.Name).IsTruthAlphanumericUnderline().NotB() {
		return fmt.Errorf("name can only contain letters, numbers, underscores, the first can not be a number, maximum length of 128 digits")
	}
	if reg.String(ct.FullName).MaxLen(128).AllowEmpty().NotB() {
		return fmt.Errorf("full name maximum length of 128 digits")
	}
	if reg.Version(ct.Version).IsVersion().NotB() {
		return fmt.Errorf("incorrect version format, only supports x.x.x standard formats")
	}
	if reg.Version(ct.MoredVersion).IsVersionSupport().AllowEmpty().NotB() {
		return fmt.Errorf("mored version supports prefixes: (~) patch, (^) minor, (>=) greater than or equal to, (<=) less than or equal to, default >= 0.0.0")
	}
	for i, dep := range ct.DepKits {
		if reg.Version(dep.Version).IsVersionSupport().NotB() {
			return fmt.Errorf("dep kit %s version supports prefixes: (~) patch, (^) minor, (>=) greater than or equal to, (<=) less than or equal to, default >= 0.0.0", dep.Name)
		}
		if reg.String(dep.Remote).IsUrl().AllowEmpty().NotB() {
			return fmt.Errorf("dep kit %s repository address error, only domains starting with http(s):// are supported", dep.Name)
		}
		ct.DepKits[i].Remote = util.FirstTruthValue(dep.Remote, c.oss.Remote())
	}
	for i, dep := range ct.DepSuites {
		if reg.Version(dep.Version).IsVersionSupport().NotB() {
			return fmt.Errorf("dep suite %s version supports prefixes: (~) patch, (^) minor, (>=) greater than or equal to, (<=) less than or equal to, default >= 0.0.0", dep.Name)
		}
		if reg.String(dep.Remote).IsUrl().AllowEmpty().NotB() {
			return fmt.Errorf("dep suite %s repository address error, only domains starting with http(s):// are supported", dep.Name)
		}
		ct.DepSuites[i].Remote = util.FirstTruthValue(dep.Remote, c.oss.Remote())
	}
	return nil
}
