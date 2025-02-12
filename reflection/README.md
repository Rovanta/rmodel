### Reflection / Self-Critique

When output quality becomes a primary concern, a combination of self-reflection, self-critique, and external validation is often used to optimize the system output. The example below shows how to implement such a design.

- [Basic Reflection](./main.go): Adds a simple "reflect" step in the `Brain` to prompt your system for output modification.

<img src="./reflection-base.png" width="476" height="238">

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/Rovanta/rmodel"
)

func main() {
	bp := rModel.NewBrainPrint()
	// add neuron
	bp.AddNeuron("generation", generate)
	bp.AddNeuron("reflection", reflect)

	/* This example omits error handling */
	// add entry link
	_, _ = bp.AddEntryLink("generation")

	// add link
	continueLink, _ := bp.AddLink("generation", "reflection")
	_, _ = bp.AddLink("reflection", "generation")

	// add end link
	endLink, _ := bp.AddEndLink("generation")

	// add link to cast group of a neuron
	_ = bp.AddLinkToCastGroup("generation", "reflect", continueLink)
	_ = bp.AddLinkToCastGroup("generation", "end", endLink)
	// bind cast group select function for neuron
	_ = bp.BindCastGroupSelectFunc("generation", generationNext)

	// build brain
	brain := bp.Build()
	// set memory and trig all entry links
	_ = brain.EntryWithMemory(
		"messages", []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Generate an essay on the topicality of The Little Prince and its message in modern life",
			},
		})

	// block process util brain sleeping
	brain.Wait()

	messages, _ := json.Marshal(brain.GetMemory("messages"))
	fmt.Printf("messages: %s\n", messages)
}

func generate(b rModel.BrainRuntime) error {
	fmt.Println("generation assistant running...")

	// get messages form memory
	messages, _ := b.GetMemory("messages").([]openai.ChatCompletionMessage)

	prompt := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `You are an essay assistant tasked with writing excellent 5-paragraph essays.
	Generate the best essay possible for the user's request.
	If the user provides critique, respond with a revised version of your previous attempts.`,
	}

	ctx := context.Background()
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo0125,
			Messages: append([]openai.ChatCompletionMessage{prompt}, messages...),
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

func reflect(b rModel.BrainRuntime) error {
	fmt.Println("reflection assistant running...")

	// get messages form memory
	messages, _ := b.GetMemory("messages").([]openai.ChatCompletionMessage)
	roleReverse := func(msgs []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
		ret := []openai.ChatCompletionMessage{}
		for _, msg := range msgs {
			if msg.Role == openai.ChatMessageRoleAssistant {
				msg.Role = openai.ChatMessageRoleUser
			}
			if msg.Role == openai.ChatMessageRoleUser {
				msg.Role = openai.ChatMessageRoleAssistant
			}
			ret = append(ret, msg)
		}

		return ret
	}

	prompt := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `You are a teacher grading an essay submission. Generate critique and recommendations for the user's submission.
Provide detailed recommendations, including requests for length, depth, style, etc.`,
	}

	ctx := context.Background()
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo0125,
			// reverse role of user and assistant
			Messages: append([]openai.ChatCompletionMessage{prompt}, roleReverse(messages)...),
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		return fmt.Errorf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
	}

	msg := resp.Choices[0].Message
	fmt.Printf("LLM response: %+v\n", msg)
	msg.Role = openai.ChatMessageRoleUser // We treat the output of this as human feedback for the generator
	messages = append(messages, msg)
	_ = b.SetMemory("messages", messages)

	return nil
}

func generationNext(b rModel.BrainRuntime) string {
	if !b.ExistMemory("messages") {
		return "end"
	}
	messages, _ := b.GetMemory("messages").([]openai.ChatCompletionMessage)
	if len(messages) > 6 {
		return "end"
	}

	return "reflect"
}

```