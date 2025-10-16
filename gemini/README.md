# Gemini

A Go client library for Google's Gemini AI models with advanced rate limiting, model management, and OpenTelemetry integration.

## Features

* **Multi-model Support**: Manage multiple Gemini models with different configurations
* **Advanced Rate Limiting**: Built-in rate limiting with multiple tiers and backoff strategies
* **OpenTelemetry Integration**: Full observability with tracing and metrics
* **Context-aware**: Proper context propagation throughout the API
* **Type-safe Model Management**: Strongly typed model names and configurations

## Installation

```bash
go get github.com/agentflare-ai/agentml/gemini
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/agentflare-ai/agentml/gemini"
    "google.golang.org/genai"
)

func main() {
    ctx := context.Background()
    
    // Define your models
    models := map[gemini.ModelName]*gemini.Model{
        "gemini-pro": {
            Name: "gemini-pro",
            // Configure model parameters
        },
    }
    
    // Create client
    client, err := gemini.NewClient(ctx, models, &genai.ClientConfig{
        // Your API configuration
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Generate content
    response, err := client.GenerateContent(ctx, "gemini-pro", 
        []*genai.Content{{
            Parts: []genai.Part{genai.Text("Hello, world!")},
        }}, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use response...
}
```

## Rate Limiting

The client includes sophisticated rate limiting with multiple tiers:

* **Tier-based Limits**: Different rate limits for different usage patterns
* **Adaptive Backoff**: Intelligent backoff strategies when limits are hit
* **Request Prioritization**: Priority queuing for different request types

## Model Management

Models are managed through a strongly-typed system:

```go
type Model struct {
    Name        ModelName
    RateLimit   *RateLimit
    Config      *ModelConfig
    // Additional model-specific settings
}
```

## Observability

Full OpenTelemetry support for monitoring and debugging:

* Request/response tracing
* Rate limiting metrics
* Error tracking
* Performance monitoring

## License

This project is part of the agentml ecosystem.
