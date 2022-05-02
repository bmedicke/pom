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

//go:embed callbacks/default.sh
var defaultCallbackContent string

const callbackfolder string = ".pom/callbacks"

func main() {
	createConfig := flag.Bool(
		"create-config",
		false,
		"create .pom config folder with example callbacks")

	flag.Parse()

	if *createConfig {
		createConfigFilesAndFolders()
	} else {
		spawnTUI()
	}
}

func createConfigFilesAndFolders() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	callbackpath := filepath.Join(home, callbackfolder)
	err = os.MkdirAll(callbackpath, 0700)
	if err != nil {
		log.Panic(err)
	}

	defaultCallbacks := []string{
		"work_start.sh",
		"work_done.sh",
		"break_start.sh",
		"break_done.sh",
	}

	for _, callback := range defaultCallbacks {
		file := filepath.Join(callbackpath, callback)
		_, err := os.Stat(file)
		if errors.Is(err, os.ErrNotExist) {
			f, _ := os.Create(file)
			f.WriteString(defaultCallbackContent)
			os.Chmod(file, 0700) // make it executable.
			defer f.Close()
		}
	}

	fmt.Println("the config folder can be found at: ~/.pom/")
}
