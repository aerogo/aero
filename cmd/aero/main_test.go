package main

import (
	"os"
	"path"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNoParameters(t *testing.T) {
	c := qt.New(t)

	oldPath, err := os.Getwd()
	c.Assert(err, qt.IsNil)

	projectPath := path.Join(os.TempDir(), "aero-app-test")
	err = os.Mkdir(projectPath, 0777)
	c.Assert(err, qt.IsNil)

	defer func() {
		err := os.RemoveAll(projectPath)
		c.Assert(err, qt.IsNil)

		err = os.Chdir(oldPath)
		c.Assert(err, qt.IsNil)
	}()

	err = os.Chdir(projectPath)
	c.Assert(err, qt.IsNil)
	main()
}

func TestNewApp(t *testing.T) {
	c := qt.New(t)

	oldPath, err := os.Getwd()
	c.Assert(err, qt.IsNil)

	projectPath := path.Join(os.TempDir(), "aero-app-test")
	err = os.Mkdir(projectPath, 0777)
	c.Assert(err, qt.IsNil)

	defer func() {
		err := os.RemoveAll(projectPath)
		c.Assert(err, qt.IsNil)

		err = os.Chdir(oldPath)
		c.Assert(err, qt.IsNil)
	}()

	err = os.Chdir(projectPath)
	c.Assert(err, qt.IsNil)
	os.Args = append(os.Args, "-newapp")
	main()
}
