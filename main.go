package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

//go:embed hooks/default.sh
var defaultHookContent string

//go:embed config.json
var defaultConfigContent string

const configfolder = ".pom"

var hookfolder = "hooks/"

func main() {
	createConfig := flag.Bool(
		"create-config",
		false,
		"create .pom config folder with example hooks")

	hookProfile := flag.String(
		"profile",
		"default",
		"select hook profile from ~/.pom/hooks/",
	)

	flag.Parse()

	hookfolder = filepath.Join(hookfolder, *hookProfile)

	// TODO: read configJSON from file.
	// config := map[string]interface{}{}
	// json.Unmarshal([]byte(configJSON), &config)

	if *createConfig {
		createConfigFilesAndFolders()
	} else {
		spawnTUI()
		clearTmuxFile()
	}
}

func createConfigFilesAndFolders() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	// create config folders:
	hookpath := filepath.Join(home, configfolder, hookfolder)
	err = os.MkdirAll(hookpath, 0700)
	if err != nil {
		log.Panic(err)
	}

	// create configfile
	configfile := filepath.Join(home, configfolder, "config.json")
	_, err = os.Stat(configfile)
	if errors.Is(err, os.ErrNotExist) {
		f, _ := os.Create(configfile)
		f.WriteString(defaultConfigContent)
		defer f.Close()
	}

	defaultHooks := []string{
		"work_start",
		"work_done",
		"break_start",
		"break_done",
	}

	// create default hook scripts:
	for _, hook := range defaultHooks {
		file := filepath.Join(hookpath, hook)
		_, err := os.Stat(file)
		if errors.Is(err, os.ErrNotExist) {
			f, _ := os.Create(file)
			f.WriteString(defaultHookContent)
			os.Chmod(file, 0700) // make it executable.
			defer f.Close()
		}
	}

	fmt.Println("the config folder can be found at: ~/.pom/")
}
