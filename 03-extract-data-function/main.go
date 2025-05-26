package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

	data, err := ExtractDataFromResults(results)
	if err != nil {
		fmt.Println("Error extracting data:", err)
		return
	}
	fmt.Println("Extracted Data:")
	for i, item := range data {
		fmt.Printf("Item %d:\n", i+1)
		fmt.Printf("  Title: %s\n", item["title"])
		fmt.Printf("  URL: %s\n", item["url"])
		fmt.Printf("  Summary: %s\n", item["summary"])
	}

}

func ExtractDataFromResults(results []string) ([]map[string]any, error) {
	//model := "ai/qwen2.5:latest" 
	//model := "ai/qwen2.5:0.5B-F16"
	//model := "ai/qwen2.5:1.5B-F16"
	// NOTE: ai/qwen2.5:0.5B-F16 and ai/qwen2.5:1.5B-F16 are too small for this task
	model := "ai/qwen2.5:3B-F16"

	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "The first line of the section",
				},
				"url": map[string]any{
					"type": "string",
				},
				"summary": map[string]any{
					"type":        "string",
					"description": "A short summary of the section",
				},
			},
			"required": []string{"title", "url", "summary"},
		},
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "search_results",
		Description: openai.String("Notable information about search results"),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	Riker, _ := robby.NewAgent(
		robby.WithDMRClient(
			context.Background(),
			"http://model-runner.docker.internal/engines/llama.cpp/v1/",
		),
		robby.WithParams(
			openai.ChatCompletionNewParams{
				Model: model,
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.SystemMessage(strings.Join(results, "\n")),
					openai.UserMessage("give me the list of the results."),
				},
				Temperature: openai.Opt(0.0),
				ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
					OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
						JSONSchema: schemaParam,
					},
				},
			},
		),
	)
	jsonResults, err := Riker.ChatCompletion()
	if err != nil {
		return nil, err
	}

	// Transform the json string into a map
	var jsonResultsMap []map[string]any
	err = json.Unmarshal([]byte(jsonResults), &jsonResultsMap)
	if err != nil {
		return nil, err
	}
	return jsonResultsMap, nil
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
