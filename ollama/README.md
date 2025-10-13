# Ollama

A Go client library for Ollama models with SCXML integration.

## Features

* **Model Support**: Support for various Ollama models (Llama, Mistral, CodeLlama, etc.)
* **SCXML Integration**: Native SCXML executable content for state machine-based AI workflows
* **Template Processing**: Go template support for dynamic prompt generation
* **OpenTelemetry Integration**: Full observability with tracing
* **Context-aware**: Proper context propagation throughout the API

## Installation

```bash
go get github.com/gogo-agent/ollama
```

## Prerequisites

This package communicates directly with the Ollama REST API. Make sure you have Ollama running locally:

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama server
ollama serve

# Pull a model (in another terminal)
ollama pull llama3.2
```

The package connects to `http://localhost:11434` by default, which is Ollama's standard API endpoint.

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/gogo-agent/ollama"
)

func main() {
    ctx := context.Background()

    // Define your models
    models := map[ollama.ModelName]*ollama.Model{
        "llama3.2": ollama.NewModel(ollama.Llama3_2, false),
    }

    // Create client
    client, err := ollama.NewClient(ctx, models, &ollama.ClientOptions{
        BaseURL: "http://localhost:11434", // Default Ollama server
    })
    if err != nil {
        log.Fatal(err)
    }

    // Generate content
    response, err := client.Generate(ctx, "llama3.2", "Hello, world!", false)
    if err != nil {
        log.Fatal(err)
    }

    // Use response...
    println(response)
}
```

## SCXML Usage

The Ollama package provides SCXML executable content for integrating AI generation into state machines:

```xml
<scxml xmlns="http://www.w3.org/2005/07/scxml"
       xmlns:ollama="http://www.gogo-agent.dev/ollama"
       version="1.0"
       datamodel="ecmascript"
       initial="start">

  <datamodel>
    <data id="greeting" expr="''"/>
  </datamodel>

  <state id="start">
    <onentry>
      <ollama:generate model="llama3.2"
                       prompt="Say hello in a friendly way"
                       location="greeting" />
    </onentry>
    <transition event="content.generated" target="end"/>
  </state>

  <final id="end"/>
</scxml>
```

## Template Support

Use child `<ollama:prompt>` elements with Go templates:

```xml
<ollama:generate model="llama3.2" location="result">
  <ollama:prompt>
    Hello {{.name}}, you are {{.age}} years old.
  </ollama:prompt>
</ollama:generate>
```

## Models

Common Ollama models supported:

* `llama3.2` - Meta Llama 3.2
* `llama3.1` - Meta Llama 3.1
* `codellama` - Code Llama
* `mistral` - Mistral 7B

## Configuration

```go
// Client options
options := &ollama.ClientOptions{
    BaseURL: "http://localhost:11434", // Ollama server URL
}

// Model configuration
model := ollama.NewModel(ollama.Llama3_2, true) // Enable streaming
```

## Observability

Full OpenTelemetry support for monitoring and debugging:

* Request/response tracing
* Error tracking
* Performance monitoring

## License

This project is part of the gogo-agent ecosystem.