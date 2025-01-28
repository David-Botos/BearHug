package daily

// DailyRequestBody represents the main request structure
type DailyRequestBody struct {
	BotProfile      string                 `json:"bot_profile"`
	MaxDuration     int                    `json:"max_duration"`
	ServiceOptions  ServiceOptions         `json:"service_options"`
	Services        Services               `json:"services"`
	Config          []ConfigItem           `json:"config"`
	APIKeys         map[string]string      `json:"api_keys"`
	RecordSettings  RecordingSettings      `json:"recording_settings"`
	DialoutSettings []DialoutSettings      `json:"dialout_settings"`
	DialinSettings  DialinSettings         `json:"dialin_settings"`
	WebhookTools    map[string]interface{} `json:"webhook_tools"`
}

type Deepgram struct {
	Dictation       bool   `json:"dictation,omitempty"`
	FillerWords     bool   `json:"filler_words,omitempty"`
	Numerals        bool   `json:"numerals,omitempty"`
	ProfanityFilter bool   `json:"profanity_filter,omitempty"`
	Punctuate       bool   `json:"punctuate,omitempty"`
	Keywords        string `json:"keywords,omitempty"`
	Redact          string `json:"redact,omitempty"`
	Replace         string `json:"replace,omitempty"`
}

type Anthropic struct {
	Model string `json:"model,omitempty"`
}

type Together struct {
	Model string `json:"model,omitempty"`
}

type Cartesia struct {
	Voice      string `json:"voice,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
}

type ServiceOptions struct {
	Deepgram Deepgram `json:"deepgram,omitempty"`
	// Anthropic Anthropic `json:"anthropic,omitempty"`
	Together Together `json:"together,omitempty"`
	Cartesia Cartesia `json:"cartesia,omitempty"`
}

type Services struct {
	STT string `json:"stt"`
	LLM string `json:"llm"`
	TTS string `json:"tts"`
}

type ConfigItem struct {
	Service string       `json:"service"`
	Options []OptionItem `json:"options"`
}

type OptionItem struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RecordingSettings struct {
	Type             string           `json:"type"`
	RecordingsBucket RecordingsBucket `json:"recordings_bucket"`
}

type RecordingsBucket struct {
	AllowAPIAccess           bool   `json:"allow_api_access"`
	AllowStreamingFromBucket bool   `json:"allow_streaming_from_bucket"`
	AssumeRoleARN            string `json:"assume_role_arn"`
	BucketName               string `json:"bucket_name"`
	BucketRegion             string `json:"bucket_region"`
}

type DialoutSettings struct {
	PhoneNumber string `json:"phoneNumber"`
}

type DialinSettings struct {
	CallID     string `json:"call_id"`
	CallDomain string `json:"call_domain"`
}

type RecordingConfig struct {
	AssumeRoleARN string
	BucketName    string
	BucketRegion  string
}

type ServiceKeys struct {
	STT string
	LLM string
	TTS string
}

// BuildRequestBody creates a request body for the Daily API with the given phone number and prompt
func BuildRequestBody(phoneNumber, prompt string, serviceKeys ServiceKeys, recordingConfig RecordingConfig) (*DailyRequestBody, error) {
	// Initialize DAC Service Options concisely
	DACServiceOptions := ServiceOptions{
		Deepgram: Deepgram{
			Dictation:       true,
			FillerWords:     false,
			Numerals:        true,
			ProfanityFilter: true,
			Punctuate:       true,
		},
		Together: Together{
			Model: "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
		},
		Cartesia: Cartesia{
			Voice: "5345cf08-6f37-424d-a5d9-8ae1101b9377",
		},
	}

	vadParams := map[string]interface{}{
		"stop_secs": 0.5,
	}

	DACServiceConfig := []ConfigItem{
		{
			Service: "vad",
			Options: []OptionItem{
				{
					Name:  "params",
					Value: vadParams,
				},
			},
		},
		{
			Service: "stt",
			Options: []OptionItem{
				{
					Name:  "model",
					Value: "nova-2-phonecall",
				},
			},
		},
		{
			Service: "llm",
			Options: []OptionItem{
				{
					Name:  "model",
					Value: "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
				},
				// {
				// 	Name:  "temperature",
				// 	Value: 0.7,
				// },
				{
					Name: "initial_messages",
					Value: []Message{
						{
							Role:    "system",
							Content: prompt,
						},
						{
							Role:    "user",
							Content: "Ok let's start the roleplay.  Begin speaking when I greet you:",
						},
					},
				},
				{Name: "run_on_config", Value: false},
			},
		},
		{
			Service: "tts",
			Options: []OptionItem{
				{
					Name:  "voice",
					Value: "5345cf08-6f37-424d-a5d9-8ae1101b9377",
				},
				{
					Name:  "speed",
					Value: "normal",
				},
				{
					Name:  "emotion",
					Value: []string{"positivity:high", "curiosity"},
				},
				{
					Name: "text_filter",
					Value: map[string]interface{}{
						"enable_text_filter": true,
						"filter_code":        true,
						"filter_tables":      true,
					},
				},
			},
		},
	}

	return &DailyRequestBody{
		BotProfile:     "voice_2024_10",
		MaxDuration:    900,
		ServiceOptions: DACServiceOptions,
		Services: Services{
			STT: "deepgram",
			LLM: "together",
			TTS: "cartesia",
		},
		Config: DACServiceConfig,
		APIKeys: map[string]string{
			"deepgram": serviceKeys.STT,
			"together": serviceKeys.LLM,
			"cartesia": serviceKeys.TTS,
		},
		RecordSettings: RecordingSettings{
			Type: "cloud",
			RecordingsBucket: RecordingsBucket{
				AllowAPIAccess:           true,
				AllowStreamingFromBucket: true,
				AssumeRoleARN:            recordingConfig.AssumeRoleARN,
				BucketName:               recordingConfig.BucketName,
				BucketRegion:             recordingConfig.BucketRegion,
			},
		},
		DialoutSettings: []DialoutSettings{
			{PhoneNumber: phoneNumber},
		},
		WebhookTools: make(map[string]interface{}),
	}, nil
}
