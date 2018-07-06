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
}

// Main
func main() {
	flag.Parse()

	if !newApp {
		flag.Usage()
		return
	}

	faint := color.New(color.Faint).SprintFunc()

	color.Yellow("Creating new app...")
	println()

	fmt.Println(color.GreenString(" ✔ "), ".gitignore")
	gitignore()

	fmt.Println(color.GreenString(" ✔ "), "config.json")
	config()

	fmt.Println(color.GreenString(" ✔ "), "tsconfig.json")
	tsconfig()

	fmt.Println(color.GreenString(" ✔ "), "main.go")
	mainFile()

	fmt.Println(color.GreenString(" ✔ "), "main_test.go")
	mainTestFile()

	fmt.Println(color.GreenString(" ✔ "), faint("layout"))
	createDirectory("layout")

	fmt.Println(color.GreenString(" ✔ "), faint("pages"))
	createDirectory("pages")

	fmt.Println(color.GreenString(" ✔ "), faint("scripts"))
	createDirectory("scripts")
	panicOnError(ioutil.WriteFile("scripts/main.ts", []byte(`console.log("Hello World")`), 0644))

	fmt.Println(color.GreenString(" ✔ "), faint("security"))
	createDirectory("security")
	gitignoreAll("security")

	fmt.Println(color.GreenString(" ✔ "), faint("styles"))
	createDirectory("styles")

	println()
	color.Green("Finished.")
}

func createDirectory(name string) {
	err := os.Mkdir(name, 0777)

	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func mainFile() {
	err := ioutil.WriteFile("main.go", []byte(mainCode), 0644)
	panicOnError(err)
}

func mainTestFile() {
	err := ioutil.WriteFile("main_test.go", []byte(mainTestCode), 0644)
	panicOnError(err)
}

func config() {
	config := aero.Configuration{}
	config.Reset()
	config.Styles = []string{}
	config.Fonts = []string{}
	config.Push = []string{}
	config.Scripts.Main = "main"
	bytes, err := json.MarshalIndent(config, "", "\t")
	panicOnError(err)

	err = ioutil.WriteFile("config.json", bytes, 0644)
	panicOnError(err)
}

func tsconfig() {
	err := ioutil.WriteFile("tsconfig.json", []byte(tsconfigText), 0644)
	panicOnError(err)
}

func gitignoreAll(directory string) {
	err := ioutil.WriteFile(directory+"/.gitignore", []byte("*"), 0644)
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
