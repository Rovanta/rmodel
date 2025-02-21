### ChatAgetn: With `Reflect`

Here provided an easy example about ChatAgetn with Reflect function. In this example, the AI is required to write a novel within the given `maxWordsNum`. Everytime it outputs new novel, a "Judge Man" will check if the story is accecptable. If true, it pass, otherwise it will withdraw it.

- [Chat Agent With Tools](./main.go): An example of creating a chat agent with reflect from scratch.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Rovanta/rmodel"
	"github.com/Rovanta/rmodel/brainlocal"
	"github.com/Rovanta/rmodel/processor"
	"github.com/sashabaranov/go-openai"
)

var maxWordsNum int = 100

func main() {
	bp := rModel.NewBlueprint()

	// add neuron
	llm := bp.AddNeuron(chatLLM)
	reflect := bp.AddNeuron(judgeMan)

	/* This example omits error handling */
	// add entry link
	_, _ = bp.AddEntryLinkTo(llm)
	
	// add link
	handupLink, _ := bp.AddLink(llm, reflect)
	unpassLink, _ := bp.AddLink(reflect, llm)

	// add end link
	endLink, _ := bp.AddEndLinkFrom(reflect)
	
	// add link to cast group of a neuron
	_ = llm.AddCastGroup("handup", handupLink)
	_ = llm.AddCastGroup("unpass", unpassLink)
	_ = llm.AddCastGroup("end", endLink)
	// bind cast group select function for neuron
	reflect.BindCastGroupSelectFunc(judgeNext)

	// buid brain
	brain := brainlocal.BuildBrain(bp)
	// set memory and trig all entry links
	_ = brain.EntryWithMemory(
		"messages", []openai.ChatCompletionMessage{{
				Role: openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Write a novel within %d words, please.", maxWordsNum),
		}})
	
	// block process until brain sleeping
	brain.Wait()

	messages, _ := json.Marshal(brain.GetMemory("messages"))
	fmt.Printf("messages: %s\n", messages)
}

func chatLLM(bc processor.BrainContext) error {
	fmt.Println("run here chatLLM...")

	// get need info from memory
	messages, _ := bc.GetMemory("messages").([]openai.ChatCompletionMessage)

	ctx := context.Background()
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo0125,
			Messages: messages,
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		return fmt.Errorf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
	}

	msg := resp.Choices[0].Message
	fmt.Printf("LLM response: %+v\n", msg)
	messages = append(messages, msg)
	_ = bc.SetMemory("messages", messages)

	return nil
}

func judgeMan(bc processor.BrainContext) error {
	fmt.Println("run here judgeMan")

	// get need info from memory
	messages, _ := bc.GetMemory("messages").([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]

	passOrNot := fmt.Sprintf("Unpass. The number of words exceeds %d. Try again!", maxWordsNum)
	if len(lastMsg.Content) <= maxWordsNum {
		passOrNot = "Pass. Thank you for your assistance!"
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleUser,
		Content:    passOrNot,
	})

	_ = bc.SetMemory("messages", messages)

	return nil
}

func judgeNext(bcr processor.BrainContextReader) string {
	if !bcr.ExistMemory("messages") {
		return "end"
	}
	messages, _ := bcr.GetMemory("messages").([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]
	if lastMsg.Content[0] == 'P' { // pass
		return "end"
	}

	return "unpass"
}
```
