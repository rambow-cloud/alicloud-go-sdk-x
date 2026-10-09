package codegen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DriftError lists missing, changed or stale owned files, relative to the repository.
type DriftError struct{ Files []string }

func (e *DriftError) Error() string {
	return "codegen: generated files differ; run sdkgen generate:\n" + strings.Join(e.Files, "\n")
}

func reconcileOwned(ctx context.Context, root string, files map[string][]byte, check bool, bases []string, owns func([]byte) bool) error {
	var drift []string
	for _, name := range sortedKeys(files) {
		if err := safeOutputPath(root, name); err != nil {
			return err
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		existing, err := os.ReadFile(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err == nil && !owns(existing) {
			return fmt.Errorf("codegen: refusing to overwrite unmarked file %s", name)
		}
		if !bytes.Equal(existing, files[name]) {
			drift = append(drift, name)
		}
	}
	var stale []string
	for _, base := range bases {
		if err := safeOutputPath(root, base+"/.sdkgen-scan"); err != nil {
			return err
		}
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(base)), func(path string, entry fs.DirEntry, walkErr error) error {
			if errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&fs.ModeSymlink != 0 {
				return fmt.Errorf("codegen: symlink in output tree: %s", entry.Name())
			}
			if entry.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".json") && entry.Name() != "LICENSE" && entry.Name() != "NOTICE" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !owns(data) {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			name := filepath.ToSlash(relative)
			if _, expected := files[name]; !expected {
				stale = append(stale, name)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	drift = append(drift, stale...)
	sort.Strings(drift)
	sort.Strings(stale)
	if err := ctx.Err(); err != nil {
		return err
	}
	if check {
		if len(drift) > 0 {
			return &DriftError{Files: drift}
		}
		return nil
	}
	for _, name := range drift {
		data, expected := files[name]
		if !expected {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := atomicWrite(path, data); err != nil {
			return err
		}
	}
	for _, name := range stale {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			return err
		}
	}
	return nil
}

func safeOutputPath(root, name string) error {
	if filepath.IsAbs(name) || strings.Contains(name, "\\") {
		return errors.New("codegen: invalid output path")
	}
	path := filepath.Join(root, filepath.FromSlash(name))
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("codegen: output path escapes root")
	}
	current := root
	if info, err := os.Lstat(current); err != nil {
		return err
	} else if info.Mode()&fs.ModeSymlink != 0 {
		return errors.New("codegen: output root is a symlink")
	}
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return errors.New("codegen: output path contains a symlink")
		}
	}
	return nil
}

func atomicWrite(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".sdkgen-*")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Chmod(0644); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
