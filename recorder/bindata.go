package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"os"
	"time"
	"io/ioutil"
	"path"
	"path/filepath"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindata_file_info struct {
	name string
	size int64
	mode os.FileMode
	modTime time.Time
}

func (fi bindata_file_info) Name() string {
	return fi.name
}
func (fi bindata_file_info) Size() int64 {
	return fi.size
}
func (fi bindata_file_info) Mode() os.FileMode {
	return fi.mode
}
func (fi bindata_file_info) ModTime() time.Time {
	return fi.modTime
}
func (fi bindata_file_info) IsDir() bool {
	return false
}
func (fi bindata_file_info) Sys() interface{} {
	return nil
}

var _assets_index_html = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x7c\x52\xb1\x6e\xeb\x30\x0c\xdc\xdf\x57\xf0\x69\xb7\xb5\xbe\x41\xce\x5b\x8a\x02\x05\x3a\xb5\x53\x47\x59\x62\x62\x03\xb2\x64\x50\x74\xd3\xc0\xf0\xbf\x57\xb2\x9c\x38\xed\x90\xc9\xa2\x8e\x77\xe2\xf9\xa8\xfe\xda\x60\xf8\x32\x22\x74\x3c\xb8\xc3\x1f\x55\x3e\x00\xaa\x43\x6d\xf3\x21\x1d\x07\x64\x0d\xa6\xd3\x14\x91\x1b\x31\xf1\xb1\xfa\x27\x36\x88\x7b\x76\x78\x78\x43\x13\xc8\x22\x29\x59\xea\xcc\x97\x57\x01\xd5\x06\x7b\xb9\xb6\xeb\xd6\x21\xf4\xb6\x11\x51\x0f\xa3\xc3\xb8\xe9\x00\xcc\x33\x69\x7f\x42\xa8\xdf\x0b\xb0\x2c\x1b\xa0\x98\xc0\xe9\x16\x5d\x95\x69\xf3\x5c\xbf\x3c\x2d\xcb\x8d\x96\x71\x7b\x48\xec\x0a\xea\xd7\xdc\x05\xd5\xb2\xa4\xc7\xd9\xee\x1d\xf3\xdc\x1f\xa1\x7e\xee\x1d\xde\x44\x0b\xed\x56\xa4\x52\x4f\xb6\x0f\x60\x82\x67\x0a\x2e\xde\x43\x09\x8c\x61\x22\x83\x10\xc9\x34\x42\xd2\x6a\xb6\xf7\xa7\xfa\xac\x3f\xff\xa7\xa1\xae\x33\x41\xfe\x8f\x8d\x58\x95\xe4\x57\x95\x50\xf1\x53\xe7\x23\xc9\x40\x4b\xe1\x1c\x91\xc0\x06\x8c\xe0\x03\x43\x9c\xc6\x31\x10\x03\x77\x08\x65\x0a\x74\x38\xa0\xe7\xfa\x7e\x3e\xb9\x42\x77\xb6\x7f\x59\x44\x17\x1f\xd9\x6b\x27\xe6\xe0\xc1\x38\x1d\x63\x23\x8a\x85\xaa\x5c\x8a\x2d\x3e\x25\x4b\xfd\xe0\x0d\x6f\xf7\x58\x24\xd3\x9e\xdd\x8e\xa4\xfb\x9c\x71\x59\x81\x92\x7c\x5a\x85\x75\xa9\xbe\x03\x00\x00\xff\xff\x38\x17\x80\x0e\x6c\x02\x00\x00")

func assets_index_html_bytes() ([]byte, error) {
	return bindata_read(
		_assets_index_html,
		"assets/index.html",
	)
}

func assets_index_html() (*asset, error) {
	bytes, err := assets_index_html_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "assets/index.html", size: 620, mode: os.FileMode(420), modTime: time.Unix(1467991314, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if (err != nil) {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/index.html": assets_index_html,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() (*asset, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"assets": &_bintree_t{nil, map[string]*_bintree_t{
		"index.html": &_bintree_t{assets_index_html, map[string]*_bintree_t{
		}},
	}},
}}

// Restore an asset under the given directory
func RestoreAsset(dir, name string) error {
        data, err := Asset(name)
        if err != nil {
                return err
        }
        info, err := AssetInfo(name)
        if err != nil {
                return err
        }
        err = os.MkdirAll(_filePath(dir, path.Dir(name)), os.FileMode(0755))
        if err != nil {
                return err
        }
        err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
        if err != nil {
                return err
        }
        err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
        if err != nil {
                return err
        }
        return nil
}

// Restore assets under the given directory recursively
func RestoreAssets(dir, name string) error {
        children, err := AssetDir(name)
        if err != nil { // File
                return RestoreAsset(dir, name)
        } else { // Dir
                for _, child := range children {
                        err = RestoreAssets(dir, path.Join(name, child))
                        if err != nil {
                                return err
                        }
                }
        }
        return nil
}

func _filePath(dir, name string) string {
        cannonicalName := strings.Replace(name, "\\", "/", -1)
        return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

