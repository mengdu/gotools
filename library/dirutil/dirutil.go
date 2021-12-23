package dirutil

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryanuber/go-glob"
)

type File struct {
	FilePath     string
	RelativePath string
	Entry        fs.DirEntry
}

type ChangeType string

type ChangeFile struct {
	File File
	Type ChangeType
}

const (
	CHANGE_CHANGE ChangeType = "change"
	CHANGE_REMOVE ChangeType = "remove"
	CHANGE_ADD    ChangeType = "add"
)

func ReadDir(dir string, ignore string, onlyFile bool) (map[string]File, error) {
	ignore = filepath.FromSlash(ignore)
	ignoresArr := strings.Split(ignore, ",")
	ignores := []string{}

	for _, v := range ignoresArr {
		if v != "" {
			ignores = append(ignores, v)
		}
	}

	m := map[string]File{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if path == dir {
			return nil
		}

		if onlyFile {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
		}

		relative, _ := filepath.Rel(dir, path)
		isFilter := false

		if len(ignores) > 0 {
			for _, v := range ignores {
				if glob.Glob(v, relative) {
					isFilter = true
					break
				}
			}
		}

		if !isFilter {
			m[relative] = File{
				FilePath:     path,
				RelativePath: relative,
				Entry:        d,
			}
		}
		return nil
	})

	return m, err
}

func Diff(new string, old string, ignore string, onlyFile bool) ([]ChangeFile, error) {
	arr := []ChangeFile{}
	afiles, err := ReadDir(new, ignore, onlyFile)
	adds := []ChangeFile{}
	changes := []ChangeFile{}
	removes := []ChangeFile{}

	if err != nil {
		return arr, err
	}

	bfiles, err := ReadDir(old, ignore, onlyFile)

	if err != nil {
		return arr, err
	}

	for relative, v := range afiles {
		if bv, ok := bfiles[relative]; ok {
			ainfo, err := v.Entry.Info()
			if err != nil {
				return arr, err
			}

			binfo, err := bv.Entry.Info()
			if err != nil {
				return arr, err
			}

			if !ainfo.IsDir() && !binfo.IsDir() {
				if ainfo.Size() != binfo.Size() {
					changes = append(changes, ChangeFile{
						File: v,
						Type: CHANGE_CHANGE,
					})
				} else {
					amd5, _ := Md5(v.FilePath)
					bmd5, _ := Md5(bv.FilePath)

					if amd5 != bmd5 {
						changes = append(changes, ChangeFile{
							File: v,
							Type: CHANGE_CHANGE,
						})
					}
				}
			} else if (!ainfo.IsDir() && binfo.IsDir()) || (ainfo.IsDir() && !binfo.IsDir()) {
				adds = append(adds, ChangeFile{
					File: v,
					Type: CHANGE_ADD,
				})
				removes = append(removes, ChangeFile{
					File: bv,
					Type: CHANGE_REMOVE,
				})
			}
		} else {
			adds = append(adds, ChangeFile{
				File: v,
				Type: CHANGE_ADD,
			})
		}
	}

	for relative, bv := range bfiles {
		if _, ok := afiles[relative]; !ok {
			removes = append(removes, ChangeFile{
				File: bv,
				Type: CHANGE_REMOVE,
			})
		}
	}

	arr = append(arr, adds...)
	arr = append(arr, changes...)
	arr = append(arr, removes...)

	return arr, nil
}

func Md5(file string) (string, error) {
	p, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer p.Close()

	if info, err := p.Stat(); err == nil {
		if info.IsDir() {
			return "", errors.New("Cannot be a folder.")
		}
	} else {
		return "", err
	}

	m := md5.New()
	io.Copy(m, p)
	return hex.EncodeToString(m.Sum(nil)), nil
}
