# AgentML

> **🚧 Early Alpha - Building in Public**
> 
> AgentML is in early alpha and being built openly with the community. The vision is ambitious, the foundation is solid, but many features are still in development. Join us in shaping the future of agent standards.
>
> **⚠️ Repository Status:** Many packages in this repository are proof-of-concept (POC) or work-in-progress (WIP). Code quality, APIs, and implementation approaches are evolving rapidly. Use with caution in production environments.

---

## The Vision: A Universal Language for AI Agents

The AI agent landscape is fragmented and accelerating, with new frameworks appearing weekly. This creates vendor lock-in and forces costly rewrites when a chosen framework becomes limiting or unmaintained.

**AgentML is the universal language for agents, inspired by the success of HTML for the web.**

```
AgentML : Agent Frameworks  =  HTML : Web Browsers
```

Just as HTML lets you write content once and have it render in any browser, AgentML lets you define your agent's behavior once and run it anywhere. This is achieved by building on the battle-tested [W3C SCXML standard](https://www.w3.org/TR/scxml/), a formal model for state machines that has been proven for over 20 years in complex industrial systems.

This provides two primary paths for execution:

1.  **Native Execution (Recommended)**: Run agents with **`agentmlx`**, the reference runtime built in Go/WASM. It's designed for high performance and portability.
2.  **Transformation (Planned)**: To integrate with existing ecosystems, we are planning transformers to convert AgentML into other popular frameworks like LangGraph, CrewAI, n8n, and more. **This feature is not yet implemented but is a key part of our roadmap.**

By separating behavior from runtime, your agents outlive framework trends.

---
## Table of Contents

