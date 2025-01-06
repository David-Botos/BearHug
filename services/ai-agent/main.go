package main

import (
	"context"
	"fmt"
	"log"

	"github.com/henomis/lingoose/assistant"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
	"github.com/joho/godotenv"
)

func main() {
	
	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }


	openaiLLM := assistant.New(openai.New().
    WithTemperature(0).
    WithModel(openai.GPT4))

	myThread := thread.New().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent("You are a powerful AI assistant."),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, how are you?"),
		),
	)
	
	err := openaiLLM.RunWithThread(context.Background(), myThread)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(myThread)
}