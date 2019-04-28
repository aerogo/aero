package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestNoParameters(t *testing.T) {
	oldPath, err := os.Getwd()
	assert.NoError(t, err)

	projectPath := path.Join(os.TempDir(), "aero-app-test")
	err = os.Mkdir(projectPath, 0777)
	assert.NoError(t, err)

	defer func() {
		err := os.RemoveAll(projectPath)
		assert.NoError(t, err)

		err = os.Chdir(oldPath)
		assert.NoError(t, err)
	}()

	err = os.Chdir(projectPath)
	assert.NoError(t, err)
	main()
}

func TestNewApp(t *testing.T) {
	oldPath, err := os.Getwd()
	assert.NoError(t, err)

	projectPath := path.Join(os.TempDir(), "aero-app-test")
	err = os.Mkdir(projectPath, 0777)
	assert.NoError(t, err)

	defer func() {
		err := os.RemoveAll(projectPath)
		assert.NoError(t, err)

		err = os.Chdir(oldPath)
		assert.NoError(t, err)
	}()

	err = os.Chdir(projectPath)
	assert.NoError(t, err)
	os.Args = append(os.Args, "-newapp")
	main()
}
