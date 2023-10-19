package main

import (
	"os"

	"github.com/mdonahue-godaddy/chatgpt-api/internal/runner"
)

func main() {
	os.Exit(runner.Run())
}
