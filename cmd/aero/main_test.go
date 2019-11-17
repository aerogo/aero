package main

import (
	"os"
	"path"
	"testing"

	"github.com/akyoto/assert"
)

func TestNoParameters(t *testing.T) {
	oldPath, err := os.Getwd()
	assert.Nil(t, err)

	projectPath := path.Join(os.TempDir(), "aero-app-test")
	err = os.Mkdir(projectPath, 0777)
	assert.Nil(t, err)

	defer func() {
		err := os.RemoveAll(projectPath)
		assert.Nil(t, err)

		err = os.Chdir(oldPath)
		assert.Nil(t, err)
	}()

	err = os.Chdir(projectPath)
	assert.Nil(t, err)
	main()
}

func TestNewApp(t *testing.T) {
	oldPath, err := os.Getwd()
	assert.Nil(t, err)

	projectPath := path.Join(os.TempDir(), "aero-app-test")
	err = os.Mkdir(projectPath, 0777)
	assert.Nil(t, err)

	defer func() {
		err := os.RemoveAll(projectPath)
		assert.Nil(t, err)

		err = os.Chdir(oldPath)
		assert.Nil(t, err)
	}()

	err = os.Chdir(projectPath)
	assert.Nil(t, err)
	os.Args = append(os.Args, "-new")
	main()
}
