package elevenlabs

import (
	"context"
	"log"
	"os/exec"
	"time"

	"github.com/haguro/elevenlabs-go"
)

type ElevenLabs struct {
	elevenlabs *elevenlabs.Client
}

func New(key string) *ElevenLabs {
	client := elevenlabs.NewClient(context.Background(), key, 30*time.Second)
	elevenlabs.SetTimeout(1 * time.Minute)
	return &ElevenLabs{elevenlabs: client}
}

func (eleven *ElevenLabs) StreamAudio(message string) {
	// We'll use mpv to play the audio from the stream piped to standard input
	cmd := exec.CommandContext(context.Background(), "mpv", "--no-cache", "--no-terminal", "--", "fd://0")

	// Get a pipe connected to the mpv's standard input
	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Attempt to run the command in a separate process
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := elevenlabs.TextToSpeechStream(pipe,
		"",
		elevenlabs.TextToSpeechRequest{Text: message, ModelID: ""}); err != nil {
		log.Fatal(err)
	}

	// Close the pipe when all stream has been copied to the pipe
	if err := pipe.Close(); err != nil {
		log.Fatalf("Could not close pipe: %s", err)
	}
	log.Print("Streaming finished.")

	// Wait for mpv to exit. With the pipe closed, it will do that as
	// soon as it finishes playing
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	log.Print("All done.")
}
