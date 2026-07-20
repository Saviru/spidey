package bundler

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/saviru/spidey/internal/config"
)

func CompileBinary(projectDir string, cfg *config.Config) error {
	outPath := cfg.Directories.OutputDir
	if runtime.GOOS == "windows" {
		outPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", outPath, "./api")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
