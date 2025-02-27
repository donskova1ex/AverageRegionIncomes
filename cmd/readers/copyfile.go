package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func testCopying() {
	mainDir := "./files"
	containerName := "reader.reader"
	containerDir := "/db-files/"

	checkCmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")
	out, err := checkCmd.Output()
	if err != nil {
		fmt.Printf("failed to check container existence")
	}
	if len(out) == 0 || string(out) != containerName+"\n" {
		fmt.Printf("container %s does not exist", containerName)
	}

	files, err := filepath.Glob(filepath.Join(mainDir, "*"))
	if err != nil {
		fmt.Printf("failed to list files in directory")
	}
	if len(files) == 0 {
		fmt.Printf("directory %s is empty", mainDir)
	}

	for _, file := range files {
		copyCmd := exec.Command("docker", "cp", file, fmt.Sprintf("%s:%s", containerName, containerDir))
		output, err := copyCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("failed to copy %s: %s.", file, string(output))
		}
	}

}
