package module

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/afero"
)

const (
	gitIgnoreFilename = ".gitignore"
)

// MakeZip takes dir and module info and generates vgo valid zip
// the dir must end with a "/"
func MakeZip(fs afero.Fs, dir, module, version string) *io.PipeReader {
	ignoreParser := getIgnoreParser(fs, dir)
	pr, pw := io.Pipe()

	go func() {
		zw := zip.NewWriter(pw)
		defer zw.Close()

		walkFn := walkFunc(fs, zw, dir, module, version, ignoreParser)
		err := afero.Walk(fs, dir, walkFn)
		pw.CloseWithError(err)
	}()
	return pr
}

func walkFunc(fs afero.Fs, zw *zip.Writer, dir, module, version string, ignoreParser ignore.IgnoreParser) filepath.WalkFunc {
	const op errors.Op = "module.walkFunc"
	return func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return errors.E(op, err)
		}

		fileName := getFileName(path, dir, module, version)

		if ignoreParser.MatchesPath(fileName) {
			return nil
		}

		fileContent, err := afero.ReadFile(fs, path)
		if err != nil {
			return errors.E(op, err)
		}

		f, err := zw.Create(fileName)
		if err != nil {
			return errors.E(op, err)
		}

		_, err = f.Write(fileContent)
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}
}

func getIgnoreParser(fs afero.Fs, dir string) ignore.IgnoreParser {
	gitFilePath := filepath.Join(dir, gitIgnoreFilename)
	gitParser := compileIgnoreFileAndLines(fs, gitFilePath, gitIgnoreFilename)
	dsStoreParser := dsStoreIgnoreParser{}

	return newMultiIgnoreParser(gitParser, dsStoreParser)
}

func compileIgnoreFileAndLines(fs afero.Fs, fpath string, lines ...string) ignore.IgnoreParser {
	buffer, err := afero.ReadFile(fs, fpath)
	if err != nil {
		return nil
	}
	s := strings.Split(string(buffer), "\n")
	ip, err := ignore.CompileIgnoreLines(append(s, lines...)...)
	if err != nil {
		// if we return ip, then it won't be a nil interface,
		// even if ip is a nil pointer.
		return nil
	}

	return ip
}

// getFileName composes filename for zip to match standard specified as
// module@version/{filename}
func getFileName(path, dir, module, version string) string {
	filename := strings.TrimPrefix(path, dir)
	filename = strings.TrimLeftFunc(filename, func(r rune) bool { return r == os.PathSeparator })

	moduleID := fmt.Sprintf("%s@%s", module, version)

	return filepath.Join(moduleID, filename)
}
