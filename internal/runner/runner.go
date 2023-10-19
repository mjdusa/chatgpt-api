package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	ghc "github.com/bozd4g/go-http-client"
	"github.com/mdonahue-godaddy/chatgpt-api/internal/version"
)

const (
	OsExistCode             int           = 1
	DefaultBaseURL          string        = "https://api.openai.com"
	DefaultModel            string        = "text-davinci-003"
	DefaultMaxTokens        int           = 4000
	DefaultTemperature      float64       = 1.0
	DefaultFileMode         os.FileMode   = 0600
	DefaultFileAccess       int           = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	DefaultTimeout          time.Duration = time.Second * 5
	HTTPHeaderAccept        string        = "Accept"
	HTTPHeaderContent       string        = "Content-Type"
	HTTPHeaderAppJSON       string        = "application/json"
	HTTPHeaderAuthorization string        = "Authorization"
)

var panicOnExit = false // Set to true to tell Exit() to Panic rather than
// call os.Exit() - should ONLY be used for testing

func Exit(code int) {
	if panicOnExit {
		panic(fmt.Sprintf("PanicOnExit is true, code=%d", code))
	}

	os.Exit(code)
}

func GetParameters() (string, string, bool, bool) {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	flagSet.SetOutput(os.Stderr)

	var auth string
	var logFileName string
	var dbg bool
	var verbose bool

	// add flags
	flagSet.StringVar(&auth, "auth", "", "GitHub Auth Token")
	flagSet.StringVar(&logFileName, "log", "ChatGPT.log", "Log File Name")
	flagSet.BoolVar(&dbg, "debug", false, "Log Debug")
	flagSet.BoolVar(&verbose, "verbose", false, "Show Verbose Logging")

	// Parse the flags
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		flagSet.Usage()
		Exit(OsExistCode)
	}

	if len(auth) == 0 {
		flagSet.Usage()
		Exit(OsExistCode)
	}

	if verbose {
		fmt.Println(version.GetVersion())
	}

	if dbg {
		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			fmt.Println(buildInfo.String())
		}
	}

	return auth, logFileName, dbg, verbose
}

type PostRequestBody struct {
	Prompt      string  `json:"prompt"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type Choices struct {
	Text string `json:"text"`
}

type PostResponseBody struct {
	Choices []Choices `json:"choices"`
}

func Run() int {
	ctx := context.Background()
	auth, logFileName, _, _ := GetParameters()

	logFile, err := os.OpenFile(logFileName, DefaultFileAccess, DefaultFileMode)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	for {
		pmt := prompt("Ask: ")

		if pmt == "" {
			break
		}

		logger.Printf("Ask: %s", pmt)

		choices, err := ask(ctx, auth, pmt)
		if err != nil {
			logger.Fatalf("error: %v", err)
		}

		fmt.Println("ChatGPT:")
		for idx, choice := range *choices {
			fmt.Printf("[%d]: %s", idx, choice.Text)
			logger.Printf(choice.Text)
		}
	}

	return 0
}

func prompt(text string) string {
	var str string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, text+" ")
		str, _ = r.ReadString('\n')
		if str != "" {
			break
		}
	}
	return strings.TrimSpace(str)
}

func ask(ctx context.Context, auth string, prompt string) (*[]Choices, error) {
	opts := []ghc.ClientOption{
		ghc.WithDefaultHeaders(),
		ghc.WithTimeout(DefaultTimeout),
	}

	client := ghc.New(DefaultBaseURL, opts...)

	postRequestBody := PostRequestBody{
		Prompt:      prompt,
		Model:       DefaultModel,
		MaxTokens:   DefaultMaxTokens,
		Temperature: DefaultTemperature,
	}

	requestBody, err := json.Marshal(postRequestBody)
	if err != nil {
		return nil, err
	}

	reqOpts := []ghc.Option{
		ghc.WithBody(requestBody),
		ghc.WithHeader(HTTPHeaderAccept, HTTPHeaderAppJSON),
		ghc.WithHeader(HTTPHeaderContent, HTTPHeaderAppJSON),
		ghc.WithHeader(HTTPHeaderAuthorization, fmt.Sprintf("Bearer %s", auth)),
	}

	response, err := client.Post(ctx, "/v1/completions", reqOpts...)
	if err != nil {
		return nil, err
	}

	var postResponseBody PostResponseBody
	if err := response.Unmarshal(&postResponseBody); err != nil {
		return nil, err
	}

	return &postResponseBody.Choices, nil
}
