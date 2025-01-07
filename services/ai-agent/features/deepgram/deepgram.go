package deepgram

import (
	client "github.com/deepgram/deepgram-go-sdk/pkg/client/live"
)

type Deepgram struct {
	deepgram *client.Client
}

func New() *Deepgram {
	client.InitWithDefault()
}
