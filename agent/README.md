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

- **`use:*`** (optional): Namespace declarations (equivalent to `xmlns:*`) that also load the namespace loader
  - Format: `use:prefix="namespace-uri"`
  - Multiple `use:*` attributes can be specified
  - Example: `use:memory="github.com/agentflare-ai/agentml/memory"` is equivalent to `xmlns:memory="..."`
  - Example: `use:gemini="github.com/agentflare-ai/agentml/gemini"`

### `event:schema` Attribute on `<transition>`

The agent namespace also supports an `event:schema` attribute on `<transition>` elements to specify a JSON schema for event payload validation.

- **`event:schema`**: JSON schema string for validating the event payload
  - Applied to `<transition>` elements
  - Format: `event:schema='{"type": "object", "properties": {...}}'`
  - Used to validate event data matches the expected schema

#### Child Elements

The `<agent>` element contains standard SCXML content (states, transitions, data, etc.) as direct children.

## Usage Example

```xml
<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
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

    <!-- Transition with event schema validation -->
    <transition event="user.input"
                event:schema='{"type": "object", "properties": {"text": {"type": "string"}}, "required": ["text"]}'
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
- **Namespace Loading**: Use `use:*` attributes (equivalent to `xmlns:*`) to declare and load namespaces
- **Event Schema Validation**: Use `event:schema` attribute on `<transition>` elements to specify JSON schemas for event payloads
- **Repository Integration**: Each `use:*` declaration references a repository that contains:
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
