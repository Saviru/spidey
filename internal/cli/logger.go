package cli

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
)

func Ready(msg string) {
	fmt.Printf("%s%s✓ [READY]%s %s\n", Bold, Green, Reset, msg)
}
func Info(scope, msg string) {
	fmt.Printf("%s%sℹ [%s]%s %s\n", Bold, Cyan, scope, Reset, msg)
}
func Event(event, detail string) {
	fmt.Printf("%s%s⚡ [%s]%s %s\n", Bold, Magenta, event, Reset, detail)
}
func Warn(scope, msg string) {
	fmt.Printf("%s%s⚠ [WARN][%s]%s %s\n", Bold, Yellow, scope, Reset, msg)
}
func Error(scope, msg string) {
	fmt.Fprintf(os.Stderr, "%s%s✖ [ERROR][%s]%s %s\n", Bold, Red, scope, Reset, msg)
}

func Success(scope, msg string) {
	fmt.Printf("%s%s✓ [%s]%s %s\n", Bold, Green, strings.ToUpper(scope), Reset, msg)
}

func DevBanner(version, port string, duration time.Duration) {
	ms := duration.Milliseconds()
	fmt.Printf("\n  %s%s✓%s %s%sspidey %s%s %sready in %dms%s\n\n",
		Bold, Green, Reset,
		Bold, "", version, Reset,
		Dim, ms, Reset,
	)
	fmt.Printf("  %s%s➜%s  %sLocal:%s   %shttp://localhost:%s%s\n",
		Bold, Cyan, Reset,
		Bold, Reset,
		Cyan, port, Reset,
	)

	/*TODO if netIP != "" {
		fmt.Printf("  %s%s➜%s  %sNetwork:%s %shttp://%s:%s%s\n",
			Bold, Cyan, Reset,
			Bold, Reset,
			Cyan, netIP, port, Reset,
		)
	}*/

	fmt.Println()
}
