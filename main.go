package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	paths := os.Args[1:]
	before, err := resolveAbsPath(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	after, err := editWithEditor(before)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(before) != len(after) {
		fmt.Fprintf(os.Stderr, "error: the number of lines has changed\n")
		os.Exit(1)
	}

	for i, b := range before {
		if b != after[i] {
			a := after[i]
			err := moveFile(b, a)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func resolveAbsPath(paths []string) ([]string, error) {
	result := make([]string, len(paths))
	for i, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return result, err

		}
		result[i] = absPath
	}

	return result, nil
}

func editWithEditor(lines []string) (result []string, err error) {
	file, err := os.CreateTemp("", "rn.*.txt")
	if err != nil {
		return result, err
	}
	defer os.Remove(file.Name())

	file.WriteString(strings.Join(lines, "\n"))

	editor := os.Getenv("EDITOR")
	if editor == "" {
		return result, errors.New("EDITOR environment variable is not set")
	}

	c := exec.Command("sh", "-c", editor+" "+file.Name())
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err = c.Run()
	if err != nil {
		return result, err
	}

	if err := file.Close(); err != nil {
		return result, err
	}

	content, err := os.ReadFile(file.Name())
	if err != nil {
		return result, err
	}

	return strings.Split(strings.TrimRightFunc(string(content), unicode.IsSpace), "\n"), nil
}

func moveFile(from string, to string) error {
	dir := filepath.Dir(to)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.Rename(from, to)
}
