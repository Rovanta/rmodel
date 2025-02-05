### ChatAgent: With `Function Calling`

ChatAgent takes a list of chat messages as input and outputs new messages to this list. In this example, OpenAI's `function calling` feature is utilized. It is recommended to use in models facilitated with the `function calling` feature.

- [Chat Agent With Tools](./main.go): An example of creating a chat agent from scratch using tools.

<img src="./chat-agent-with-tools.png" width="476" height="238" alt="">

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/Rovanta/rmodel"
)

func main() {
	bp := rModel.NewBrainPrint()
	// add neuron
	bp.AddNeuron("llm", chatLLM)
	bp.AddNeuron("action", callTools)

	/* This example omits error handling */
	// add entry link
	_, _ = bp.AddEntryLink("llm")

	// add link
	continueLink, _ := bp.AddLink("llm", "action")
	_, _ = bp.AddLink("action", "llm")

	// add end link
	endLink, _ := bp.AddEndLink("llm")

	// add link to cast group of a neuron
	_ = bp.AddLinkToCastGroup("llm", "continue", continueLink)
	_ = bp.AddLinkToCastGroup("llm", "end", endLink)
	// bind cast group select function for neuron
	_ = bp.BindCastGroupSelectFunc("llm", llmNext)

	// build brain
	brain := bp.Build()
	// set memory and trig all entry links
	_ = brain.EntryWithMemory(
		"messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"}})

	// block process util brain sleeping
	brain.Wait()

	messages, _ := json.Marshal(brain.GetMemory("messages"))
	fmt.Printf("messages: %s\n", messages)
}

// describe the function & its inputs
var tools = []openai.Tool{
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_current_weather",
			Description: "Get the current weather in a given location",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"location": {
						Type:        jsonschema.String,
						Description: "The city and state, e.g. San Francisco, CA",
					},
					"unit": {
						Type: jsonschema.String,
						Enum: []string{"celsius", "fahrenheit"},
					},
				},
				Required: []string{"location"},
			},
		},
	},
}

func chatLLM(b rModel.BrainRuntime) error {
	fmt.Println("run here chatLLM...")

	// get need info form memory
	messages, _ := b.GetMemory("messages").([]openai.ChatCompletionMessage)

	ctx := context.Background()
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo0125,
			Messages: messages,
			Tools:    tools,
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		return fmt.Errorf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
	}

	msg := resp.Choices[0].Message
	fmt.Printf("LLM response: %+v\n", msg)
	messages = append(messages, msg)
	_ = b.SetMemory("messages", messages)

	return nil
}

func callTools(b rModel.BrainRuntime) error {
	fmt.Println("run here callTools...")

	// get need info form memory
	messages, _ := b.GetMemory("messages").([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]

	for _, call := range lastMsg.ToolCalls {
		if call.Function.Name == "get_current_weather" {
			fmt.Printf("call tool [%s] with arguments [%s]\n", call.Function.Name, call.Function.Arguments)
			toolCalledResp := "Sunny and 80 degrees."

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    toolCalledResp,
				Name:       lastMsg.ToolCalls[0].Function.Name,
				ToolCallID: lastMsg.ToolCalls[0].ID,
			})
		}
	}
	_ = b.SetMemory("messages", messages)

	return nil
}

func llmNext(b rModel.BrainRuntime) string {
	if !b.ExistMemory("messages") {
		return "end"
	}
	messages, _ := b.GetMemory("messages").([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]
	if len(lastMsg.ToolCalls) == 0 { // no need to call any tools
		return "end"
	}

	return "continue"
}

```