package cmdutil

import (
	"bufio"
	"fmt"
	"os"
)

// Pause pauses the output so it can be visualized before closing.
func Pause() {
	fmt.Fprintf(os.Stdout, "Press Enter to continue...\n")
	bufio.NewScanner(os.Stdin).Scan()
}