- [Installation](#installation)
- [Core Concepts](#core-concepts)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Write Once, Deploy Anywhere (Planned)](#write-once-deploy-anywhere-)
- [Extensibility with WebAssembly (Planned)](#extensibility-with-webassembly-planned)
- [Remote Agent Communication](#remote-agent-communication)
- [Current Status & Roadmap](#current-status--roadmap)
- [Getting Started](#getting-started)
- [Namespaces](#namespaces)
- [Best Practices](#best-practices)

---

## Installation

### Installing agentmlx Runtime

To run AgentML files (`.aml`), you need the `agentmlx` runtime. Install it with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/agentflare-ai/agentml/main/install.sh | sh
```

This will:
- Automatically detect your platform (Linux/macOS, amd64/arm64)
- Download the latest release
- Verify checksums for security
- Install to `~/.agentmlx/bin`
- Add to your PATH

**Install from different channels:**
```bash
# Latest stable release
curl -fsSL https://raw.githubusercontent.com/agentflare-ai/agentml/main/install.sh | sh

# Next (release candidate)
curl -fsSL https://raw.githubusercontent.com/agentflare-ai/agentml/main/install.sh | sh -s -- --channel next

# Beta releases
curl -fsSL https://raw.githubusercontent.com/agentflare-ai/agentml/main/install.sh | sh -s -- --channel beta
```

**Install specific version:**
```bash
curl -fsSL https://raw.githubusercontent.com/agentflare-ai/agentml/main/install.sh | sh -s -- --version 1.0.0-rc.1
```

**Install to custom directory:**
```bash
export AGENTMLX_INSTALL_DIR=/usr/local
curl -fsSL https://raw.githubusercontent.com/agentflare-ai/agentml/main/install.sh | sh
```

**Release Channels:**
- `latest` - Stable releases (v1.0.0) - **Default**
- `next` - Release candidates (v1.0.0-rc.1)
- `beta` - Beta releases (v1.0.0-beta.1)

The installer automatically falls back if a channel is empty: `latest` → `next` → `beta`

**Verify installation:**
```bash
agentmlx --version
# or use the shorter alias
amlx --version
```

Both `agentmlx` and `amlx` commands are available after installation.

For more installation options and manual downloads, see the [agentmlx documentation](https://github.com/agentflare-ai/agentmlx#installation).

---

## Core Concepts

AgentML uses SCXML state machines to define **deterministic behavior**, moving beyond prompt-only approaches where behavior is emergent and unpredictable.

1.  **State Machines**: Explicitly define valid states, transitions, and the agent's lifecycle. This enables formal verification and testing.

2.  **Schema-Guided Events**: LLM outputs are constrained to structured JSON events validated against schemas using the `event:schema` attribute. This ensures reliability and type safety.

    > ⚠️ **Work in Progress**: Event schema validation and external schema loading (`import` directive) are in active development. APIs and features may change as we refine the implementation based on community feedback.

    The `event:schema` attribute on a transition provides JSON schema validation for an event. **It is critical to include `description` fields** at both the schema and property level, as these descriptions are the primary way to guide LLMs in generating correct event data.

    **With Descriptions (Good):**
```json
{
        "type": "object",
      "description": "User intent to perform a flight-related action (search, book, update, cancel)",
        "properties": {
          "action": {
          "type": "string",
            "enum": ["search", "book", "update", "cancel"],
          "description": "The specific action: search for flights, book a new flight, etc."
          },
          "details": {
            "type": "object",
          "description": "Flight-specific information extracted from user message",
            "properties": {
              "from": {
                "type": "string",
              "description": "Departure location: city name or airport code (e.g., 'New York' or 'JFK')"
            }
          }
        }
  }
}
```

3.  **Efficient Token Usage**: The runtime provides the LLM with a "snapshot" of the current state, datamodel, and available events. This context allows prompts to be minimal, and the static parts (the agent's SCXML definition) can be cached by the LLM provider, reducing token consumption.

4.  **Decomposition**: Complex agents can be broken down into smaller, reusable state machines using the `<invoke>` tag. This is ideal for managing complexity and sharing components like authentication or payment processing.

5.  **Compiler-Inspired Validation**: To ensure reliability, especially when agents are building other agents, AgentML includes a powerful validation system. Inspired by the Rust compiler (`rustc`), it provides detailed, actionable error messages that help developers (and other agents) pinpoint issues quickly and achieve a high success rate when generating AgentML documents.

    Here is an example of the validator's output:
    ```bash
    ./agentml/examples/customer_support/customer_support.aml:89:5: WARNING[W340] State 'await_user_input' has only conditional transitions and may deadlock if no events match
          88 |     <!-- Await User Input and Classify Intent (Combined State) -->
          89 |     <state id="await_user_input">
                   ^
          90 |       <onentry>
      hint: Add an unconditional fallback transition (without 'event' or 'cond' attributes)
      hint: Or ensure all possible events are handled
      hint: Example: <transition target="fallback_state" />

    summary: 0 error(s), 1 warning(s), 1 total
    ```

### Schema References with `import`

To keep agent files clean and promote reuse, schemas can be defined in external JSON or YAML files (including OpenAPI specs) and loaded with an `import` directive. The runtime intelligently detects the file type.

This enables schema reuse via **JSON Pointer (RFC 6901)** references with namespace prefixes.

**schemas/events.json:**
```json
{
  "components": {
    "schemas": {
      "FlightRequest": {
  "type": "object",
        "description": "Schema for a flight-related request.",
  "properties": {
            "action": { "$ref": "#/components/schemas/FlightAction" }
        }
      },
      "FlightAction": {
      "type": "string",
          "enum": ["search", "book", "cancel"]
      }
    }
  }
}
```

**agent.aml:**
```xml
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       import:events="./schemas/events.json">

  <!-- Reference schemas using a namespace and JSON Pointer -->
  <transition event="intent.flight"
              event:schema="events:#/components/schemas/FlightRequest"
              target="handle_flight" />
</agent>
```

This unified `import` directive is designed to work for schemas, namespace implementations, and future WASM components.

---
## Key Features

- 🎯 **Deterministic Behavior**: Predictable, auditable agent behavior via state machines.
- 📝 **Schema-Guided Events**: `event:schema` attributes validate LLM-generated events. 🚧
- 🔄 **Runtime Snapshots**: Efficiently provide LLM context, minimizing token usage. ✅
- 📦 **Modular Design**: Decompose complex agents into reusable components with `<invoke>`. ✅
- 🔌 **Extensible Namespaces**: Plug in custom functionality (LLMs, memory, I/O). ✅
- 📊 **Observable**: Foundation for OpenTelemetry tracing and logging. ✅
- 🌐 **Universal Standard**: Write once, deploy anywhere via native runtime or transformation. 🔮
- 🔗 **Remote Communication**: Built-in distributed agent communication via IOProcessors. ✅

**Legend:** ✅ Working | 🚧 In Development | 🔮 Planned

---

## Architecture

- **Document Structure**: AgentML files use an `<agent>` root element, which is a compatible extension of SCXML's `<scxml>` element. The `datamodel` attribute specifies the scripting language used for data manipulation and expressions.
- **Supported Datamodels**: AgentML supports `ecmascript`, `starlark`, and `xpath`. Support for using `wasm` components as a datamodel is planned for the future.
- **Namespace System**: Functionality is extended through namespaces (e.g., for Gemini, Ollama, Memory) declared with the `import:prefix="uri"` directive.
- **Runtime Snapshot**: At each step, the runtime creates an XML snapshot containing the active states, datamodel, and available events. This, combined with the SCXML document, gives the LLM complete and current context.

---

## Write Once, Deploy Anywhere 🔮

> **🚧 Vision Statement**: This section describes our goal for AgentML. Framework transformers are planned and not yet available.

The core promise of AgentML is to end the cycle of constant rewrites caused by framework fragmentation.

```
        AgentML (.aml)
               |
       ┌───────┴──────────────────┐
       ▼                          ▼
   agentmlx                   Transform to: (Planned)
  (Primary                    - LangGraph
   Runtime)                   - CrewAI
       |                      - n8n
       ▼                      - ...and more
   Go/WASM
   Anywhere
```

#### Native Runtime: `agentmlx`

The **`agentmlx`** runtime is the recommended way to execute AgentML files. It is a high-performance, portable Go/WASM binary that is fully compliant with the W3C SCXML specification, passing all 193 official conformance tests. `agentmlx` will be open-sourced soon.

```bash
# Future: Run directly with agentmlx (or amlx)
agentmlx run customer-support.aml
# or shorter
amlx run customer-support.aml
```

#### Framework Transformers (Planned)

When you need to integrate with an existing ecosystem, transformers will convert AgentML into framework-specific code.

```bash
# PLANNED: Transform AgentML to LangGraph
agentmlx transform customer-support.aml --target langgraph --output customer-support.py
```

This provides framework insurance, eliminating vendor lock-in and allowing you to choose a runtime based on deployment needs, not sunk costs.

---

## Extensibility with WebAssembly (Planned) 🔮

> **🚧 Vision Statement**: WASM-based namespaces are a forward-looking goal.

Our vision for true interoperability and extensibility is to load namespaces as **WebAssembly (WASM) components** that adhere to the `agentml.wit` interface.

```xml
<!-- Future: Load namespace from a WASM module -->
<agent import:gemini="https://cdn.example.com/gemini-namespace.wasm"
       import:custom="./my-namespace.wasm">
  
  <gemini:generate ... />
  <custom:process ... />
</agent>
```

This means you can:
- **Write extensions in any language** (Rust, Go, Python, C++) that compiles to WASM.
- **Run on any runtime** that supports the WASM component model.
- **Securely sandbox** custom code.

This is the "JavaScript for agents," enabling a dynamic and polyglot ecosystem on top of AgentML's "HTML for agents" structure.

---

## Remote Agent Communication

AgentML supports distributed agent communication using the W3C SCXML `IOProcessor` interface. Agents can communicate across processes and networks using standard protocols like HTTP and WebSockets.

```xml
<state id="notify_remote_agent">
  <onentry>
    <!-- Send an event to a remote agent via HTTP -->
    <send event="task.assigned"
          target="https://agent.example.com/events"
          type="github.com/agentflare-ai/agentml/ioprocessor/http">
      <param name="task_id" expr="task.id" />
    </send>
  </onentry>
  
  <!-- Wait for a response -->
  <transition event="task.acknowledged" target="confirmed" />
</state>
```
This architecture supports patterns like agent swarms, supervisor-worker delegation, and pub/sub, with built-in support for security, observability, and automatic trace propagation.

---

## Current Status & Roadmap

We are building AgentML in the open. Your feedback is critical.

**What's working now:**
- ✅ Core SCXML interpreter (Go implementation)
- ✅ Gemini & Ollama LLM integration namespaces
- ✅ Basic event-driven agent workflows
- ✅ Datamodel and state machine semantics
- ✅ OpenTelemetry tracing foundation
- ✅ IOProcessor implementations (HTTP, WebSocket)

**What's in active development:**
- 🚧 **`agentmlx` runtime** - Native Go/WASM execution (primary focus)
- 🚧 Memory namespace (vector search, graph database)
- 🚧 Event schema validation (`event:schema`) and external schema loading (`import`)

**What's planned:**
- 🔮 **Framework transformers** (LangGraph, CrewAI, n8n, OpenAI, Autogen)
- 🔮 **WASM namespace loading** via `agentml.wit`
- 🔮 Visual editor and debugger
- 🔮 Agent marketplace

**How to participate:**
- **🗣️ Share your use cases** in [GitHub Discussions](https://github.com/agentflare-ai/agentml/discussions).
- **💡 Propose features** via [GitHub Issues](https://github.com/agentflare-ai/agentml/issues).
- **🔧 Contribute code** through pull requests.

---

## Getting Started

> **⚠️ Alpha Software**: APIs may change, features may be incomplete, and you may encounter bugs.

### What You Need

**AgentML is a language specification** - you write `.aml` files that define your agent's behavior. The `agentmlx` runtime (coming soon) executes these files. No installation of AgentML itself is needed.

**To use AgentML:**
1. Write your agent in `.aml` files (see example below)
2. Run with `agentmlx run your-agent.aml` or `amlx run your-agent.aml` (runtime in development)
3. The runtime handles all namespaces, extensions, and execution

### Basic Agent Example

```xml
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       import:gemini="github.com/agentflare-ai/agentml/gemini">

  <datamodel>
    <data id="user_input" expr="''" />
    <data id="response" expr="''" />
  </datamodel>

  <state id="main">
    <state id="awaiting_input">
      <onentry>
        <!-- In a real agent, input comes from an IOProcessor -->
        <assign location="user_input" expr="getUserInput()" />
      </onentry>
      <transition target="processing" />
    </state>

    <state id="processing">
      <onentry>
        <!-- LLM generates a structured event based on the input -->
        <gemini:generate
          model="gemini-2.0-flash-exp"
          location="_event"
          promptexpr="'Process this input: ' + user_input" />
      </onentry>
      
      <!-- Transition only if the LLM output matches the event schema -->
      <transition event="action.response"
                  event:schema='{"type": "object", "properties": {"message": {"type": "string"}}, "required": ["message"]}'
                  target="responding" />
    </state>

    <state id="responding">
      <onentry>
        <assign location="response" expr="_event.data.message" />
        <log expr="'Response: ' + response" />
      </onentry>
      <transition target="awaiting_input" />
    </state>
  </state>
</agent>
```



**For most users:** Just write `.aml` files and run them with `agentmlx` (or `amlx`) - no Go code needed!

---

## Namespaces

AgentML's functionality is extended through namespaces. Here are the currently available or planned ones:

- **Agent (`.../agentml/agent`)**: Core namespace for `<agent>` root element and `event:schema` validation.
- **Gemini (`.../agentml/gemini`)**: Google Gemini LLM integration.
- **Ollama (`.../agentml/ollama`)**: Local LLM integration via Ollama.
- **Memory (`github.com/agentflare-ai/agentml/memory`)**: High-performance memory with vector search and graph database capabilities. This is powered by `sqlite-graph`, our custom extension that provides a complete, local, filesystem-based memory framework within a single SQLite file.

```xml
<agent import:memory="github.com/agentflare-ai/agentml/memory">
  <!-- Vector operations -->
  <memory:embed location="embedding" expr="text_content" />
  <memory:search location="results" expr="query_embedding" limit="10" />
  
  <!-- Graph operations -->
  <memory:graph-query location="results">
    <query>
      MATCH (p:Person)-[:KNOWS]->(friend)
      WHERE p.age > 25
      RETURN p.name, friend.name
    </query>
  </memory:graph-query>
  
  <!-- Key-value storage -->
  <memory:put key="user_preference" expr="preference_value" />
  <memory:get key="user_preference" location="preference" />
</agent>
```

Features:
- Vector similarity search
- Graph database with Cypher queries
- Embedding generation
- Persistent key-value storage

See [memory/README.md](./memory/README.md) for details.
- **Stdin (`.../agentml/stdin`)**: Simple stdin/stdout I/O for console agents.

### Creating Custom Namespaces

Custom namespaces can be implemented in Go (currently), or any language that compiles to WASM (future).

**Current:** Go implementations define types and interfaces in code.

**Future:** WebAssembly Component Model via [`agentml.wit`](./agentml.wit) 🔮

The `agentml.wit` file defines standard interfaces for namespaces using WebAssembly Interface Types (WIT). This enables:
- **Language freedom**: Implement namespaces in Rust, Go, Python, C++, or any WASM-capable language
- **Portable**: Same `.wasm` module works across all runtimes
- **Standard contract**: Defined interfaces ensure interoperability
- **Secure**: WASM sandboxing isolates namespace code

> **🚧 Migration Note:** Current Go-based type definitions (e.g., `types.go`) will be deprecated in favor of `agentml.wit` as the canonical type specification. This transition enables true polyglot namespace development.

---

## Best Practices

- **Keep `.aml` files focused**: Decompose large agents into smaller, invoked services.
- **Use meaningful state IDs**: `handle_flight_request` is better than `state_5`.
- **Validate with schemas**: Always use `event:schema` and provide detailed `description` fields to guide the LLM.
- **Use external schemas**: Define schemas in `.json`/`.yaml` files and load them with `import:` for reuse and maintainability.
- **Prefer external scripts**: Use `<script src="./utils.js" />` for better linting, IDE support, and maintainability. Only use inline scripts for simple expressions. When you must write inline scripts with comparison operators (`<`, `>`) or other special XML characters, wrap your code in `<![CDATA[...]]>`:

```xml
  <!-- Best: External script with full linting support -->
  <script src="./validation.js" />
  
  <!-- Inline without CDATA: XML parser errors -->
    <script>
    if (count < 10 && value > 5) {  <!-- This will break! -->
      return true;
    }
    </script>
  
  <!-- Inline with CDATA: Works but no linting -->
    <script>
      <![CDATA[
    if (count < 10 && value > 5) {
      return true;
    }
      ]]>
    </script>
  ```
