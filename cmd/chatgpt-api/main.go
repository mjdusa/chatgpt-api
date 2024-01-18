package main

import (
	"os"

	"github.com/mjdusa/chatgpt-api/internal/runner"
)

func main() {
	os.Exit(runner.Run())
}
