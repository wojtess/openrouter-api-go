package openrouterapigo

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenRouterClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:     apiKey,
		apiURL:     "https://openrouter.ai/api/v1",
		httpClient: &http.Client{},
	}
}

func NewOpenRouterClientFull(apiKey string, apiUrl string, client *http.Client) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:     apiKey,
		apiURL:     apiUrl,
		httpClient: client,
	}
}

func (c *OpenRouterClient) FetchChatCompletions(request Request) (*Response, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
		"Content-Type":  "application/json",
	}

	if request.Provider != nil {
		headers["HTTP-Referer"] = request.Provider.RefererURL
		headers["X-Title"] = request.Provider.SiteName
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.apiURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	output, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, output)
	}

	outputReponse := &Response{}
	err = json.Unmarshal(output, outputReponse)
	if err != nil {
		return nil, err
	}

	return outputReponse, nil
}

func (c *OpenRouterClient) FetchChatCompletionsStream(request Request, outputChan chan Response, processingChan chan interface{}, errChan chan error, ctx context.Context) {
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
		"Content-Type":  "application/json",
	}

	if request.Provider != nil {
		headers["HTTP-Referer"] = request.Provider.RefererURL
		headers["X-Title"] = request.Provider.SiteName
	}

	body, err := json.Marshal(request)
	if err != nil {
		errChan <- err
		close(errChan)
		close(outputChan)
		close(processingChan)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.apiURL), bytes.NewBuffer(body))
	if err != nil {
		errChan <- err
		close(errChan)
		close(outputChan)
		close(processingChan)
		return
	}

	req = req.WithContext(ctx)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		errChan <- err
		close(errChan)
		close(outputChan)
		close(processingChan)
		return
	}

	go func() {
		defer resp.Body.Close()

		defer close(errChan)
		defer close(outputChan)
		defer close(processingChan)
		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				line, err := reader.ReadString('\n')
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, ":") {
					select {
					case processingChan <- true:
					case <-ctx.Done():
						errChan <- ctx.Err()
						return
					}
					continue
				}

				if line != "" {
					if !strings.HasPrefix(line, "data:") || len(line) < len("data:") {
						errChan <- fmt.Errorf("unexpected response line: %q", line)
						return
					}
					payload := strings.TrimSpace(line[len("data:"):])
					if payload == "[DONE]" {
						return
					}
					response := Response{}
					err = json.Unmarshal([]byte(payload), &response)
					if err != nil {
						errChan <- err
						return
					}
					select {
					case outputChan <- response:
					case <-ctx.Done():
						errChan <- ctx.Err()
						return
					}
				}

				if err != nil {
					if err == io.EOF {
						return
					}
					errChan <- err
					return
				}
			}
		}
	}()
}
