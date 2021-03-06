package print

import (
	"fmt"
	"os"
)

var (
	debugMode = false
)

// Init ...
func Init(debug bool) {
	debugMode = debug
}

// Debug method output debug message is run as debug mode
func Debug(format string, a ...interface{}) {
	if debugMode {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf("%s\n", msg)
	}
}

// Print ...
func Print(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s\n", msg)
}

// Error method output message to STDERR
// this method use in an error caused by user
func Error(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

// Fatal ...
func Fatal(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
	os.Exit(1)
}
