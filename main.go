package main

import (
	"fmt"
	"log"

	"github.com/krasnikov138/register/cmd"
)

type logWriter struct {
}

// here you can specify your favorite logging pattern
func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	cmd.Execute()
}
