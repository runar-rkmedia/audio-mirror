package cache

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
)

type cache struct {
	baseDir string
}

func NewCache(baseDir string) *cache {
	fmt.Println("fofo")
	return &cache{baseDir}
}

func (c *cache) Retrieve(keyPaths []string, changedAfter time.Time) ([]byte, bool, error) {
	filePath, err := c.CreateFilePath(keyPaths)
	if err != nil {
		return nil, false, err
	}
	f, err := os.Open(filePath)
	fmt.Println("fif", filePath, err)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer f.Close()
	stat, err := f.Stat()
	fmt.Println("fif", filePath, stat, err)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if stat.ModTime().Before(changedAfter) {
		return nil, false, nil
	}
	b, err := io.ReadAll(f)
	return b, true, err
}

func (c *cache) Write(keyPaths []string, value []byte) (string, error) {
	filePath, err := c.CreateFilePath(keyPaths)
	if err != nil {
		return filePath, err
	}

	dir := filepath.Dir(filePath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return filePath, err
	}

	err = os.WriteFile(filePath, value, 0644)
	if err != nil {
		return filePath, fmt.Errorf("failed to write cache for file %s: %w", filePath, err)
	}

	return filePath, nil
}

func (c *cache) CreateFilePath(baseNames []string) (string, error) {
	filePath := c.baseDir

	lenght := len(baseNames)
	for i, v := range baseNames {
		fileName := sanitize.BaseName(strings.TrimSuffix(v, filepath.Ext(v)))
		if i == lenght-1 {
			ext := sanitize.BaseName(strings.TrimPrefix(filepath.Ext(v), "."))
			if ext != "" {
				fileName += "." + ext
			}
		}
		filePath = path.Join(filePath, "/"+fileName)
	}

	return filePath, nil
}
