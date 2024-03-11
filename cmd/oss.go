package cmd

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zohu/reg"
	"path"
)

type ossOpts struct {
	*rootOpts
	domain   string
	endpoint string
	key      string
	secret   string
	bucket   string
	prefix   string
}

func (s *ossOpts) Load() {
	s.domain = viper.GetString("oss.domain")
	s.endpoint = viper.GetString("oss.endpoint")
	s.key = viper.GetString("oss.key")
	s.secret = viper.GetString("oss.secret")
	s.bucket = viper.GetString("oss.bucket")
	s.prefix = viper.GetString("oss.prefix")
	if s.endpoint == "" || s.key == "" || s.secret == "" || s.bucket == "" || s.prefix == "" {
		s.exit("OSS configuration is incomplete, please use the mrd oss [flags...] command. command.")
	}
}
func (s *ossOpts) Remote() string {
	return fmt.Sprintf("%s/%s", s.domain, s.prefix)
}
func (s *ossOpts) Bucket() *oss.Bucket {
	clt, err := oss.New(s.endpoint, s.key, s.secret)
	s.hasErrExit("failed to initialize OSS", err)
	bkt, err := clt.Bucket(s.bucket)
	s.hasErrExit("failed to initialize OSS bucket", err)
	return bkt
}
func (s *ossOpts) Upload(dir string, files ...string) {
	bkt := s.Bucket()
	ossName := s.prefix
	if dir != "" {
		ossName = fmt.Sprintf("%s/%s", ossName, dir)
	}
	for _, f := range files {
		ossName = fmt.Sprintf("%s/%s", ossName, path.Base(f))
		s.hasErrExit("push to remote failed", bkt.PutObjectFromFile(ossName, f))
	}
}

type ossCmd struct {
	*ossOpts
	cmd *cobra.Command
}

func newOssCmd(opts *rootOpts) *ossCmd {
	c := &ossCmd{
		ossOpts: &ossOpts{rootOpts: opts},
	}
	c.cmd = &cobra.Command{
		Use:   "oss",
		Short: "oss config for repository.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if reg.String(c.domain).IsUrl().NotB() {
				c.domain = fmt.Sprintf("https://%s.%s", c.bucket, c.endpoint)
			}
			c.save()
		},
	}
	c.cmd.Flags().StringVarP(&c.domain, "domain", "d", "", "oss domain.")
	c.cmd.Flags().StringVarP(&c.endpoint, "endpoint", "p", "", "oss endpoint.")
	c.cmd.Flags().StringVarP(&c.key, "key", "k", "", "oss key.")
	c.cmd.Flags().StringVarP(&c.secret, "secret", "s", "", "oss secret.")
	c.cmd.Flags().StringVarP(&c.bucket, "bucket", "b", "", "oss bucket.")
	c.cmd.Flags().StringVarP(&c.prefix, "prefix", "", "repo", "oss prefix")

	_ = c.cmd.MarkFlagRequired("endpoint")
	_ = c.cmd.MarkFlagRequired("key")
	_ = c.cmd.MarkFlagRequired("secret")
	_ = c.cmd.MarkFlagRequired("bucket")

	_ = viper.BindPFlag("oss.domain", c.cmd.Flags().Lookup("domain"))
	_ = viper.BindPFlag("oss.endpoint", c.cmd.Flags().Lookup("endpoint"))
	_ = viper.BindPFlag("oss.key", c.cmd.Flags().Lookup("key"))
	_ = viper.BindPFlag("oss.secret", c.cmd.Flags().Lookup("secret"))
	_ = viper.BindPFlag("oss.bucket", c.cmd.Flags().Lookup("bucket"))
	_ = viper.BindPFlag("oss.prefix", c.cmd.Flags().Lookup("prefix"))
	return c
}

func (c *ossCmd) save() {
	viper.Set("oss.domain", c.domain)
	viper.Set("oss.endpoint", c.endpoint)
	viper.Set("oss.key", c.key)
	viper.Set("oss.secret", c.secret)
	viper.Set("oss.bucket", c.bucket)
	viper.Set("oss.prefix", c.prefix)
	c.saveConfig()
}
