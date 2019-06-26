package tools

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/henrylee2cn/aster/internal/imports"
	"github.com/henrylee2cn/goutil"
)

// Options specifies options for processing files.
//
// type Options struct {
// 	Fragment  bool // Accept fragment of a source file (no package statement)
// 	AllErrors bool // Report all errors (not just the first 10 on different lines)
// 	Comments  bool // Print comments (true if nil *Options provided)
// 	TabIndent bool // Use tabs for indent (true if nil *Options provided)
// 	TabWidth  int  // Tab width (8 if nil *Options provided)
// 	FormatOnly bool // Disable the insertion and deletion of imports
// }
type Options = imports.Options

// Format formats and adjusts imports for the provided file.
// If opt is nil the defaults are used.
//
// Note that filename's directory influences which imports can be chosen,
// so it is important that filename be accurate.
// To process data ``as if'' it were in filename, pass the data as a non-nil src.
func Format(filename string, src []byte, opt *Options) ([]byte, error) {
	return imports.Process("", src, nil)
}

var pkglineRegexp = regexp.MustCompile("\n*package[\t ]+([^/\n]+)[/\n]")

// ChangePkgName change package name of the code and return the new code.
func ChangePkgName(code string, pkgname string) string {
	s := strings.TrimSpace(pkglineRegexp.FindString(code))
	s = strings.TrimSpace(strings.TrimRight(s, "/"))
	if s == "" {
		return code
	}
	return strings.Replace(code, s, "package "+pkgname, 1)
}

// PkgName get the package name of the code, file or directory.
// NOTE:
//  If src==nil, find the package name from the file or directory specified by 'filenameOrDirectory';
//  If src!=nil, find the package name from the code represented by 'src'.
func PkgName(filenameOrDirectory string, src interface{}) (string, error) {
	if src == nil {
		existed, isDir := goutil.FileExist(filenameOrDirectory)
		if !existed {
			return "", errors.New("file or directory is not existed")
		}
		if isDir {
			err := filepath.Walk(filenameOrDirectory, func(path string, f os.FileInfo, err error) error {
				if err != nil || f.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".go") {
					filenameOrDirectory = path
					return errors.New("")
				}
				return nil
			})
			if err == nil || err.Error() != "" {
				return "", err
			}
		}
	}
	b, err := readSource(filenameOrDirectory, src)
	if err != nil {
		return "", err
	}
	r := pkglineRegexp.FindSubmatch(b)
	if len(r) < 2 {
		return "", nil
	}
	return goutil.BytesToString(bytes.TrimSpace(r[1])), nil
}

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return ioutil.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}
