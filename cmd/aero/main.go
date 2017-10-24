package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aerogo/aero"

	"github.com/fatih/color"
)

var newApp bool

// Shell flags
func init() {
	flag.BoolVar(&newApp, "newapp", false, "Creates the basic structure of a new app in an empty directory")
	flag.Parse()
}

// Main
func main() {
	if newApp {
		color.Yellow("Creating new app...")

		fmt.Println("Creating", color.MagentaString(".gitignore"))
		gitignore()

		fmt.Println("Creating", color.MagentaString("config.json"))
		config()

		fmt.Println("Creating", color.MagentaString("main.go"))
		mainFile()

		color.Green("Finished.")
	}
}

func mainFile() {
	err := ioutil.WriteFile("main.go", []byte(mainCode), 0644)
	panicOnError(err)
}

func config() {
	config := aero.Configuration{}
	config.Reset()
	config.Styles = []string{}
	config.Fonts = []string{}
	bytes, err := json.MarshalIndent(config, "", "\t")
	panicOnError(err)

	err = ioutil.WriteFile("config.json", bytes, 0644)
	panicOnError(err)
}

func gitignore() {
	// Ignore current directory name as binary
	wd, err := os.Getwd()
	panicOnError(err)

	binaryName := "/" + filepath.Base(wd)
	err = ioutil.WriteFile(".gitignore", []byte(gitignoreText+"\n"+binaryName+"\n"), 0644)
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
