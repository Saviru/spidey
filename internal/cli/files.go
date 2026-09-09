package cli

import (
	"fmt"
	"runtime"
)

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func DetectOS() string {
	switch runtime.GOOS {
	case "windows":
		return ".exe"
	case "darwin":
		return ".app"
	default:
		return ".bin"
	}
}
