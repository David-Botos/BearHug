package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const MAX_ANALYSIS_ROUTINES = 1

// Transcript represents the incoming transcript data
type Transcript struct {
	RoomURL   string `json:"room_url"`
	Content   string `json:"transcript"`
}

// TranscriptQueue manages the transcript processing
type TranscriptQueue struct {
	transcripts []Transcript
	mutex       sync.Mutex
	workChan    chan struct{}    // Channel to signal work is available
	workerWg    sync.WaitGroup   // WaitGroup to track active workers
	activeWorkers int            // Count of currently active workers
}

var queue TranscriptQueue

func main() {
	// Initialize the queue with a buffered channel
	queue = TranscriptQueue{
		transcripts: make([]Transcript, 0),
		workChan:    make(chan struct{}, MAX_ANALYSIS_ROUTINES),
	}

	// Define the handlers
	http.HandleFunc("/transcript", handleTranscript)
	http.HandleFunc("/status", handleStatus)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleTranscript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var transcript Transcript
	if err := json.NewDecoder(r.Body).Decode(&transcript); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	queue.mutex.Lock()
	// Add transcript to queue
	queue.transcripts = append(queue.transcripts, transcript)
	
	// Start a new worker if we haven't reached MAX_ANALYSIS_ROUTINES
	if queue.activeWorkers < MAX_ANALYSIS_ROUTINES {
		queue.activeWorkers++
		go startWorker()
	}
	
	// Signal that work is available
	queue.workChan <- struct{}{}
	queue.mutex.Unlock()

	log.Printf("Received and queued transcript from room: %s", transcript.RoomURL)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Transcript received and queued for processing",
	})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queue.mutex.Lock()
	response := map[string]interface{}{
		"pending_transcripts": len(queue.transcripts),
		"active_workers":      queue.activeWorkers,
	}
	queue.mutex.Unlock()

	json.NewEncoder(w).Encode(response)
}

func startWorker() {
	queue.workerWg.Add(1)
	defer queue.workerWg.Done()
	
	log.Printf("Started new analysis worker. Total workers: %d", queue.activeWorkers)

	for {
		// Wait for work signal
		_, ok := <-queue.workChan
		if !ok {
			// Channel closed, time to exit
			break
		}

		// Check if there's work to do
		queue.mutex.Lock()
		if len(queue.transcripts) == 0 {
			// Decrement active workers and exit if no work
			queue.activeWorkers--
			queue.mutex.Unlock()
			log.Printf("Worker shutting down. Remaining workers: %d", queue.activeWorkers)
			return
		}

		// Get the next transcript
		transcript := queue.transcripts[0]
		queue.transcripts = queue.transcripts[1:]
		queue.mutex.Unlock()

		// Process the transcript
		log.Printf("Processing transcript from room: %s", transcript.RoomURL)
		
		// put storage and analysis logic here
		
		log.Printf("Completed processing transcript from room: %s", transcript.RoomURL)
	}
}