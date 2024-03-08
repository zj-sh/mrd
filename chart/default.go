package chart

import (
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

func NewToolkit(dir string) (*Chart, error) {
	chart, err := readYaml(dir)
	if err != nil {

	}
	return chart, nil
}
func NewSuite() *Chart {
	return &Chart{
		Name:         "",
		FullName:     "",
		Version:      "",
		Command:      "",
		MoredVersion: "",
		Os:           nil,
		Arch:         nil,
		Effects:      nil,
		DepTools:     nil,
		DepSuites:    nil,
		Metadata:     nil,
	}
}

func read(dir string) (*Chart, error) {
	filename := path.Join(dir, "Mored.yaml")
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

func isExist(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}
}
