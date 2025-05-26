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

	results, err := WebSearch("What is Docker Compose? (Only 3 results)")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	data, err := ExtractDataFromResults(results)
	if err != nil {
		fmt.Println("Error extracting data:", err)
		return
	}

	content, err := FetchContent(data)

	if err != nil {
		fmt.Println("Error fetching content:", err)
		return
	}

	_, err = Summarize(`/no_think [Brief]
		Make a clear, and structured summaryt with the provided information.
		- Use markdown format.
		- Provide only verified refrences (URLs).
		- Stay focused and do not repeat the same information.	
		- Do not use any other external information.
		- Do not include the error messages in the report.
	`, content)

	if err != nil {
		fmt.Println("Error summarizing content:", err)
		return
	}

}

func Summarize(instructions string, content []string) (string, error) {

	//model := "ai/qwen2.5:latest"
	//model := "ai/qwen2.5:1.5B-F16"
	//model := "ai/qwen2.5:3B-F16"
	//model := "ai/mistral:latest"
	//model := "ai/mistral-nemo"
	//model := "ai/llama3.2:latest"
	model := "ai/qwen3:latest"

	Milo, _ := robby.NewAgent(
		robby.WithDMRClient(
			context.Background(),
			"http://model-runner.docker.internal/engines/llama.cpp/v1/",
		),
		robby.WithParams(
			openai.ChatCompletionNewParams{
				Model: model,
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.SystemMessage(strings.Join(content, "\n")),
					openai.UserMessage(instructions),
				},
				Temperature: openai.Opt(0.0),
				TopP:        openai.Opt(0.3), // Lowering TopP to reduce randomness
				// NOTE: To limit hallucinations and obtain more reliable responses,
				// lower both the “temperature” and “top_p” parameters.
				// This forces the model to choose the safest and most predictable answers.

			},
		),
	)
	result, err := Milo.ChatCompletionStream(func(self *robby.Agent, content string, err error) error {

		fmt.Print(content)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error in ChatCompletionStream: %w", err)
	}
	fmt.Println("\nReport Generated ✅")
	return result, nil
}

func FetchContent(data []map[string]any) ([]string, error) {

	//model := "ai/qwen2.5:latest"
	model := "ai/qwen2.5:0.5B-F16"
	//model := "ai/qwen2.5:1.5B-F16"
	//model := "ai/qwen2.5:3B-F16"

	Bill, _ := robby.NewAgent(
		robby.WithDMRClient(
			context.Background(),
			"http://model-runner.docker.internal/engines/llama.cpp/v1/",
		),
		robby.WithParams(
			openai.ChatCompletionNewParams{
				Model:             model,
				Messages:          []openai.ChatCompletionMessageParamUnion{},
				Temperature:       openai.Opt(0.0),
				ParallelToolCalls: openai.Bool(true),
			},
		),
		robby.WithMCPClient(robby.WithDockerMCPToolkit()),
		robby.WithMCPTools([]string{"fetch"}),
	)

	prompt := ""
	for _, result := range data {
		prompt += fmt.Sprintf("Fetch this URL: %s\n", result["url"])
	}

	fmt.Println("🛠️ Prompt for tool calls:")
	fmt.Println(prompt)

	Bill.Params.Messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt),
	}

	_, err := Bill.ToolsCompletion()
	if err != nil {
		return nil, fmt.Errorf("error in tools completion: %w", err)
	}

	toolCallsJSON, _ := Bill.ToolCallsToJSON()
	fmt.Println("Tool Calls:", toolCallsJSON)

	results, err := Bill.ExecuteMCPToolCalls()
	if err != nil {
		return nil, fmt.Errorf("error executing tool calls: %w", err)
	}

	fmt.Println("Fetched Content completed ✅")

	return results, nil
}

func ExtractDataFromResults(results []string) ([]map[string]any, error) {

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

	fmt.Println("📝 JSON Results:\n", jsonResults)

	// Transform the json string into a map
	var jsonResultsMap []map[string]any
	err = json.Unmarshal([]byte(jsonResults), &jsonResultsMap)
	if err != nil {
		return nil, err
	}

	fmt.Println("Extracted Data from Results completed ✅")
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
		//robby.WithMCPTools([]string{"brave_web_search"}),
		// DuckDuckGo NOTE: this tool is rate-limited
		robby.WithMCPTools([]string{"search"}),
	)

	// Execute the tool calls == tool calls detection
	_, err := Bob.ToolsCompletion()
	if err != nil {
		return nil, err
	}

	// DEBUG: Uncomment the following lines to display the tool calls in JSON format
	// Display the tool calls in JSON format
	toolCallsJSON, _ := Bob.ToolCallsToJSON()
	fmt.Println("Tool Calls:", toolCallsJSON)

	// Execute the tool calls and get the results
	results, _ := Bob.ExecuteMCPToolCalls()

	// Display the results
	for _, result := range results {
		fmt.Println(result)
	}
	fmt.Println("Web Search Results completed ✅")

	return results, nil
}
