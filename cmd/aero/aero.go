package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

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

		fmt.Println("Creating config.json")

		config := aero.Configuration{}
		config.Reset()
		config.Styles = []string{}
		config.Fonts = []string{}
		bytes, err := json.MarshalIndent(config, "", "\t")

		if err != nil {
			panic(err)
		}

		ioutil.WriteFile("config.json", bytes, 0644)

		color.Green("Finished.")
	}
}
