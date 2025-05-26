package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/sea-monkeys/robby"
)

func main() {

	Bob, _ := robby.NewAgent(
		robby.WithDMRClient(
			context.Background(),
			"http://model-runner.docker.internal/engines/llama.cpp/v1/",
		),
		robby.WithParams(
			openai.ChatCompletionNewParams{
				Model: "ai/qwen2.5:latest",
				//Model: "ai/qwen2.5:0.5B-F16",
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(`
						Search the  latests information about Docker.
						(Only 3 results)
					`),
				},
				Temperature:       openai.Opt(0.0),
				ParallelToolCalls: openai.Bool(true),
			},
		),
		robby.WithMCPClient(robby.WithDockerMCPToolkit()),
		// Brave NOTE: this tool needs a valid API key
		robby.WithMCPTools([]string{"brave_web_search"}),
		// DuckDuckGo NOTE: this tool is rate-limited
		//robby.WithMCPTools([]string{"search"}),
	)

	// Execute the tool calls == tool calls detection
	_, err := Bob.ToolsCompletion()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Display the tool calls in JSON format
	toolCallsJSON, _ := Bob.ToolCallsToJSON()
	fmt.Println("Tool Calls:", toolCallsJSON)

	// Execute the tool calls and get the results
	results, _ := Bob.ExecuteMCPToolCalls()
	fmt.Println("🎉", len(results), "\nResults:\n", results)
}
