package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2aclient/agentcard"
	"github.com/dominicnunez/agentos-a2a-go/executionkind"
)

func main() {
	baseURL := strings.TrimSpace(os.Getenv("AGENTOS_A2A_URL"))
	token := strings.TrimSpace(os.Getenv("AGENTOS_A2A_TOKEN"))
	if baseURL == "" || token == "" {
		log.Fatal("AGENTOS_A2A_URL and AGENTOS_A2A_TOKEN are required")
	}
	origin, err := parseOrigin(baseURL)
	if err != nil {
		log.Fatalf("invalid AGENTOS_A2A_URL: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	discoveryClient := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("redirects are disabled")
		},
	}
	card, err := agentcard.NewResolver(discoveryClient).Resolve(ctx, baseURL)
	if err != nil {
		log.Fatalf("resolve Agent Card: %v", err)
	}
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("redirects are disabled")
		},
		Transport: bearerTransport{origin: origin, token: token, base: http.DefaultTransport},
	}
	client, err := a2aclient.NewFromCard(ctx, card, a2aclient.WithJSONRPCTransport(httpClient))
	if err != nil {
		log.Fatalf("create A2A client: %v", err)
	}
	message := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart("echo hello"))
	if err := executionkind.Set(message, executionkind.KindDeterministic); err != nil {
		log.Fatalf("set execution kind: %v", err)
	}
	result, err := client.SendMessage(ctx, &a2a.SendMessageRequest{Message: message})
	if err != nil {
		log.Fatalf("send message: %v", err)
	}
	summary, err := summarizeResult(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(summary)
}

func summarizeResult(result a2a.SendMessageResult) (string, error) {
	switch value := result.(type) {
	case *a2a.Task:
		if value == nil {
			return "", errors.New("received a nil task response")
		}
		return fmt.Sprintf("task %s: %s", value.ID, value.Status.State), nil
	case *a2a.Message:
		if value == nil {
			return "", errors.New("received a nil message response")
		}
		return fmt.Sprintf("message %s received", value.ID), nil
	default:
		return "", fmt.Errorf("unexpected response type %T", result)
	}
}

type bearerTransport struct {
	origin *url.URL
	token  string
	base   http.RoundTripper
}

func (t bearerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != t.origin.Scheme || request.URL.Host != t.origin.Host {
		return nil, errors.New("refusing to send credential to a different origin")
	}
	clone := request.Clone(request.Context())
	clone.Header = request.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(clone)
}

func parseOrigin(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("URL must have a host and no user information, query, or fragment")
	}
	if parsed.Scheme != "https" {
		hostname := parsed.Hostname()
		if parsed.Scheme != "http" || (hostname != "localhost" && hostname != "127.0.0.1" && hostname != "::1") {
			return nil, errors.New("URL must use HTTPS except for a loopback HTTP endpoint")
		}
	}
	return &url.URL{Scheme: parsed.Scheme, Host: parsed.Host}, nil
}
