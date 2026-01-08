package ui

import (
	"fmt"
)

const (
	Check   = "✔"
	Arrow   = "→"
	Cross   = "✗"
    
    // ANSI Colors
    Reset  = "\033[0m"
    Red    = "\033[31m"
    Green  = "\033[32m"
    Yellow = "\033[33m"
    Blue   = "\033[34m"
    Cyan   = "\033[36m"
    Bold   = "\033[1m"
)

func PrintHeader() {
	fmt.Println()
	fmt.Printf("  %s================= MyTM Packager =================%s\n", Bold, Reset)
	fmt.Println()
}

func Section(title string) {
    fmt.Println()
	fmt.Printf("%s %s%s\n", Arrow, title, "...")
}

func SubItem(msg string) {
	fmt.Printf("    %s %s\n", Arrow, msg)
}

func SubItem2(msg string) {
	fmt.Printf("        %s %s\n", Arrow, msg)
}

func Success(msg string) {
	fmt.Printf("%s%s %s%s\n", Green, Check, msg, Reset)
}

func SubSuccess(msg string) {
	fmt.Printf("    %s%s %s%s\n", Green, Check, msg, Reset)
}

func SubSuccess2(msg string) {
	fmt.Printf("        %s%s %s%s\n", Green, Check, msg, Reset)
}

func Error(msg string) {
    fmt.Printf("%s%s %s%s\n", Red, Cross, msg, Reset)
}
