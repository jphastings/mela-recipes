package prompts

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed *.prompt.txt
var promptFS embed.FS

var System string
var Ignore string
var Map map[string]string

func init() {
	Map = make(map[string]string)
	err := fs.WalkDir(promptFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		fileContent, err := promptFS.ReadFile(path)
		if err != nil {
			return err
		}

		key := strings.TrimSuffix(path, ".prompt.txt")

		switch key {
		case "_system":
			System = string(fileContent)
		case "_ignore":
			Ignore = string(fileContent)
		default:
			Map[key] = string(fileContent)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("Unable to load ePub conversion prompts: %v", err))
	}
}
