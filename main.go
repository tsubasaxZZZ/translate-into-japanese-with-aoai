package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Struct to represent the message content
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Struct to represent the message object
type Message struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

type TranslationSchema struct {
	Type        string `json:"type"`
	Original    string `json:"Original"`
	Translation string `json:"Translation"`
}
type JsonSchema struct {
	Description string                 `json:"description"`
	Name        string                 `json:"name"`
	Schema      map[string]interface{} `json:"schema"`

	Strict bool `json:"strict"`
}
type ResponseFormatJsonSchema struct {
	Type       string     `json:"type"`
	JsonSchema JsonSchema `json:"json_schema"`
}

// Struct for the payload
type Payload struct {
	Messages       []Message                `json:"messages"`
	Temperature    float64                  `json:"temperature"`
	TopP           float64                  `json:"top_p"`
	MaxTokens      int                      `json:"max_tokens"`
	ResponseFormat ResponseFormatJsonSchema `json:"response_format"`
}

// Response struct
type ContentFilterResults struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity"`
}

type ResponseMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Choice struct {
	ContentFilterResults map[string]ContentFilterResults `json:"content_filter_results"`
	FinishReason         string                          `json:"finish_reason"`
	Index                int                             `json:"index"`
	Logprobs             interface{}                     `json:"logprobs"`
	Message              ResponseMessage                 `json:"message"`
}

type Response struct {
	Choices             []Choice               `json:"choices"`
	Created             float64                `json:"created"`
	ID                  string                 `json:"id"`
	Model               string                 `json:"model"`
	Object              string                 `json:"object"`
	PromptFilterResults []interface{}          `json:"prompt_filter_results"`
	SystemFingerprint   string                 `json:"system_fingerprint"`
	Usage               map[string]interface{} `json:"usage"`
}

func main() {
	// get from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatalf("OPENAI_API_KEY environment variable is not set")
	}

	// Set up headers
	headers := map[string]string{
		"Content-Type": "application/json",
		"api-key":      apiKey,
	}

	// Prepare payload
	systemMessage := Message{
		Role: "system",
		Content: []MessageContent{
			{
				Type: "text",
				Text: "You are an AI assistant that helps people find information.",
			},
		},
	}

	userMessage := Message{
		Role: "user",
		Content: []MessageContent{
			{
				Type: "text",
				Text: "日本語に翻訳してください。\n What is the capital of China? ",
			},
		},
	}

	// TranslationSchema := TranslationSchema{
	// 	Type:        "object",
	// 	Original:    "string",
	// 	Translation: "string",
	// }

	payload := Payload{
		Messages:    []Message{systemMessage, userMessage},
		Temperature: 0.7,
		TopP:        0.95,
		MaxTokens:   800,
		ResponseFormat: ResponseFormatJsonSchema{
			Type: "json_schema",
			JsonSchema: JsonSchema{
				Description: "Translation schema",
				Name:        "translation",
				//Schema:      TranslationSchema,
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"Original":   map[string]string{"type": "string", "description": "The original text to be translated."},
						"Translated": map[string]string{"type": "string", "description": "The translated text."},
					},
					"required":             []string{"Original", "Translated"},
					"additionalProperties": false,
				},
				Strict: true,
			},
		},
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshal payload: %v", err)
	}

	// Send request
	endpoint := "https://tsunomuropenai.openai.azure.com/openai/deployments/gpt-4o/chat/completions?api-version=2024-08-01-preview"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	log.Println(req.Body)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make the request: %v", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Request failed with status code: %v %v", resp.StatusCode, string(body))
	}

	// Read and handle the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// get status and body
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Print the response
	fmt.Println(string(body))

	// Marshal the response
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(response.Choices) > 0 {
		var contentMap map[string]string
		err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &contentMap)
		if err != nil {
			log.Fatalf("Failed to unmarshal content: %v", err)
		}

		translated, ok := contentMap["Translated"]
		if !ok {
			log.Fatalf("Failed to get Translated from content")
		}
		fmt.Println(translated)
	} else {
		log.Fatalf("No choices found in response")
	}
}
