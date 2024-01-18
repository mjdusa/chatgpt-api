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
	"github.com/mjdusa/chatgpt-api/internal/version"
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

func GetParameters() (string, string, string, bool, bool) {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	flagSet.SetOutput(os.Stderr)

	var auth string
	var org string
	var logFileName string
	var dbg bool
	var verbose bool

	// add flags
	flagSet.StringVar(&auth, "auth", "", "OpenAI Auth Token")
	flagSet.StringVar(&org, "org", "", "OpenAI Org Token")
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

	return auth, org, logFileName, verbose, dbg
}

type PostRequestBody struct {
	Prompt      string  `json:"prompt"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type Message struct {
	Content      *string `json:"content"`
	FunctionCall string  `json:"function_call"`
	Role         string  `json:"role"`
}

type Choices struct {
	FinishReason string  `json:"finish_reason"`
	Index        float64 `json:"index"`
	Message      Message `json:"message"`
}

type Usage struct {
	CompletionTokens int64 `json:"completion_tokens"`
	PromptTokens     int64 `json:"prompt_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type PostResponseBody struct {
	Id      string    `json:"id"`
	Choices []Choices `json:"choices"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Object  string    `json:"object"`
	Usage   Usage     `json:"usage"`
}

func Run() int {
	ctx := context.Background()
	auth, org, logFileName, verbose, dbg := GetParameters()

	logFile, err := os.OpenFile(logFileName, DefaultFileAccess, DefaultFileMode)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		log.Fatal(err)
	}

	logger := log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	for {
		pmt := prompt("Ask: ")
		if pmt == "" {
			break
		}

		logger.Printf("Ask: %s", pmt)

		choices, err := ask(ctx, auth, org, pmt, verbose, dbg)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			logger.Fatalf("error: %v", err)
			continue
		}

		fmt.Println("ChatGPT:")
		for idx, choice := range *choices {
			fmt.Printf("[%d]: %v\n", idx, choice.Message.Content)
			logger.Printf("[%d]: %v", idx, choice.Message.Content)
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

func ask(ctx context.Context, auth string, org string, prompt string, verbose bool, dbg bool) (*[]Choices, error) {
	opts := []ghc.ClientOption{
		ghc.WithDefaultHeaders(),
		ghc.WithTimeout(DefaultTimeout),
	}

	if dbg {
		fmt.Printf("ClientOption: %+v\n", opts)
	}

	client := ghc.New(DefaultBaseURL, opts...)

	postRequestBody := PostRequestBody{
		Prompt:      prompt,
		Model:       DefaultModel,
		MaxTokens:   DefaultMaxTokens,
		Temperature: DefaultTemperature,
	}

	if dbg {
		fmt.Printf("postRequestBody: %+v\n", postRequestBody)
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

	if len(org) > 0 {
		reqOpts = append(reqOpts, ghc.WithHeader("OpenAI-Organization", org))
	}

	if dbg {
		fmt.Printf("[]Option: %+v\n", reqOpts)
	}

	response, err := client.Post(ctx, "/v1/completions", reqOpts...)
	if err != nil {
		return nil, err
	}

	var postResponseBody PostResponseBody
	if err := response.Unmarshal(&postResponseBody); err != nil {
		return nil, err
	}

	if dbg {
		fmt.Printf("postResponseBody: %+v\n", postResponseBody)
	}

	if !response.Ok() {
		if dbg {
			fmt.Printf("Response: %+v\n", response)
		}

		return nil, fmt.Errorf("error - HTTP Response Status Code: %s", response.Get().Status)
	}

	return &postResponseBody.Choices, nil
}
