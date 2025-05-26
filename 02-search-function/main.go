package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/sea-monkeys/robby"
)

func main() {

	results, err := WebSearch("Search the latests information about Docker. (Only 3 results)")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("🎉", len(results), "\nResults:\n", results)

}

func WebSearch(query string) ([]string, error) {
	// NOTE: trying to use a smaller model to increase performance and reduce costs
	model := "ai/qwen2.5:0.5B-F16" // "ai/qwen2.5:latest"
	Bob, _ := robby.NewAgent(
		robby.WithDMRClient(
			context.Background(),
			"http://model-runner.docker.internal/engines/llama.cpp/v1/",
		),
		robby.WithParams(
			openai.ChatCompletionNewParams{
				Model: model,
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(query),
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
		return nil, err
	}
	// Display the tool calls in JSON format
	//toolCallsJSON, _ := Bob.ToolCallsToJSON()
	//fmt.Println("Tool Calls:", toolCallsJSON)

	// Execute the tool calls and get the results
	results, _ := Bob.ExecuteMCPToolCalls()
	return results, nil
}
