package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func IsExisted(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
func Compress(src, dest string, exclude ...string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	return filepath.Walk(src, func(filename string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			for _, ex := range exclude {
				if ok, _ := regexp.MatchString(ex, fi.Name()); ok {
					return nil
				}
			}
			color.Blue("compressing: %s %s", Filesize(fi.Size()), filename)
			file, err := os.Open(filename)
			if err != nil {
				return err
			}
			if err = compress(file, "", tw); err != nil {
				return err
			}
		}
		return nil
	})
}
func compress(file *os.File, prefix string, tw *tar.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func UnCompress(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return err
		}
		_, _ = io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}

func Filesize(size int64) string {
	sz := Div(size, 1024)
	if sz < 1024 {
		return fmt.Sprintf("%dKB", int64(sz))
	}
	exts := []string{"MB", "GB", "TB"}
	for i, ext := range exts {
		sz = Div(sz, 1024)
		if sz < 1024 || i == len(exts)-1 {
			return fmt.Sprintf("%.2f%s", Round(sz, 2), ext)
		}
	}
	return ""
}
