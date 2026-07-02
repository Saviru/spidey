package bundler

import (
	"os"
	"os/exec"
)

func CompileBinary(projectDir string) error {
	cmd := exec.Command("go", "build", "-o", "bin/server.exe", "./api/main.go")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
