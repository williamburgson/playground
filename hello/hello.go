package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"example.com/greeting"
)

// defines a custom flag type that can hold an array of strings
type arrFlag []string

func (i *arrFlag) String() string {
	return ""
}

func (i *arrFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	usrNames    arrFlag
)

func main() {
	// builtin lib for cmd flags
	// (name of the flag, default value, help string)
	flag.Var(&usrNames, "names", "Name of the User")
	flag.Parse()
	// the parsed flag is stored as pointers
	msgs, err := greeting.Hellos(usrNames)
	if err != nil {
		errorLogger.Fatal(err)
	}
	for _, msg := range msgs {
		infoLogger.Println(msg)
	}
}

func buildLogger(level string) *log.Logger {
	prefix := fmt.Sprintf("%v: ", level)
	return log.New(os.Stderr, prefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func init() {
	infoLogger = buildLogger("INFO")
	errorLogger = buildLogger("ERROR")
}
