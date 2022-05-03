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

const hookfolder string = ".pom/hooks"

func main() {
	createConfig := flag.Bool(
		"create-config",
		false,
		"create .pom config folder with example hooks")

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

	hookpath := filepath.Join(home, hookfolder)
	err = os.MkdirAll(hookpath, 0700)
	if err != nil {
		log.Panic(err)
	}

	defaultHooks := []string{
		"work_start.sh",
		"work_done.sh",
		"break_start.sh",
		"break_done.sh",
	}

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
