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
	DialoutSettings DialoutSettings        `json:"dialout_settings"`
	DialinSettings  DialinSettings         `json:"dialin_settings"`
	WebhookTools    map[string]interface{} `json:"webhook_tools"`
}

type ServiceOptions struct {
	GeminiLive struct {
		Voice string `json:"voice"`
	} `json:"gemini_live"`
}

type Services struct {
	LLM string `json:"llm"`
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

// BuildRequestBody creates a request body for the Daily API with the given phone number and prompt
func BuildRequestBody(phoneNumber, prompt, geminiApiKey string, recordingConfig RecordingConfig) (*DailyRequestBody, error) {
	return &DailyRequestBody{
		BotProfile:  "gemini_multimodal_live_2024_12",
		MaxDuration: 300,
		ServiceOptions: ServiceOptions{
			GeminiLive: struct {
				Voice string `json:"voice"`
			}{
				Voice: "Aoede",
			},
		},
		Services: Services{
			LLM: "gemini_live",
		},
		Config: []ConfigItem{
			{
				Service: "llm",
				Options: []OptionItem{
					{
						Name: "initial_messages",
						Value: []Message{
							{
								Role:    "user",
								Content: prompt,
							},
						},
					},
					{
						Name:  "run_on_config",
						Value: true,
					},
				},
			},
		},
		APIKeys: map[string]string{
			"gemini_live": geminiApiKey,
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
		DialoutSettings: DialoutSettings{
			PhoneNumber: phoneNumber,
		},
		WebhookTools: make(map[string]interface{}),
	}, nil
}
