package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aerogo/aero"
	"github.com/akyoto/color"
)

var newApp bool

// Shell flags
func init() {
	flag.BoolVar(&newApp, "new", false, "Creates the basic structure of a new app in an empty directory")
}

// Main
func main() {
	flag.Parse()

	if !newApp {
		flag.Usage()
		return
	}

	color.Yellow("Creating new app...\n\n")
	icon := color.GreenString(" ✔ ")

	fmt.Println(icon, ".gitignore")
	gitignore()

	fmt.Println(icon, "config.json")
	config()

	fmt.Println(icon, "tsconfig.json")
	tsconfig()

	fmt.Println(icon, "main.go")
	mainFile()

	fmt.Println(icon, "main_test.go")
	mainTestFile()

	fmt.Println(icon, "go.mod")
	makeModule()

	color.Green("\nFinished.")
}

func mainFile() {
	err := ioutil.WriteFile("main.go", []byte(mainCode), 0644)
	check(err)
}

func mainTestFile() {
	err := ioutil.WriteFile("main_test.go", []byte(mainTestCode), 0644)
	check(err)
}

func config() {
	config := aero.Configuration{}
	config.Reset()

	bytes, err := json.MarshalIndent(config, "", "\t")
	check(err)

	err = ioutil.WriteFile("config.json", bytes, 0644)
	check(err)
}

func tsconfig() {
	err := ioutil.WriteFile("tsconfig.json", []byte(tsconfigText), 0644)
	check(err)
}

func gitignore() {
	// Ignore current directory name as binary
	wd, err := os.Getwd()
	check(err)

	binaryName := "/" + filepath.Base(wd)
	err = ioutil.WriteFile(".gitignore", []byte(gitignoreText+"\n"+binaryName+"\n"), 0644)
	check(err)
}

func makeModule() {
	wd, err := os.Getwd()
	check(err)

	cmd := exec.Command("go", "mod", "init", "example.com/"+filepath.Base(wd))
	err = cmd.Run()
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
