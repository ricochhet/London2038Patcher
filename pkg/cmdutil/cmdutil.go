package cmdutil

import (
	"bufio"
	"os"

	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

// Pause pauses the output so it can be visualized before closing.
func Pause() {
	logutil.Infof(logutil.Get(), "Press Enter to continue...\n")
	bufio.NewScanner(os.Stdin).Scan()
}
