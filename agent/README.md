# Agent Namespace

The agent namespace provides a root element that replaces `<scxml>` with support for datamodel specification and dynamic namespace declarations.

## Namespace URI

```
github.com/agentflare-ai/agentml/agent
```

## Elements

### `<agent>`

Root element that replaces `<scxml>` in agent documents. Provides namespace declarations and datamodel specification.

#### Attributes

- **`datamodel`** (optional): Specifies the datamodel implementation to use for this agent
  - Example: `datamodel="ecmascript"`, `datamodel="xpath"`

- **`use:*`** (optional): Universal loader for both namespace implementations and external specifications
  - Format: `use:prefix="namespace-uri-or-spec-path"`
  - Multiple `use:*` attributes can be specified
  - The loader automatically detects whether the value is:
    - A namespace URI (e.g., `github.com/...`) - loads custom namespace implementation
    - A spec file (e.g., `./api.json`, `https://...`) - loads OpenAPI/JSON Schema

  **Namespace Examples:**
  - `use:memory="github.com/agentflare-ai/agentml/memory"` - loads memory namespace
  - `use:gemini="github.com/agentflare-ai/agentml/gemini"` - loads Gemini namespace

  **Spec Examples (OpenAPI, JSON Schema, etc.):**
  - `use:api="./api-spec.json"` - loads local spec, referenced as "api"
  - `use:events="https://api.example.com/events.json"` - loads remote spec, referenced as "events"
  - `use:models="./schemas/models.json"` - loads another spec, referenced as "models"

  Supported spec formats: OpenAPI 3.x, Swagger 2.0, JSON Schema (auto-detected)

### `event:schema` Attribute on `<transition>`

The agent namespace also supports an `event:schema` attribute on `<transition>` elements to specify a JSON schema for event payload validation.

- **`event:schema`**: JSON schema string or JSON Pointer reference for validating the event payload
  - Applied to `<transition>` elements
  - **Inline format**: `event:schema='{"type": "object", "properties": {...}}'`
  - **Namespace-prefixed format**: `event:schema="api:#/components/schemas/UserInput"` (references `use:api` spec)
  - **Namespace-prefixed format**: `event:schema="events:#/definitions/WebhookEvent"` (references `use:events` spec)
  - **Default format**: `event:schema="#/components/schemas/MySchema"` (uses first loaded spec)
  - Supports standard JSON Pointer syntax (RFC 6901)

  **Supported JSON Pointer paths:**
  - OpenAPI 3.x: `#/components/schemas/SchemaName`
  - Swagger 2.0: `#/definitions/SchemaName`
  - JSON Schema: `#/definitions/SchemaName` or `#/$defs/SchemaName`

#### Child Elements

The `<agent>` element contains standard SCXML content (states, transitions, data, etc.) as direct children.

## Usage Example

```xml
<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       use:api="./schemas/api-spec.json"
       use:events="https://api.example.com/events.json"
       use:memory="github.com/agentflare-ai/agentml/memory"
       use:gemini="github.com/agentflare-ai/agentml/gemini">

  <!-- Standard SCXML content as direct children -->
  <state id="initial">
    <onentry>
      <memory:put key="greeting" value="Hello, World!" />
      <gemini:generate model="gemini-pro" location="response">
        <prompt>Analyze this greeting</prompt>
      </gemini:generate>
    </onentry>

    <!-- Transition with inline event schema -->
    <transition event="user.input"
                event:schema='{"type": "object", "properties": {"text": {"type": "string"}}, "required": ["text"]}'
                target="processing" />

    <!-- Transition with namespace-prefixed JSON Pointer reference -->
    <transition event="user.message"
                event:schema="api:#/components/schemas/UserMessage"
                target="processing" />

    <!-- Transition with reference to different spec -->
    <transition event="webhook.received"
                event:schema="events:#/definitions/WebhookEvent"
                target="processing" />
  </state>

  <state id="processing">
    <transition event="done" target="finished" />
  </state>

  <final id="finished" />

</agent>
```

## Loading the Namespace

To use this namespace in your Go code:

```go
package main

import (
    "context"
    "github.com/agentflare-ai/agentml/agent"
    "github.com/agentflare-ai/agentml"
)

func main() {
    // Get the namespace loader
    loader := agent.Loader()

    // Load the namespace (typically done by the interpreter)
    ns, err := loader(ctx, interpreter, document)
    if err != nil {
        // handle error
    }

    // The namespace is now available for handling agent elements
    _ = ns
}
```

## XSD Schema

The agent namespace includes an XSD schema file (`agent.xsd`) that defines the structure of the `<agent>` element for validation and IDE support.

## Features

- **Replaces `<scxml>` root element**: Use `<agent>` as the document root instead of `<scxml>`
- **Datamodel Declaration**: Explicitly specify which datamodel implementation to use
- **Universal Loader (`use:*`)**: Unified pattern for loading both namespaces and specifications
  - Auto-detects namespace URIs vs. spec files
  - Loads custom namespace implementations (e.g., `use:memory="github.com/..."`)
  - Loads OpenAPI/JSON Schema specs (e.g., `use:api="./api.json"`)
  - Supports both local files and remote URLs
- **Event Schema Validation**: Use `event:schema` attribute on `<transition>` elements with:
  - Inline JSON schemas
  - Namespace-prefixed JSON Pointer references (e.g., `api:#/components/schemas/User`)
  - Default JSON Pointer references (uses first loaded spec)
- **Multiple Spec Support**: Load multiple OpenAPI/JSON Schema specifications with different namespace prefixes
- **Format Auto-Detection**: Automatically detects OpenAPI 3.x, Swagger 2.0, and JSON Schema formats
- **JSON Pointer Support**: Reference schemas using RFC 6901 JSON Pointers
- **Repository Integration**: Each namespace `use:*` declaration references a repository that contains:
  - An XSD schema file for validation
  - A namespace loader implementation
- **Transparent Execution**: Child elements (states, transitions, etc.) are executed normally by the interpreter

## Implementation Details

The `<agent>` root element:

1. Parses the `datamodel` attribute to determine which datamodel to use
2. Parses all `use:*` attributes to collect namespace declarations
3. Recursively processes all `<transition>` elements to collect `event:schema` attributes
4. Logs the configuration for debugging
5. Executes all child elements (SCXML states, transitions, etc.) in order using the interpreter

This allows for flexible agent definitions where the required namespaces and datamodel can be explicitly declared at the document root, replacing the traditional `<scxml>` element.
