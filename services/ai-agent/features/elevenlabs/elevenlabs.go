package elevenlabs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/haguro/elevenlabs-go"
)

// ReaderWithClose wraps a bytes.Reader to implement io.ReadCloser
type ReaderWithClose struct {
	*bytes.Reader
}

func (r *ReaderWithClose) Close() error {
	// No-op Close method (not needed in this case)
	return nil
}

type ElevenLabs struct {
	elevenlabs *elevenlabs.Client
}

func New(key string) *ElevenLabs {
	client := elevenlabs.NewClient(context.Background(), key, 30*time.Second)
	elevenlabs.SetAPIKey(key)
	elevenlabs.SetTimeout(1 * time.Minute)
	return &ElevenLabs{elevenlabs: client}
}

func (eleven *ElevenLabs) StreamAudio(message string) {

	fmt.Println("Message", message)

	// Create a buffer to store audio data
	mp3Buffer := &bytes.Buffer{}

	// Fetch audio data from ElevenLabs API
	err := elevenlabs.TextToSpeechStream(
		mp3Buffer,
		"cjVigY5qzO86Huf0OWal",
		elevenlabs.TextToSpeechRequest{
			Text:    message,
			ModelID: "eleven_multilingual_v1",
		},
	)

	if err != nil {
		log.Fatalf("Error streaming audio: %v", err)
	}

	// Create a ReaderWithClose from the bytes.Buffer
	closableReader := &ReaderWithClose{Reader: bytes.NewReader(mp3Buffer.Bytes())}

	// Decode the MP3 data using beep's MP3 decoder
	streamer, format, err := mp3.Decode(closableReader)
	if err != nil {
		log.Fatalf("Error decoding MP3: %v", err)
	}

	// Initialize the speaker with the correct format
	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		log.Fatalf("Failed to initialize speaker: %v", err)
	}

	// Calculate the playback duration in seconds
	playbackDuration := time.Duration(streamer.Len()) * time.Second / time.Duration(format.SampleRate)

	// Play the audio
	log.Println("Playing audio...")
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		log.Println("Audio playback finished.")
	})))

	// Wait for playback to finish
	time.Sleep(playbackDuration)
	log.Println("Playback completed.")
}
