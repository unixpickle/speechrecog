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

var _assets_index_html = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x7c\x52\xb1\xb2\xdb\x20\x10\xec\xdf\x57\x5c\xe8\x25\xda\x14\xc8\x69\x32\x99\xc9\x4c\xaa\xa4\x4a\x89\xe0\xfc\x44\x82\x41\x73\x9c\xf2\x9e\x47\xa3\x7f\x0f\x08\x6c\x39\x29\x5c\x49\xcb\x72\x7b\xbb\xdc\xa9\x0f\x36\x1a\xbe\xce\x08\x13\x5f\xfc\xe9\x45\xd5\x0f\x80\x9a\x50\xdb\xf2\x93\x7f\x2f\xc8\x1a\xcc\xa4\x29\x21\x0f\x62\xe1\x73\xf7\x51\x34\x8a\x1d\x7b\x3c\x7d\x47\x13\xc9\x22\x29\x59\xf1\x4b\x25\xbd\x0b\xbf\x81\xd0\x0f\x22\xf1\xd5\x63\x9a\x10\x59\x40\xe9\x36\x08\xc6\x77\x96\x26\x25\x01\x13\xe1\x79\x10\x72\xbf\xd2\x97\x93\x26\x9d\x0c\xb9\x99\x21\x91\x29\xec\x0e\xfa\x5f\x99\x55\x0d\xec\x2e\xe5\xcd\xa6\x1a\xa3\xbd\xde\x4c\xe9\xd1\x23\x38\x9b\xfb\xea\xcb\x9c\x1b\x37\x49\x80\x75\x25\x1d\x5e\x11\xfa\x1f\x95\xd8\xb6\x46\x28\x26\xf0\x7a\x44\xdf\x95\xb2\x75\xed\xbf\x7e\xde\xb6\x7b\x59\xe1\xed\x29\x57\x77\xd0\x7f\x2b\xb7\xa0\xdb\xb6\xdc\x9c\xed\x71\x63\x5d\xdd\x19\xfa\x2f\xce\xe3\x5d\xb4\x96\xdd\x41\x86\x7a\xb1\x2e\x82\x89\x81\x29\xfa\xf4\x48\x95\xc0\x71\x21\x83\x2d\x30\xed\x4f\xea\xc2\x6b\xff\xa6\xff\x7c\xca\xa6\x6e\x9e\xda\xfb\xed\x4a\xf2\xbd\xcb\xac\xf8\x57\xe7\x67\x96\x81\x91\xe2\x5b\x42\x02\x1b\x31\x41\x88\xf9\x19\x97\x79\x8e\xc4\xc0\x13\x42\x75\x81\x1e\x2f\x18\xb8\x7f\xf4\x27\x77\xea\x21\xf6\x7f\x11\xd1\xa7\x67\xf1\xc6\x85\x39\x06\x30\x5e\xa7\x34\x88\x1a\xa1\xab\x87\xa2\x2d\x89\x92\x15\x3f\xe9\x11\xec\x31\x16\xc9\x74\xcc\xee\x60\xf2\x79\x99\x71\x5d\x81\x3a\xf9\xbc\x0a\xfb\xea\xfe\x0d\x00\x00\xff\xff\xc9\x8a\xe7\x7a\xd2\x02\x00\x00")

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

	info := bindata_file_info{name: "assets/index.html", size: 722, mode: os.FileMode(420), modTime: time.Unix(1467994771, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _assets_script_js = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func assets_script_js_bytes() ([]byte, error) {
	return bindata_read(
		_assets_script_js,
		"assets/script.js",
	)
}

func assets_script_js() (*asset, error) {
	bytes, err := assets_script_js_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "assets/script.js", size: 0, mode: os.FileMode(420), modTime: time.Unix(1467994553, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _assets_style_css = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x52\x2e\x4e\xcc\x2d\xc8\x49\x2d\x56\xa8\xe6\x52\x50\xc8\x4d\x2c\x4a\xcf\xcc\xb3\x52\x48\x2c\x2d\xc9\xb7\xe6\xaa\xe5\x02\x04\x00\x00\xff\xff\xd9\xe1\x58\x03\x1d\x00\x00\x00")

func assets_style_css_bytes() ([]byte, error) {
	return bindata_read(
		_assets_style_css,
		"assets/style.css",
	)
}

func assets_style_css() (*asset, error) {
	bytes, err := assets_style_css_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "assets/style.css", size: 29, mode: os.FileMode(420), modTime: time.Unix(1467994924, 0)}
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
	"assets/script.js": assets_script_js,
	"assets/style.css": assets_style_css,
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
		"script.js": &_bintree_t{assets_script_js, map[string]*_bintree_t{
		}},
		"style.css": &_bintree_t{assets_style_css, map[string]*_bintree_t{
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

