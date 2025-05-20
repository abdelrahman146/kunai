package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

func AskOllama(serverURL, model, prompt string) (string, error) {
	client := &http.Client{
		Timeout: time.Minute * 2,
	}
	reqPayload := generateRequest{
		Model:  model,
		Prompt: strings.TrimSpace(prompt),
		Stream: false,
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	resp, err := client.Post(fmt.Sprintf("%s/api/generate", serverURL), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	var body map[string]interface{}
	err = json.Unmarshal(respBody, &body)
	if err != nil {
		return "", err
	}
	if body["error"] != nil {
		return "", errors.New(body["error"].(string))
	}
	if body["response"] != nil {
		return body["response"].(string), nil
	}
	return "", errors.New("no response from ollama")
}
