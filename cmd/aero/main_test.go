package main

import (
	"os"
	"path"
	"testing"
)

func TestNoParameters(t *testing.T) {
	oldPath, _ := os.Getwd()
	defer os.Chdir(oldPath)

	projectPath := path.Join(os.TempDir(), "aero-app-test")

	os.Mkdir(projectPath, 0777)
	defer os.RemoveAll(projectPath)

	os.Chdir(projectPath)

	main()
}

func TestNewApp(t *testing.T) {
	oldPath, _ := os.Getwd()
	defer os.Chdir(oldPath)

	projectPath := path.Join(os.TempDir(), "aero-app-test")

	os.Mkdir(projectPath, 0777)
	defer os.RemoveAll(projectPath)

	os.Chdir(projectPath)

	os.Args = append(os.Args, "-newapp")
	main()
}
