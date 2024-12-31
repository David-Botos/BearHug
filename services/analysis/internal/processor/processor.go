package processor

import "log"

type Transcript struct {
	RoomURL string `json:"room_url"`
	Content string `json:"transcript"`
}

type Processor struct {}

func New() *Processor {
	return &Processor{}
}

func (p *Processor) ProcessTranscript(transcript Transcript) {
	log.Printf("Started processing transcript from room: %s", transcript.RoomURL)

	// Add your transcript processing logic here
	// For example:
	// - Analyze the content
	// - Store results in a database
	// - Send notifications
	// - Update processing status

	log.Printf("Completed processing transcript from room: %s", transcript.RoomURL)
}