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

type StreamEvent struct {
	Response   *Response
	Processing bool
	Err        error
}

// ChatCompletionsStream wraps streaming channels into a pull-based iterator.
type ChatCompletionsStream struct {
	output     <-chan Response
	processing <-chan interface{}
	errs       <-chan error
	done       bool
}

// Recv returns the next stream event. ok=false means the stream ended.
// If an error is returned in the event, the stream is finished.
func (s *ChatCompletionsStream) Recv(ctx context.Context) (StreamEvent, bool) {
	if s.done {
		return StreamEvent{}, false
	}

	for {
		if s.output == nil && s.processing == nil && s.errs == nil {
			s.done = true
			return StreamEvent{}, false
		}

		select {
		case resp, ok := <-s.output:
			if !ok {
				s.output = nil
				continue
			}
			return StreamEvent{Response: &resp}, true
		case _, ok := <-s.processing:
			if !ok {
				s.processing = nil
				continue
			}
			return StreamEvent{Processing: true}, true
		case err, ok := <-s.errs:
			if !ok {
				s.errs = nil
				continue
			}
			s.errs = nil
			s.done = true
			if err != nil {
				return StreamEvent{Err: err}, false
			}
			return StreamEvent{}, false
		case <-ctx.Done():
			s.done = true
			return StreamEvent{Err: ctx.Err()}, false
		}
	}
}

// StartChatCompletionsStream starts a streaming request and returns an iterator-style wrapper.
func (c *OpenRouterClient) StartChatCompletionsStream(request Request, ctx context.Context) *ChatCompletionsStream {
	outputChan := make(chan Response)
	processingChan := make(chan interface{})
	errChan := make(chan error)
	c.FetchChatCompletionsStream(request, outputChan, processingChan, errChan, ctx)
	return &ChatCompletionsStream{
		output:     outputChan,
		processing: processingChan,
		errs:       errChan,
	}
}
