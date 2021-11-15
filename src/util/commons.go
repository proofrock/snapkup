package util

import (
	"fmt"
	"os"
)

const Mega = 1024 * 1024

const ChunkSize = 32 * Mega // 32Mb

func PrintlnOut(vals ...interface{}) {
	vals = append(vals, "\n")
	fmt.Fprintln(os.Stdout, vals...)
}

func PrintlnErr(vals ...interface{}) {
	vals = append(vals, "\n")
	fmt.Fprintln(os.Stderr, vals...)
}

func PrintlnfOut(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", vals...)
}

func PrintlnfErr(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", vals...)
}

func PrintOut(vals ...interface{}) {
	fmt.Fprintln(os.Stdout, vals...)
}

func PrintErr(vals ...interface{}) {
	fmt.Fprintln(os.Stderr, vals...)
}

func PrintfOut(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stdout, format, vals...)
}

func PrintfErr(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, format, vals...)
}
