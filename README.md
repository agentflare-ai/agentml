# AgentML

> **🚧 Early Alpha - Building in Public**
> 
> AgentML is in early alpha and being built openly with the community. The vision is ambitious, the foundation is solid, but many features are still in development. Join us in shaping the future of agent standards.

## What We're Building

**AgentML** is an agentic workflow framework built on [W3C SCXML](https://www.w3.org/TR/scxml/) that aims to provide deterministic, event-driven state machines for building LLM-powered agents. By combining SCXML's formal state machine semantics with modern LLM capabilities, AgentML will enable highly composable, maintainable, and predictable agent behaviors.

**Our Vision - The Universal Agent Standard**: Define your agent once in AgentML and deploy anywhere. Transform to LangGraph, CrewAI, n8n, OpenAI Agent Builder, or run natively with the agentmlx WASM runtime. AgentML aims to unify the fractured agentic ecosystem by establishing W3C SCXML as the standard language for agents.

---

## 🌱 Current Status & Building in Public

**Where we are today:** We have a working Go implementation of the SCXML interpreter with support for the Gemini LLM namespace and basic event-driven agent workflows. The core concepts work, and we're actively developing the ecosystem around it.

**What's working now:**
- ✅ Core SCXML interpreter (Go implementation)
- ✅ Gemini LLM integration namespace
- ✅ Event-driven agent workflows
- ✅ Basic state machine semantics
- ✅ OpenTelemetry tracing foundation

**What's in active development:**
- 🚧 Event schema validation (`event:schema` attribute)
- 🚧 External schema loading (`use:*` pattern for JSON/OpenAPI specs)
- 🚧 Framework transformers (LangGraph, CrewAI, n8n, etc.)
- 🚧 Memory namespace (vector search, graph database)
- 🚧 agentmlx WASM runtime
- 🚧 IOProcessor implementations (HTTP, WebSocket)

**What's planned:**
- 🔮 Complete transformer ecosystem
- 🔮 Visual editor
- 🔮 Agent marketplace
- 🔮 Additional LLM provider namespaces

### Why Build in Public?

**We need your help.** The vision of a universal agent standard can only succeed if it serves the real needs of the community. We're building this openly because:

1. **Community-driven design**: Your use cases and pain points should shape the standard
2. **Transparency**: See exactly what works, what doesn't, and what's coming
3. **Collaboration**: Build together rather than in isolation
4. **Avoid mistakes**: Learn from the community before APIs are set in stone
5. **Real-world testing**: Get feedback on actual implementations, not theoretical designs

**How to participate:**
- **🗣️ Share your use cases** in [GitHub Discussions](https://github.com/agentflare-ai/agentml/discussions) - what problems are you trying to solve?
- **💡 Propose features** via [GitHub Issues](https://github.com/agentflare-ai/agentml/issues) - what's missing?
- **🔧 Contribute code** through pull requests - help build the features
- **🧪 Test and report** - try it out, break it, tell us what doesn't work
- **📚 Improve docs** - help others understand and use AgentML
- **🏗️ Build example agents** - show what's possible

### Our Goals

**Just as HTML became the universal language of the web**, enabling any browser to render any website, **we're working to make AgentML the universal language for agents**—enabling any runtime to execute any agent, and any agent to communicate with any other agent, regardless of the framework.

**Why this matters:**
- **For developers**: Learn one standard instead of multiple frameworks
- **For organizations**: Avoid vendor lock-in, standardize on one language
- **For the ecosystem**: Enable true interoperability between frameworks
- **For the future**: Build on W3C standards that will outlive current trends

**The foundation is W3C SCXML** - a mature, battle-tested specification with 20+ years of research behind it. We're extending it with agent-specific capabilities while maintaining full compatibility.

---

## Table of Contents

- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Architecture](#architecture)
- [Write Once, Deploy Anywhere](#write-once-deploy-anywhere)
- [Remote Agent Communication](#remote-agent-communication)
- [Getting Started](#getting-started)
- [Event-Driven LLM Integration](#event-driven-llm-integration)
- [Decomposition & Composition](#decomposition--composition)
- [Namespaces](#namespaces)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [LLM Generation Guide](#llm-generation-guide)

## Overview

AgentML extends SCXML with agent-specific capabilities:

- **Deterministic Behavior**: State machines provide predictable, auditable agent behavior ✅
- **Event-Driven Architecture**: LLMs generate structured events validated by JSON schemas 🚧
- **Efficient Token Usage**: Runtime snapshots and model caching minimize token consumption ✅
- **Composability**: Invoked services enable modular, reusable agent components ✅
- **Namespace Extensibility**: Plug in custom functionality through SCXML namespaces ✅

### Key Features

- 🎯 **Schema-Guided Events**: `event:schema` attributes validate LLM-generated events 🚧
- 📝 **Description-Driven**: Schema `description` fields guide LLMs to generate correct events 🚧
- 🔄 **Runtime Snapshots**: Current state + datamodel provided to LLM for context ✅
- 📦 **Modular Design**: Decompose complex agents into reusable services ✅
- 🔌 **Extensible**: Custom namespaces for memory, LLM providers, I/O, and more ✅
- 📊 **Observable**: Built-in OpenTelemetry tracing and logging ✅
- 🌐 **Universal Standard**: Write once, deploy anywhere - transform to any framework or runtime 🔮
- 🔗 **Remote Communication**: Built-in distributed agent communication via IOProcessors 🚧

**Legend:** ✅ Working | 🚧 In Development | 🔮 Planned

> **💡 Note on Schema Validation**: While the documentation describes our vision for schema-guided events, this feature is still in active development. The concepts are proven and the API design is stable, but implementation is ongoing. Follow along in [GitHub Issues](https://github.com/agentflare-ai/agentml/issues) to see progress.

## Core Concepts

### 1. AgentML as a Behavioral Model

AgentML uses SCXML state machines to define **deterministic behavior** for agents. Unlike prompt-only approaches where behavior emerges from LLM responses, AgentML:

- Explicitly defines valid states and transitions
- Constrains LLM outputs to structured events
- Provides clear lifecycle management for agent components
- Enables formal verification and testing of agent behavior

### 2. Efficient Token Caching

AgentML optimizes LLM token usage through two mechanisms:

**a) Model + Runtime Snapshot as System Prompt**

```
System Prompt = SCXML Document + Runtime Snapshot
```

The SCXML document (which changes rarely) and runtime snapshot (current state + datamodel) are provided as the system prompt. This allows:

- **Prompt Caching**: LLM providers cache the system prompt across requests
- **Minimal Updates**: Only the runtime snapshot changes between requests
- **Reduced Tokens**: User prompts can be minimal since context is in the system prompt

**b) Runtime-Generated Event Lists**

The AgentML runtime automatically generates the list of **available events** based on current state transitions. The LLM receives:

- Available event names (e.g., `intent.flight`, `action.response`)
- JSON schemas for each event's payload
- Current state configuration and datamodel

This means prompts don't need to list all possible actions—the runtime provides them dynamically.

### 3. Event Schema Validation

> ⚠️ **Work in Progress**: Event schema validation and external schema loading are in active development. APIs and features may change as we refine the implementation based on community feedback.

The `event:schema` attribute on transitions provides JSON schema validation for events:

```xml
<transition event="intent.flight"
            event:schema='{
              "type": "object",
              "description": "User wants to perform a flight-related action",
              "properties": {
                "category": {
                  "const": "flight",
                  "description": "Category must be flight for this event"
                },
                "action": {
                  "type": "string",
                  "description": "The specific action: search, book, update, or cancel"
                }
              },
              "required": ["category", "action"]
            }'
            target="handle_flight_request" />
```

Benefits:
- **Guides LLM**: Schema + descriptions clearly define what data is expected
- **Validates Output**: Runtime rejects malformed events
- **Documentation**: Schema serves as machine-readable documentation
- **Type Safety**: Ensures consistent event structure

**Critical**: Always include `description` fields at both the schema level and property level. These descriptions are the primary way to guide LLMs in generating correct event data.

### JSON Pointer References with `use:*`

> ⚠️ **Work in Progress**: External schema loading with `use:*` for OpenAPI/JSON Schema specifications is in active development. The unified pattern and automatic detection are still being refined.

AgentML uses a **unified `use:*` pattern** that automatically detects whether you're loading a namespace implementation or an external specification. The loader intelligently determines the type based on the URI/path format.

**How it works:**
- Namespace URI (e.g., `github.com/...`) → Loads custom namespace implementation
- Spec file (e.g., `./api.json`, `https://...`) → Loads OpenAPI/JSON Schema specification

This enables schema reuse via **JSON Pointer (RFC 6901)** references with namespace prefixes.

#### External Schema Files

**schemas/events.json:**
```json
{
  "components": {
    "schemas": {
      "FlightRequest": {
        "type": "object",
        "description": "User intent to search, book, update, or cancel a flight",
        "properties": {
          "category": {
            "const": "flight",
            "description": "Category identifier for flight requests"
          },
          "action": {
            "enum": ["search", "book", "update", "cancel"],
            "description": "The specific flight action to perform"
          },
          "details": {
            "type": "object",
            "description": "Flight-specific details",
            "properties": {
              "from": {
                "type": "string",
                "description": "Departure city or airport code"
              },
              "to": {
                "type": "string",
                "description": "Arrival city or airport code"
              },
              "date": {
                "type": "string",
                "format": "date",
                "description": "Departure date in YYYY-MM-DD format"
              }
            }
          }
        },
        "required": ["category", "action"]
      },
      "HotelRequest": {
        "type": "object",
        "description": "User intent for hotel booking",
        "properties": {
          "category": {"const": "hotel"},
          "action": {"enum": ["search", "book", "update", "cancel"]},
          "details": {"type": "object"}
        }
      },
      "ConfirmationAccepted": {
        "type": "object",
        "description": "User accepted the proposed action",
        "properties": {
          "confirmed": {
            "const": true,
            "description": "Must be true for acceptance"
          }
        },
        "required": ["confirmed"]
      }
    }
  }
}
```

**agent.aml:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       use:events="./schemas/events.json"
       use:gemini="github.com/agentflare-ai/agentml/gemini">

  <state id="classify_intent">
    <onentry>
      <gemini:generate location="_event" promptexpr="'Classify: ' + user_input" />
    </onentry>
    
    <!-- Reference schemas using namespace-prefixed JSON pointers -->
    <transition event="intent.flight"
                event:schema="events:#/components/schemas/FlightRequest"
                target="handle_flight" />
    
    <transition event="intent.hotel"
                event:schema="events:#/components/schemas/HotelRequest"
                target="handle_hotel" />
  </state>
  
  <state id="confirm_action">
    <transition event="confirmation.accepted"
                event:schema="events:#/components/schemas/ConfirmationAccepted"
                target="execute" />
  </state>
</agent>
```

#### Benefits of Unified `use:*` Pattern

1. **Single Pattern**: One `use:*` syntax for both namespaces and specifications
2. **Auto-Detection**: Runtime automatically detects namespace URIs vs. spec files
3. **DRY Principle**: Define schemas once, reference everywhere with namespace prefix
4. **Maintainability**: Update schema in one place, affects all references
5. **Readability**: Keep `.aml` files clean and focused on behavior
6. **Reusability**: Share schemas across multiple agents
7. **Standards Compliance**: Use OpenAPI 3.x specifications directly
8. **Validation**: External schemas can be validated independently
9. **Versioning**: Version control schemas separately from agents
10. **Clear References**: Namespace prefix shows which spec contains the schema

#### OpenAPI 3.x Integration

Load OpenAPI specifications directly:

**api-spec.yaml:**
```yaml
openapi: 3.0.0
info:
  title: Agent Event API
  version: 1.0.0
components:
  schemas:
    TaskRequest:
      type: object
      description: Request to execute a task
      properties:
        task_id:
          type: string
          format: uuid
          description: Unique task identifier
        priority:
          type: string
          enum: [low, medium, high, critical]
          description: Task priority level
        payload:
          type: object
          description: Task-specific payload data
      required:
        - task_id
        - priority
```

**agent.aml:**
```xml
<agent use:api="https://api.example.com/openapi.yaml">
  <state id="handle_task">
    <transition event="task.request"
                event:schema="api:#/components/schemas/TaskRequest"
                target="process_task" />
  </state>
</agent>
```

#### Multiple Specification Files

Load multiple specs with different namespace prefixes:

```xml
<agent use:events="./events.json"
       use:api="./openapi.yaml"
       use:external="https://external-api.com/schema.json"
       use:gemini="github.com/agentflare-ai/agentml/gemini">
  
  <!-- Reference schemas from different specs using namespace prefixes -->
  <transition event="local.event"
              event:schema="events:#/components/schemas/LocalEvent"
              target="..." />
  
  <transition event="api.event"
              event:schema="api:#/components/schemas/ApiEvent"
              target="..." />
  
  <transition event="external.event"
              event:schema="external:#/definitions/ExternalEvent"
              target="..." />
</agent>
```

**Note**: The same `use:*` pattern works for both namespace implementations (like `use:gemini`) and specification files (like `use:events`). The runtime auto-detects the type.

#### JSON Pointer Syntax

AgentML supports standard JSON Pointer syntax (RFC 6901) with namespace prefixes:

**Namespace-Prefixed Format (Recommended):**
- **`events:#/components/schemas/MySchema`** - OpenAPI 3.x style with namespace
- **`api:#/definitions/MySchema`** - JSON Schema / Swagger 2.0 with namespace
- **`models:#/$defs/MySchema`** - JSON Schema Draft 2019-09+ with namespace

**Default Format (uses first loaded spec):**
- **`#/components/schemas/MySchema`** - References first loaded specification
- **`#/definitions/MySchema`** - Alternative path for JSON Schema

**Supported JSON Pointer Paths:**
- OpenAPI 3.x: `#/components/schemas/SchemaName`
- Swagger 2.0: `#/definitions/SchemaName`
- JSON Schema: `#/definitions/SchemaName` or `#/$defs/SchemaName`

#### Remote Specifications

Load specifications from HTTP(S) URLs using the same `use:*` pattern:

```xml
<agent use:github="https://api.github.com/openapi.yaml">
  <!-- Schemas are fetched and cached at agent startup -->
  <transition event="github.webhook"
              event:schema="github:#/components/schemas/WebhookEvent"
              target="handle_webhook" />
</agent>
```

**Caching Behavior:**
- Specifications loaded once at agent initialization (both local and remote)
- Cached for agent lifetime
- Remote specs support ETags and conditional requests
- Optional refresh intervals for long-running agents

**Format Auto-Detection:**
The runtime automatically detects the specification format:
- OpenAPI 3.x (JSON/YAML)
- Swagger 2.0 (JSON/YAML)
- JSON Schema (Draft 4, 2019-09, 2020-12)

#### Schema Composition

Combine inline schemas with namespace-prefixed references:

```xml
<agent use:base="./schemas/base.json">
  <transition event="custom.event"
              event:schema='{
                "allOf": [
                  {"$ref": "base:#/components/schemas/BaseEvent"},
                  {
                    "type": "object",
                    "properties": {
                      "custom_field": {
                        "type": "string",
                        "description": "Custom field for this transition"
                      }
                    }
                  }
                ]
              }'
              target="handle_custom" />
</agent>
```

**Note**: When using `$ref` inside inline schemas, you can reference loaded specs with the same namespace-prefixed syntax.

#### Why Descriptions Are Critical

**Without Descriptions (Bad):**
```json
{
  "type": "object",
  "properties": {
    "category": {"const": "flight"},
    "action": {"type": "string"},
    "details": {"type": "object"}
  }
}
```
The LLM sees type information but has no guidance on:
- What `action` values are valid
- What should go in `details`
- What this event represents

**With Descriptions (Good):**
```json
{
  "type": "object",
  "description": "User intent to perform a flight-related action (search, book, update, cancel)",
  "properties": {
    "category": {
      "const": "flight",
      "description": "Must be 'flight' to identify this as a flight request"
    },
    "action": {
      "type": "string",
      "enum": ["search", "book", "update", "cancel"],
      "description": "The specific action: search for flights, book a new flight, update existing booking, or cancel a booking"
    },
    "details": {
      "type": "object",
      "description": "Flight-specific information extracted from user message",
      "properties": {
        "from": {
          "type": "string",
          "description": "Departure location: city name or airport code (e.g., 'New York' or 'JFK')"
        },
        "to": {
          "type": "string",
          "description": "Destination location: city name or airport code (e.g., 'London' or 'LHR')"
        },
        "date": {
          "type": "string",
          "format": "date",
          "description": "Departure date in ISO 8601 format (YYYY-MM-DD)"
        }
      }
    }
  }
}
```
The LLM now knows:
- The purpose of the event
- Valid action values and their meanings
- What information to extract for `details`
- Format requirements (e.g., ISO date format)

### 4. Minimal Prompts

Because the runtime provides:
- Current state configuration
- Available events with schemas
- Datamodel state
- Conversation history (if stored in datamodel)

The actual prompt to the LLM can be **minimal**:

```xml
<gemini:generate
  model="gemini-2.0-flash-exp"
  location="_event"
  promptexpr="'Classify user intent: ' + current_user_message" />
```

The LLM knows:
- What events it can generate (from runtime snapshot)
- What data each event needs (from event:schema)
- Current context (from datamodel in snapshot)

### 5. Decomposition Through Invoked Services

Complex agents should be **decomposed** into smaller, reusable services using SCXML's `<invoke>` mechanism:

```xml
<state id="booking">
  <invoke type="scxml" src="./payment-processor.aml">
    <param name="amount" expr="total_cost" />
    <param name="customer_id" expr="customer.id" />
  </invoke>
  
  <transition event="done.invoke.payment" target="confirmed" />
  <transition event="error.invoke.payment" target="payment_failed" />
</state>
```

**Key Lifecycle Properties:**

- **State-Scoped**: Invoked services live only while their parent state is active
- **Hierarchical**: Services in parent states remain active in child states
- **Long-Running Services**: Place invokes in hierarchical parent states for agent-lifetime services
- **Clean Termination**: Services automatically terminate when their state exits
- **Event Communication**: Services communicate via events (`done.invoke.*`, `error.invoke.*`)

**Why Decomposition Matters:**

- **Reusability**: Payment processing, authentication, etc. can be shared across agents
- **Maintainability**: Smaller state machines are easier to understand and modify
- **Testing**: Individual services can be tested in isolation
- **Composition**: Build complex agents from simple, well-tested components
- **Size Management**: Keeps individual `.aml` files focused and manageable

## Architecture

### Document Structure

AgentML documents use the `<agent>` root element instead of `<scxml>`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       use:memory="github.com/agentflare-ai/agentml/memory"
       use:gemini="github.com/agentflare-ai/agentml/gemini"
       use:stdin="github.com/agentflare-ai/agentml/stdin">

  <datamodel>
    <data id="conversation_history" expr="[]" />
    <data id="current_user_message" expr="''" />
  </datamodel>

  <!-- States and transitions -->
  
</agent>
```

### Namespace System

AgentML extends SCXML through namespaces declared with `use:*` attributes:

- `use:prefix="namespace-uri"` loads a namespace and binds it to a prefix
- Namespaces provide custom executable content (e.g., `<gemini:generate>`, `<memory:put>`)
- See [Namespaces](#namespaces) section for available namespaces

### Runtime Snapshot

The runtime snapshot is an XML document containing:

```xml
<snapshot session-id="..." timestamp="...">
  <!-- Current active states -->
  <configuration>
    <state id="awaiting_input" />
    <state id="main_flow" />
  </configuration>
  
  <!-- Datamodel values -->
  <datamodel>
    <data id="conversation_history">...</data>
    <data id="current_user_message">...</data>
  </datamodel>
  
  <!-- Event queues -->
  <queues>
    <internal>...</internal>
    <external>...</external>
  </queues>
  
  <!-- Available events with schemas -->
  <available-transitions>
    <transition event="user.input" schema="...">
      <description>User provides input</description>
    </transition>
  </available-transitions>
  
  <!-- Invoked child services (recursive snapshots) -->
  <invoked-services>
    <service id="..." session-id="...">...</service>
  </invoked-services>
</snapshot>
```

This snapshot, combined with the SCXML document, gives the LLM complete context.

## Write Once, Deploy Anywhere 🔮

> **🚧 Vision Statement**: This section describes our goal for AgentML. Framework transformers are in active development and not yet available. This is where we're heading, and we're building it in public.

AgentML aims to address a critical problem in the agentic ecosystem: **framework fragmentation**. Today, agents built for LangGraph won't run on CrewAI, n8n workflows can't be used in OpenAI's Agent Builder, and each framework has its own learning curve and limitations.

### The AgentML Solution (In Development)

**Our goal - Define once, deploy everywhere:**

```
        AgentML (.aml)
               |
    ┌──────────┼──────────┐
    ▼          ▼          ▼
LangGraph   CrewAI     n8n
    ▼          ▼          ▼
OpenAI    Autogen  agentmlx
 Agent               WASM
Builder            Runtime
```

### Supported Transformations

AgentML documents can be transformed to:

- **LangGraph**: Convert to LangGraph StateGraph with nodes and edges
- **CrewAI**: Transform to CrewAI agents with roles and tasks
- **n8n**: Generate n8n workflow JSON for visual workflow execution
- **OpenAI Agent Builder**: Export to OpenAI's agent configuration format
- **Autogen**: Convert to Microsoft Autogen conversation patterns
- **agentmlx WASM Runtime**: Run natively with our high-performance WebAssembly runtime

### Transformation Example

```bash
# Transform AgentML to LangGraph
agentmlx transform customer-support.aml --target langgraph --output customer-support.py

# Transform to n8n workflow
agentmlx transform customer-support.aml --target n8n --output customer-support.json

# Transform to OpenAI Agent Builder
agentmlx transform customer-support.aml --target openai --output customer-support-config.json
```

### Native Runtime: agentmlx

The **agentmlx** WebAssembly runtime provides native AgentML execution:

```bash
# Run directly with agentmlx
agentmlx run customer-support.aml

# Deploy to edge with WASM
agentmlx build customer-support.aml --target wasm --output agent.wasm
```

**Benefits of agentmlx:**
- **Zero Dependencies**: Pure WASM, runs anywhere (browser, edge, server)
- **High Performance**: Compiled to WebAssembly for optimal speed
- **Small Footprint**: Minimal runtime size for edge deployments
- **Sandboxed**: WASM security model isolates agent execution
- **Portable**: Run the same `.wasm` file on any platform

### Why This Matters

**For Developers:**
- Learn one standard instead of multiple frameworks
- Reuse agents across projects and platforms
- Avoid vendor lock-in
- Leverage best tool for each deployment target

**For Organizations:**
- Standardize on one agent definition language
- Migrate between frameworks without rewriting
- Deploy agents to optimal runtime for each use case
- Future-proof agent investments

**For the Ecosystem:**
- Establishes W3C SCXML as the agent standard
- Enables framework interoperability
- Accelerates agent development and sharing
- Creates a universal agent marketplace

### Framework Compatibility

Each transformation preserves AgentML semantics while adapting to framework idioms:

| Feature | LangGraph | CrewAI | n8n | OpenAI | agentmlx |
|---------|-----------|---------|-----|--------|----------|
| State Machines | ✅ StateGraph | ✅ Workflows | ✅ Workflows | ⚠️ Limited | ✅ Native |
| Event Schemas | ✅ Pydantic | ✅ Pydantic | ✅ JSON Schema | ✅ Function Calls | ✅ Native |
| Invoked Services | ✅ Subgraphs | ✅ Sub-crews | ✅ Sub-workflows | ⚠️ Limited | ✅ Native |
| Datamodel | ✅ State | ✅ Context | ✅ Variables | ✅ Memory | ✅ Native |
| LLM Integration | ✅ LangChain | ✅ LiteLLM | ✅ Built-in | ✅ OpenAI | ✅ Pluggable |

✅ Full Support | ⚠️ Partial Support | ❌ Not Supported

### Transformation Guarantees

When transforming AgentML, we guarantee:

1. **Behavioral Equivalence**: Transformed agents behave identically to the source
2. **Schema Preservation**: Event schemas are converted to native framework types
3. **State Machine Semantics**: SCXML state machine semantics are preserved
4. **Datamodel Integrity**: Data flow and transformations are maintained
5. **Error Handling**: Error events and transitions are properly mapped

### Coming Soon

Additional transformations in development:

- **Langflow**: Visual flow-based agent builder
- **Flowise**: Low-code LLM app builder
- **Dify**: LLMOps platform
- **Haystack**: NLP framework for LLM applications
- **Semantic Kernel**: Microsoft's AI orchestration SDK

### The W3C Standard for Agents

**AgentML to agents is what HTML is to the web.** Just as HTML became the universal markup language that enabled the web to flourish—allowing any browser to render any page—AgentML establishes a universal language for agents built on the solid foundation of W3C standards.

By building on **W3C SCXML**, AgentML leverages:

- **20+ years** of formal state machine research
- **Battle-tested** specification used in telephony, automotive, and more
- **Formal semantics** enabling verification and tooling
- **Vendor-neutral** W3C standard, not controlled by any company
- **Future-proof** foundation that will outlive current frameworks

AgentML extends SCXML with agent-specific capabilities while maintaining full compatibility with the W3C specification. This means:

- Any SCXML tool can read AgentML documents
- AgentML can leverage SCXML verification tools
- Standard SCXML transformations (XSLT, validation) work with AgentML
- Cross-industry standardization enables broad ecosystem
- **True interoperability**: Just as HTML enabled the web, AgentML enables the agentic ecosystem

## Remote Agent Communication 🚧

> **🚧 In Development**: The IOProcessor interface is designed and documented here, but implementations are still being built. The architecture follows W3C SCXML standards and will enable distributed agent communication.

AgentML is designed to support **distributed agent communication** using the W3C SCXML IOProcessor interface. This will enable agents to communicate across processes, machines, and networks using standard protocols.

### IOProcessor Architecture

The IOProcessor interface provides a standardized way for agents to send and receive events:

```go
type IOProcessor interface {
    // Handle processes incoming events via this transport
    Handle(ctx context.Context, event *Event) error
    
    // Location returns the URI where this agent can be reached
    Location(ctx context.Context) (string, error)
    
    // Type returns the processor type (e.g., "http://www.w3.org/TR/scxml/#HTTPEventProcessor")
    Type() string
    
    // Shutdown cleans up resources
    Shutdown(ctx context.Context) error
}
```

### Built-in IOProcessors

AgentML includes several standard IOProcessors:

#### 1. HTTP IOProcessor

Send events to agents via HTTP/HTTPS:

```xml
<state id="notify_remote_agent">
  <onentry>
    <!-- Send event to remote agent via HTTP -->
    <send event="task.assigned"
          target="https://agent.example.com/events"
          type="http://www.w3.org/TR/scxml/#HTTPEventProcessor">
      <param name="task_id" expr="task.id" />
      <param name="priority" expr="'high'" />
    </send>
  </onentry>
  
  <!-- Wait for response from remote agent -->
  <transition event="task.acknowledged" target="confirmed" />
  <transition event="error.communication" target="retry" />
</state>
```

**Features:**
- RESTful HTTP/HTTPS communication
- Automatic request/response correlation
- Built-in retry with exponential backoff
- TLS support for secure communication
- OpenTelemetry trace propagation

#### 2. WebSocket IOProcessor

Real-time bidirectional communication:

```xml
<state id="collaborative_work">
  <!-- Connect to remote agent via WebSocket -->
  <invoke type="websocket" id="partner_agent">
    <param name="url" expr="'wss://partner.example.com/ws'" />
  </invoke>
  
  <state id="working">
    <onentry>
      <!-- Send updates to partner in real-time -->
      <send event="progress.update"
            targetexpr="'#_websocket_partner_agent'"
            type="http://www.w3.org/TR/scxml/#WebSocketEventProcessor">
        <param name="percent_complete" expr="progress" />
      </send>
    </onentry>
    
    <!-- Receive real-time updates from partner -->
    <transition event="partner.progress.update" target="sync_state">
      <assign location="partner_progress" expr="_event.data.percent_complete" />
    </transition>
  </state>
</state>
```

**Features:**
- Persistent bidirectional connections
- Low-latency event delivery
- Automatic reconnection
- Message ordering guarantees
- Heartbeat/keepalive support

#### 3. SCXML IOProcessor

Direct agent-to-agent communication (same process or via network):

```xml
<state id="delegate_task">
  <onentry>
    <!-- Send event to another SCXML agent -->
    <send event="task.delegate"
          targetexpr="'#session_' + worker_agent_id"
          type="http://www.w3.org/TR/scxml/#SCXMLEventProcessor">
      <param name="task" expr="current_task" />
    </send>
  </onentry>
  
  <transition event="done.invoke" target="complete">
    <assign location="result" expr="_event.data" />
  </transition>
</state>
```

**Features:**
- Zero-copy local communication (same process)
- Networked SCXML communication via HTTP/WebSocket
- Session management and routing
- Hierarchical agent addressing

#### 4. Internal IOProcessor

Fast in-memory event delivery (default):

```xml
<state id="process">
  <onentry>
    <!-- Raise internal event (default IOProcessor) -->
    <raise event="internal.ready" />
    
    <!-- Or explicitly use internal IOProcessor -->
    <send event="internal.ready"
          type="http://www.w3.org/TR/scxml/#InternalEventProcessor" />
  </onentry>
</state>
```

### Distributed Agent Patterns

#### Agent Swarm

Multiple agents coordinating on a task:

```xml
<state id="coordinate_swarm">
  <datamodel>
    <data id="swarm_agents" expr="['http://agent1.com', 'http://agent2.com', 'http://agent3.com']" />
    <data id="responses" expr="[]" />
  </datamodel>
  
  <onentry>
    <!-- Broadcast task to all agents -->
    <foreach array="swarm_agents" item="agent_url">
      <send event="task.execute"
            targetexpr="agent_url"
            type="http://www.w3.org/TR/scxml/#HTTPEventProcessor">
        <param name="task_data" expr="task" />
      </send>
    </foreach>
  </onentry>
  
  <!-- Collect responses -->
  <transition event="task.complete">
    <script>
      responses.push(_event.data);
    </script>
    <if cond="responses.length === swarm_agents.length">
      <raise event="all.complete" />
    </if>
  </transition>
  
  <transition event="all.complete" target="aggregate_results" />
</state>
```

#### Supervisor-Worker Pattern

Hierarchical agent delegation:

```xml
<state id="supervisor">
  <state id="assign_work">
    <onentry>
      <!-- Invoke worker agents -->
      <invoke type="scxml" id="worker1" src="./worker-agent.aml">
        <param name="supervisor" expr="_ioprocessors.http.location" />
        <param name="task" expr="tasks[0]" />
      </invoke>
      
      <invoke type="scxml" id="worker2" src="./worker-agent.aml">
        <param name="supervisor" expr="_ioprocessors.http.location" />
        <param name="task" expr="tasks[1]" />
      </invoke>
    </onentry>
    
    <!-- Workers report back via HTTP -->
    <transition event="worker.complete" target="check_completion">
      <assign location="completed_tasks" expr="completed_tasks + 1" />
    </transition>
    
    <transition event="worker.error" target="reassign_work">
      <assign location="failed_tasks" expr="failed_tasks + 1" />
    </transition>
  </state>
</state>
```

#### Request-Response Pattern

Synchronous-style communication with timeout:

```xml
<state id="query_remote">
  <onentry>
    <!-- Send query with auto-generated ID -->
    <send event="query.execute"
          target="https://remote-agent.com/query"
          type="http://www.w3.org/TR/scxml/#HTTPEventProcessor"
          idlocation="query_id">
      <param name="query" expr="user_query" />
    </send>
    
    <!-- Set timeout for response -->
    <send event="query.timeout" delay="30s">
      <param name="query_id" expr="query_id" />
    </send>
  </onentry>
  
  <!-- Handle response -->
  <transition event="query.response"
              cond="_event.data.query_id === query_id"
              target="process_response">
    <cancel sendidexpr="query_id + '_timeout'" />
    <assign location="result" expr="_event.data.result" />
  </transition>
  
  <!-- Handle timeout -->
  <transition event="query.timeout"
              cond="_event.data.query_id === query_id"
              target="timeout_error" />
</state>
```

#### Pub/Sub Pattern

Event broadcasting with subscriptions:

```xml
<state id="event_hub">
  <datamodel>
    <data id="subscribers" expr="{}" />
  </datamodel>
  
  <!-- Handle subscription requests -->
  <transition event="subscribe.request">
    <script>
      var topic = _event.data.topic;
      var callback_url = _event.data.callback_url;
      
      if (!subscribers[topic]) subscribers[topic] = [];
      subscribers[topic].push(callback_url);
    </script>
    
    <!-- Acknowledge subscription -->
    <send event="subscribe.confirmed"
          targetexpr="_event.origin"
          type="http://www.w3.org/TR/scxml/#HTTPEventProcessor">
      <param name="topic" expr="_event.data.topic" />
    </send>
  </transition>
  
  <!-- Broadcast events to subscribers -->
  <transition event="publish.*">
    <script>
      var topic = _event.name.split('.')[1];
      var subs = subscribers[topic] || [];
      
      subs.forEach(function(url) {
        // Send will be called per subscriber
        sendToSubscriber(url, _event.data);
      });
    </script>
  </transition>
</state>
```

### Service Discovery

Agents can discover each other using the `_ioprocessors` system variable:

```xml
<state id="register_with_discovery">
  <onentry>
    <script>
      // Get this agent's HTTP endpoint
      var my_location = _ioprocessors.http.location;
      
      // Register with discovery service
    </script>
    
    <send event="agent.register"
          target="https://discovery.example.com/register"
          type="http://www.w3.org/TR/scxml/#HTTPEventProcessor">
      <param name="agent_id" expr="_sessionid" />
      <param name="location" expr="_ioprocessors.http.location" />
      <param name="capabilities" expr="['booking', 'search', 'payment']" />
    </send>
  </onentry>
</state>
```

### Security & Authentication

IOProcessors support standard security mechanisms:

```xml
<state id="secure_communication">
  <onentry>
    <!-- Send with authentication -->
    <send event="secure.request"
          target="https://secure-agent.com/api"
          type="http://www.w3.org/TR/scxml/#HTTPEventProcessor">
      <param name="_auth_type" expr="'bearer'" />
      <param name="_auth_token" expr="auth_token" />
      <param name="sensitive_data" expr="data" />
    </send>
  </onentry>
</state>
```

**Security Features:**
- TLS/SSL encryption
- Bearer token authentication
- OAuth 2.0 support
- mTLS (mutual TLS) for agent-to-agent auth
- Request signing
- Rate limiting per remote endpoint

### Observability

All IOProcessor communication includes:

- **Distributed Tracing**: OpenTelemetry trace context propagated automatically
- **Metrics**: Request/response times, error rates, queue depths
- **Logging**: Structured logs for all communication events
- **Health Checks**: Built-in liveness/readiness endpoints

```xml
<state id="monitored_communication">
  <onentry>
    <!-- All sends automatically include trace context -->
    <send event="task.execute"
          target="https://remote-agent.com/tasks"
          type="http://www.w3.org/TR/scxml/#HTTPEventProcessor">
      <param name="task" expr="task_data" />
      <!-- Trace context automatically added to headers/metadata -->
    </send>
  </onentry>
</state>
```

### Benefits of Built-in Remote Communication

1. **Standard Protocol**: W3C SCXML IOProcessor specification ensures interoperability
2. **Framework Agnostic**: Agents can communicate regardless of implementation (Go, Python, JS)
3. **Transport Flexibility**: Choose HTTP, WebSocket, or custom transports based on needs
4. **Automatic Routing**: Event routing handled by runtime, not application code
5. **Trace Propagation**: Distributed traces work out of the box
6. **Type Safety**: Event schemas validated across agent boundaries
7. **Retry & Resilience**: Built-in retry, timeout, and error handling
8. **Zero Configuration**: IOProcessors configured automatically by runtime

### Cross-Framework Communication

Because AgentML uses standard protocols, agents can communicate across frameworks:

```
┌─────────────┐         HTTP/WebSocket        ┌─────────────┐
│  AgentML    │  ◄──────────────────────────► │  LangGraph  │
│   (Go)      │    Event: task.execute        │   (Python)  │
└─────────────┘    Schema: {...}              └─────────────┘
       │                                              │
       │                                              │
       ▼                                              ▼
┌─────────────┐                              ┌─────────────┐
│   CrewAI    │                              │    n8n      │
│  (Python)   │                              │   (Node)    │
└─────────────┘                              └─────────────┘
```

All agents use the same event schemas and IOProcessor protocols, enabling true polyglot agent systems.

## Getting Started

> **⚠️ Alpha Software**: AgentML is in early alpha. APIs may change, features may be incomplete, and you may encounter bugs. We welcome feedback and contributions!

### Installation

```bash
# Note: Package structure is still being finalized
go get github.com/agentflare-ai/agentml
```

### Basic Agent Example

This example shows the core concepts. Note that some features (like `event:schema` validation) are still in development.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       use:gemini="github.com/agentflare-ai/agentml/gemini">

  <datamodel>
    <data id="user_input" expr="''" />
    <data id="response" expr="''" />
  </datamodel>

  <state id="main">
    <state id="awaiting_input">
      <onentry>
        <!-- Get user input -->
        <assign location="user_input" expr="getUserInput()" />
      </onentry>
      
      <transition target="processing" />
    </state>

    <state id="processing">
      <onentry>
        <!-- LLM generates event based on input -->
        <gemini:generate
          model="gemini-2.0-flash-exp"
          location="_event"
          promptexpr="'Process this input: ' + user_input" />
      </onentry>
      
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

### Running an Agent

```go
package main

import (
    "context"
    "github.com/agentflare-ai/agentml"
    "github.com/agentflare-ai/agentml/agent"
)

func main() {
    ctx := context.Background()
    
    // Load agent document
    doc, err := xmldom.ParseFile("agent.aml")
    if err != nil {
        panic(err)
    }
    
    // Create interpreter
    interp, err := agentml.NewInterpreter(ctx, doc)
    if err != nil {
        panic(err)
    }
    
    // Start the agent
    if err := interp.Start(ctx); err != nil {
        panic(err)
    }
    
    // Agent runs until completion
    <-interp.Done()
}
```

## Event-Driven LLM Integration

### How LLMs Generate Events

1. **System Prompt**: LLM receives SCXML model + runtime snapshot
2. **User Prompt**: Minimal prompt with specific task (e.g., "Classify user intent")
3. **Event Generation**: LLM generates structured JSON event
4. **Schema Validation**: Runtime validates event against `event:schema`
5. **Transition**: If valid, state machine transitions to target state

### Event Schema Best Practices

**Always Include Descriptions** ⭐

The `description` field at both the schema level and property level is **crucial** for LLM success. Descriptions guide the LLM in understanding what data to generate:

```xml
<transition event="intent.flight"
            event:schema='{
              "type": "object",
              "description": "User intent to search, book, update, or cancel a flight",
              "properties": {
                "category": {
                  "const": "flight",
                  "description": "Category identifier for flight-related requests"
                },
                "action": {
                  "enum": ["search", "book", "update", "cancel"],
                  "description": "The specific flight action the user wants to perform"
                },
                "details": {
                  "type": "object",
                  "description": "Flight-specific details extracted from user input",
                  "properties": {
                    "from": {
                      "type": "string",
                      "description": "Departure city or airport code"
                    },
                    "to": {
                      "type": "string",
                      "description": "Arrival city or airport code"
                    },
                    "date": {
                      "type": "string",
                      "format": "date",
                      "description": "Departure date in YYYY-MM-DD format"
                    }
                  }
                }
              },
              "required": ["category", "action"]
            }'
            target="handle_flight" />
```

**Use Const for Specific Values**

```xml
<transition event="confirmation.accepted"
            event:schema='{
              "type": "object",
              "description": "User accepted the proposed action",
              "properties": {
                "confirmed": {
                  "const": true,
                  "description": "Confirmation status - must be true for this event"
                }
              },
              "required": ["confirmed"]
            }'
            target="execute_action" />
```

**Use Enums with Descriptions**

```xml
event:schema='{
  "type": "object",
  "description": "Status update for a booking request",
  "properties": {
    "status": {
      "enum": ["pending", "approved", "rejected"],
      "description": "Current status of the booking request"
    },
    "reason": {
      "type": "string",
      "description": "Human-readable reason for the status (required for rejected)"
    }
  }
}'
```

### Accessing Event Data

Events are stored in the `_event` system variable:

```xml
<script>
  var action = _event.data.action;
  var details = _event.data.details;
  var category = _event.data.category;
</script>
```

## Decomposition & Composition

### When to Decompose

Decompose your agent when:

- **Document Size**: Single `.aml` file exceeds ~500-1000 lines
- **Reusability**: Logic is needed across multiple agents
- **Complexity**: Too many states/transitions to reason about
- **Team Size**: Multiple developers working on different features
- **Domain Separation**: Clear boundaries between concerns (auth, payment, booking, etc.)

### Service Lifecycle

Understanding service lifecycle is crucial:

```xml
<!-- Service lives only while in 'authenticated' state -->
<state id="authenticated">
  <invoke type="scxml" src="./session-manager.aml">
    <param name="user_id" expr="user.id" />
  </invoke>
  
  <state id="browsing">
    <!-- session-manager still running -->
  </state>
  
  <state id="checkout">
    <!-- session-manager still running -->
  </state>
  
  <transition event="user.logout" target="logged_out" />
</state>

<!-- Service terminated when exiting 'authenticated' state -->
<state id="logged_out">
  <!-- session-manager no longer running -->
</state>
```

### Agent-Lifetime Services

For services that should run for the entire agent lifetime, use a hierarchical parent state:

```xml
<agent ...>
  <state id="root">
    <!-- This service runs for the entire agent lifetime -->
    <invoke type="scxml" src="./memory-manager.aml">
      <param name="db_path" expr="db_path" />
    </invoke>
    
    <!-- All agent logic as child states -->
    <state id="initializing">...</state>
    <state id="running">...</state>
    <state id="shutting_down">...</state>
    
    <transition event="done.state.root" target="end" />
  </state>
  
  <final id="end" />
</agent>
```

### Service Communication

Services communicate through events:

```xml
<!-- Parent agent -->
<state id="processing">
  <invoke type="scxml" id="processor" src="./data-processor.aml">
    <param name="data" expr="raw_data" />
  </invoke>
  
  <!-- Listen for completion -->
  <transition event="done.invoke.processor" target="completed">
    <assign location="result" expr="_event.data" />
  </transition>
  
  <!-- Handle errors -->
  <transition event="error.invoke.processor" target="error">
    <assign location="error_message" expr="_event.data" />
  </transition>
</state>

<!-- Child service (data-processor.aml) -->
<agent ...>
  <datamodel>
    <data id="data" />
  </datamodel>
  
  <state id="process">
    <onentry>
      <script>
        var processed = processData(data);
      </script>
    </onentry>
    <transition target="done" />
  </state>
  
  <final id="done">
    <donedata>
      <param name="result" expr="processed" />
    </donedata>
  </final>
</agent>
```

### Composition Patterns

**Pipeline Pattern**

```xml
<state id="pipeline">
  <state id="step1">
    <invoke type="scxml" src="./fetch-data.aml" />
    <transition event="done.invoke" target="step2">
      <assign location="data" expr="_event.data" />
    </transition>
  </state>
  
  <state id="step2">
    <invoke type="scxml" src="./transform-data.aml">
      <param name="input" expr="data" />
    </invoke>
    <transition event="done.invoke" target="step3">
      <assign location="transformed" expr="_event.data" />
    </transition>
  </state>
  
  <state id="step3">
    <invoke type="scxml" src="./store-data.aml">
      <param name="data" expr="transformed" />
    </invoke>
    <transition event="done.invoke" target="complete" />
  </state>
</state>
```

**Parallel Services**

```xml
<parallel id="multi_task">
  <state id="task_a">
    <invoke type="scxml" src="./task-a.aml" />
  </state>
  
  <state id="task_b">
    <invoke type="scxml" src="./task-b.aml" />
  </state>
  
  <state id="task_c">
    <invoke type="scxml" src="./task-c.aml" />
  </state>
  
  <transition event="done.state.multi_task" target="all_complete" />
</parallel>
```

## Namespaces

AgentML's functionality is extended through namespaces declared with `use:*` attributes.

### Available Namespaces

#### Agent (`github.com/agentflare-ai/agentml/agent`)

Core namespace providing the `<agent>` root element and `event:schema` validation.

```xml
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript">
  <!-- Core SCXML content -->
</agent>
```

See [agent/README.md](./agent/README.md) for details.

#### Gemini (`github.com/agentflare-ai/agentml/gemini`)

Google Gemini LLM integration with rate limiting and model management.

```xml
<agent use:gemini="github.com/agentflare-ai/agentml/gemini">
  <gemini:generate
    model="gemini-2.0-flash-exp"
    location="_event"
    promptexpr="'Your prompt here'" />
</agent>
```

Features:
- Multiple model support
- Advanced rate limiting
- OpenTelemetry tracing
- Structured output generation

See [gemini/README.md](./gemini/README.md) for details.

#### Memory (`github.com/agentflare-ai/agentml/memory`)

High-performance memory with SQLite, vector search, and graph database capabilities.

```xml
<agent use:memory="github.com/agentflare-ai/agentml/memory">
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

#### Stdin (`github.com/agentflare-ai/agentml/stdin`)

Simple stdin/stdout I/O for console agents.

```xml
<agent use:stdin="github.com/agentflare-ai/agentml/stdin">
  <stdin:read location="user_input" prompt="You: " />
</agent>
```

#### Ollama (`github.com/agentflare-ai/agentml/ollama`)

Local LLM integration via Ollama.

```xml
<agent use:ollama="github.com/agentflare-ai/agentml/ollama">
  <ollama:generate
    model="llama2"
    location="_event"
    promptexpr="'Your prompt here'" />
</agent>
```

See [ollama/README.md](./ollama/README.md) for details.

### Creating Custom Namespaces

To create a custom namespace:

1. **Define the namespace URI** (e.g., `github.com/yourorg/agentml-custom`)
2. **Implement the `Namespace` interface** from `types.go`
3. **Create custom executable elements** implementing the `Executor` interface
4. **Provide an XSD schema** for IDE validation
5. **Register the namespace loader**

Example structure:

```go
package custom

import "github.com/agentflare-ai/agentml"

const NamespaceURI = "github.com/yourorg/agentml-custom"

func Loader() agentml.NamespaceLoader {
    return func(ctx context.Context, itp agentml.Interpreter, doc xmldom.Document) (agentml.Namespace, error) {
        return &customNamespace{itp: itp}, nil
    }
}

type customNamespace struct {
    itp agentml.Interpreter
}

func (n *customNamespace) URI() string { return NamespaceURI }

func (n *customNamespace) Handle(ctx context.Context, el xmldom.Element) (bool, error) {
    switch el.LocalName() {
    case "custom-action":
        exec := &customActionExec{Element: el}
        return true, exec.Execute(ctx, n.itp)
    }
    return false, nil
}

func (n *customNamespace) Unload(ctx context.Context) error { return nil }
```

## Best Practices

### Document Organization

1. **Keep `.aml` files focused**: Target 200-500 lines per file
2. **Use meaningful state IDs**: `handle_flight_request` not `state_5`
3. **Document event flows**: Use XML comments to explain transitions
4. **Group related states**: Use hierarchical states for related functionality

### Event Design

1. **Use specific event names**: `intent.flight.search` vs `process`
2. **Validate with schemas**: Always use `event:schema` on transitions
3. **Use namespace-prefixed JSON pointers**: Load with `use:events="./events.json"` and reference with `events:#/components/schemas/MySchema`
4. **Consistent data structure**: Use the same schema across similar events
5. **Const for literals**: Use `{"const": "value"}` for specific values
6. **Version your schemas**: Keep event schemas in version-controlled files
7. **Clear namespace prefixes**: Use descriptive prefixes like `api`, `events`, `models`

### Datamodel Management

1. **Minimal state**: Store only what's needed in datamodel
2. **Clear naming**: `conversation_history` vs `data1`
3. **Initialize properly**: Set default values in `<datamodel>` section
4. **Type consistency**: ECMAScript datamodel allows flexible types

### Script Handling

1. **Prefer external scripts**: Use `<script src="./scripts/utils.js" />` instead of inline scripts
2. **Use CDATA for inline scripts**: Wrap inline scripts in `<![CDATA[...]]>` to avoid XML escaping issues
3. **Keep scripts minimal**: Complex logic should be in external files or custom namespaces
4. **Avoid special characters**: If not using CDATA, escape `<`, `>`, `&` characters

**External Script (Recommended):**
```xml
<state id="process">
  <onentry>
    <script src="./scripts/data-processor.js" />
  </onentry>
</state>
```

**Inline Script with CDATA:**
```xml
<state id="process">
  <onentry>
    <script>
      <![CDATA[
        // Use CDATA to avoid escaping < > & characters
        if (data.value < 10 && data.status !== 'complete') {
          result = processData(data);
        }
      ]]>
    </script>
  </onentry>
</state>
```

**Without CDATA (Requires Escaping):**
```xml
<script>
  // Must escape special characters
  if (data.value &lt; 10 &amp;&amp; data.status !== 'complete') {
    result = processData(data);
  }
</script>
```

### Schema Organization

1. **Separate schema files**: Keep event schemas in `schemas/` directory
2. **OpenAPI format**: Use OpenAPI 3.x for API-aligned agents
3. **Descriptive paths**: Organize by domain (e.g., `schemas/flights.json`, `schemas/hotels.json`)
4. **Version control**: Track schema versions separately from agent logic
5. **Shared schemas**: Create common schema libraries for reuse across agents
6. **Documentation**: Use JSON Schema `description` fields extensively

**Example Project Structure:**
```
project/
├── agents/
│   ├── customer-support.aml
│   ├── booking-agent.aml
│   └── notification-agent.aml
├── schemas/
│   ├── events.json           # Common event schemas
│   ├── flights.json          # Flight-specific schemas
│   ├── hotels.json           # Hotel-specific schemas
│   └── common/
│       ├── base-event.json   # Base event schema
│       └── error-response.json
├── scripts/
│   ├── utils.js              # Reusable utility functions
│   ├── data-processor.js     # Data processing logic
│   └── validators.js         # Custom validation functions
└── specs/
    └── openapi.yaml          # External API specifications
```

### LLM Prompt Engineering

1. **Minimal prompts**: Rely on runtime snapshot for context
2. **Clear instructions**: "Classify user intent" vs "What should we do?"
3. **System prompts**: Store reusable context in datamodel
4. **Event examples**: Include examples in comments for LLM generation

### Service Decomposition

1. **Single responsibility**: Each service should have one clear purpose
2. **Clear interfaces**: Well-defined parameters and return events
3. **Lifecycle awareness**: Know when services start/stop
4. **Error handling**: Handle `error.invoke.*` events

### Testing

1. **Unit test services**: Test individual `.aml` files in isolation
2. **Integration tests**: Test composed agents end-to-end
3. **Mock LLMs**: Use deterministic responses for testing
4. **Event validation**: Verify event schemas match expectations
5. **Test external scripts**: Separately test JavaScript/script files used in agents

### Common Pitfalls

1. **XML Special Characters**: Always use CDATA for inline scripts with `<`, `>`, `&` characters
2. **Missing Schema Descriptions**: LLMs need descriptions to generate correct events
3. **Service Lifecycle**: Remember invoked services only live while their parent state is active
4. **Event Name Matching**: Use specific event names, wildcards (`*`) only for fallback
5. **Datamodel Initialization**: Initialize all data elements with default values

## Examples

See the [examples](./examples/) directory for complete agent implementations:

- **[customer_support](./examples/customer_support/)**: Full customer support bot for airlines
  - Flight, hotel, car rental, and excursion booking
  - Policy lookups
  - User confirmation flows
  - Multi-intent classification

## LLM Generation Guide

> **For LLMs Generating AgentML Documents**

When generating AgentML documents, follow these guidelines:

### Schema Description Checklist ⭐

Before generating any `event:schema`, ensure:

- [ ] ✅ Schema has a top-level `description` explaining what the event represents
- [ ] ✅ Every property has a `description` explaining its purpose and format
- [ ] ✅ For enums, description lists valid values and their meanings
- [ ] ✅ For objects, description explains what data should be included
- [ ] ✅ For strings with formats (date, email, etc.), description specifies the expected format
- [ ] ✅ For const values, description explains why this specific value is required
- [ ] ✅ Nested objects have descriptions at each level of nesting

**Remember**: Descriptions are how you communicate intent to the LLM. Without them, the LLM must guess what data to generate.

### Document Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       use:events="./schemas/events.json"
       use:gemini="github.com/agentflare-ai/agentml/gemini">
  
  <!-- 1. Datamodel: Define all state variables -->
  <datamodel>
    <data id="variable_name" expr="initial_value" />
  </datamodel>
  
  <!-- 2. States: Define agent behavior -->
  <state id="main">
    <!-- Hierarchical states for structure -->
  </state>
  
  <!-- 3. Final states: Define exit conditions -->
  <final id="end" />
</agent>
```

### Event Generation Pattern

When using LLMs to generate events:

```xml
<state id="classify_intent">
  <onentry>
    <!-- 
      NOTE: The runtime provides:
      - Current state configuration
      - Available events with schemas
      - Datamodel values
      
      Keep the prompt minimal!
    -->
    <gemini:generate
      model="gemini-2.0-flash-exp"
      location="_event"
      promptexpr="'Classify user intent: ' + user_message" />
  </onentry>
  
  <!-- Define ALL possible events with schemas (WITH DESCRIPTIONS!) -->
  <!-- Option 1: Inline schemas -->
  <transition event="intent.flight"
              event:schema='{
                "type": "object",
                "description": "User intent to search, book, update, or cancel a flight",
                "properties": {
                  "category": {
                    "const": "flight",
                    "description": "Category identifier for flight requests"
                  },
                  "action": {
                    "type": "string",
                    "description": "Action type: search, book, update, or cancel"
                  }
                },
                "required": ["category", "action"]
              }'
              target="handle_flight" />
  
  <!-- Option 2: Reference from loaded spec (recommended) -->
  <transition event="intent.hotel"
              event:schema="events:#/components/schemas/HotelRequest"
              target="handle_hotel" />
  
  <!-- Always include fallback for unmatched events -->
  <transition event="*" target="handle_error" />
</state>
```

### Common Patterns

**User Input Loop**

```xml
<state id="main_loop">
  <state id="await_input">
    <onentry>
      <stdin:read location="user_input" prompt="You: " />
    </onentry>
    <transition target="process_input" />
  </state>
  
  <state id="process_input">
    <onentry>
      <gemini:generate ... />
    </onentry>
    <!-- Events and transitions -->
  </state>
  
  <state id="respond">
    <onentry>
      <log expr="response_message" />
    </onentry>
    <transition target="await_input" />
  </state>
</state>
```

**Processing with External Script**

```xml
<state id="process_data">
  <onentry>
    <!-- Load external script for complex processing -->
    <script src="./scripts/data-processor.js" />
    
    <!-- Or use inline script with CDATA -->
    <script>
      <![CDATA[
        // Complex logic without XML escaping concerns
        var result = {
          valid: data.value < threshold && data.status === 'pending',
          processed: processComplexData(data)
        };
      ]]>
    </script>
  </onentry>
  
  <transition cond="result.valid" target="success" />
  <transition target="failure" />
</state>
```

**Confirmation Flow**

```xml
<state id="confirm_action">
  <onentry>
    <assign location="pending_action" expr="_event.data" />
    <log expr="'Confirm: ' + pending_action.message" />
    <stdin:read location="confirmation_response" prompt="(yes/no): " />
    <gemini:generate
      location="_event"
      promptexpr="'Did user confirm? Response: ' + confirmation_response" />
  </onentry>
  
  <transition event="confirmation.accepted"
              event:schema='{
                "type": "object",
                "description": "User accepted the proposed action",
                "properties": {
                  "confirmed": {
                    "const": true,
                    "description": "Must be true for acceptance"
                  }
                },
                "required": ["confirmed"]
              }'
              target="execute_action" />
  
  <transition event="confirmation.declined"
              event:schema='{
                "type": "object",
                "description": "User declined the proposed action",
                "properties": {
                  "confirmed": {
                    "const": false,
                    "description": "Must be false for decline"
                  }
                },
                "required": ["confirmed"]
              }'
              target="cancel_action" />
</state>
```

**Invoked Service**

```xml
<state id="parent">
  <invoke type="scxml" id="child_service" src="./child.aml">
    <!-- Pass parameters -->
    <param name="input_data" expr="data" />
  </invoke>
  
  <!-- Handle completion -->
  <transition event="done.invoke.child_service" target="completed">
    <assign location="result" expr="_event.data" />
  </transition>
  
  <!-- Handle errors -->
  <transition event="error.invoke.child_service" target="error_handler">
    <assign location="error_info" expr="_event.data" />
  </transition>
</state>
```

### Key Reminders

1. **Always use `event:schema` with descriptions** ⭐ — Both schema-level and property-level descriptions are crucial for LLM success
2. **Use namespace-prefixed JSON pointers** ⭐ — Load schemas with `use:events="./events.json"` and reference with `event:schema="events:#/path/to/schema"`
3. **Unified `use:*` pattern** ⭐ — Use the same pattern for both namespaces (`use:gemini="github.com/..."`) and specs (`use:api="./api.json"`)
4. **External scripts preferred** ⭐ — Use `<script src="..." />` for reusability; if inline, wrap in `<![CDATA[...]]>` to avoid escaping
5. **Describe every property** — Even simple properties benefit from clear descriptions
6. **Keep prompts minimal** — context comes from runtime snapshot
7. **Use hierarchical states** for agent-lifetime services
8. **Handle errors** — include fallback transitions for unexpected events
9. **Document event flows** — use XML comments to explain complex transitions
10. **Test event schemas** — ensure LLM can generate valid events

### Token Optimization

To minimize token usage:

1. **Store conversation history in datamodel** instead of passing in prompts
2. **Reference datamodel variables** via `promptexpr` instead of copying values
3. **Let runtime generate event lists** — don't list them in prompts
4. **Use external schemas with `use:*`** ⭐ — Load schemas once with `use:events="./events.json"`, reference with `event:schema="events:#/..."`
5. **Use schema descriptions** ⭐ — Put ALL documentation in external schema files, not prompts. The runtime sends schemas to the LLM automatically.

**Why External Schemas Matter:**

The runtime automatically provides available events and their schemas to the LLM. By using external schema files with the `use:*` pattern:
- **Better caching**: Schema files change rarely, maximizing LLM prompt cache hits
- **Guide the LLM** without inflating your prompt
- **Reusability**: Share schemas across multiple agents
- **Maintainability**: Update schemas independently from agent logic
- **Ensure consistent** event generation across different prompts
- **Self-documenting**: Schemas serve both LLMs and humans

Example:

```xml
<!-- GOOD: Minimal prompt, runtime provides context -->
<gemini:generate
  location="_event"
  promptexpr="'Classify: ' + user_input" />

<!-- BAD: Repeating information already in runtime snapshot -->
<gemini:generate
  location="_event"
  promptexpr="'You are in state X. Available events are Y, Z. Datamodel has A, B, C. Now classify: ' + user_input" />
```

---

## Roadmap

> **📅 Timeline Transparency**: These are our goals, not promises. As an early-stage project being built in public, timelines may shift based on community feedback, priorities, and contributions. We'll keep this updated as we learn and adapt.

### ✅ What's Working Now (Early Alpha)

- Core SCXML interpreter (Go implementation)
- Basic state machine semantics
- Gemini LLM namespace integration
- OpenTelemetry tracing foundation
- Datamodel and event handling

### 🚧 Active Development (Next Few Months)

**Foundation:**
- Event schema validation system (`event:schema` attribute)
- External schema loading (`use:*` pattern)
- Complete W3C SCXML compliance
- Runtime snapshot generation
- IOProcessor interface implementation

**Namespaces:**
- Memory namespace (SQLite, vector search, graph)
- Additional LLM providers (Ollama, Anthropic, OpenAI)
- Stdin/stdout I/O namespace
- HTTP client namespace

### 🔮 Near-Term Goals (6-12 Months)

**Transformers:** The key differentiator for AgentML
- LangGraph transformer (Python)
- CrewAI transformer (Python)
- n8n transformer (JSON workflows)
- OpenAI Agent Builder transformer

**Runtime:**
- agentmlx WASM runtime
- Browser and edge deployment
- Performance optimizations

**Tooling:**
- CLI for transformations and validation
- Development tools and debugging

### 🌟 Long-Term Vision (12+ Months)

**Ecosystem:**
- Visual AgentML editor
- Agent marketplace and sharing
- Additional framework transformers
- Enterprise features and support
- Community governance model

**Philosophy:** We're prioritizing getting the core right before expanding too quickly. Better to have solid foundations than rushed features.

## License

AgentML is part of the AgentFlare ecosystem.

## Contributing

**We're building this in public and contributions are not just welcome—they're essential!**

This is early-stage software with lots of opportunity to shape the direction. Whether you're fixing a typo, proposing a feature, or building a transformer, your input matters.

### How to Contribute

**🐛 Found a bug?** [Open an issue](https://github.com/agentflare-ai/agentml/issues/new) - the more details the better

**💡 Have an idea?** [Start a discussion](https://github.com/agentflare-ai/agentml/discussions/new) - let's talk about it before building

**📝 Improving docs?** PRs welcome - documentation is code too

**🔧 Want to code?** Check out [good first issues](https://github.com/agentflare-ai/agentml/labels/good%20first%20issue)

### What We Need Help With

**High Priority:**
- 🧪 Testing and bug reports - use it, break it, report it
- 📚 Documentation improvements and examples
- 🤔 Use case validation - tell us what you're trying to build
- 🔍 API feedback - what feels right, what feels wrong?

**Exciting Challenges:**
- 🔄 Framework transformer implementations (LangGraph, CrewAI, n8n)
- 🧩 Additional namespace providers (more LLMs, vector DBs, tools)
- ⚡ Runtime optimizations and performance
- 🎨 Developer tooling and CLI utilities
- 📦 Example agents and use cases

**Building Transformers?** We'd love help with:
- LangGraph (Python) transformer
- CrewAI (Python) transformer  
- n8n (JSON workflow) transformer
- Your favorite framework transformer

### Development Philosophy

We value:
- **Transparency** over perfection - show the work in progress
- **Simplicity** over features - keep it focused
- **Standards** over innovation - W3C SCXML is our foundation
- **Community** over speed - get input before building
- **Documentation** as important as code

### Getting Started Contributing

1. **Explore**: Read the docs, try the examples
2. **Discuss**: Share your ideas in Discussions first
3. **Code**: Fork, branch, implement, test
4. **PR**: Submit with clear description and context
5. **Iterate**: Work with maintainers to refine

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines (coming soon).

## Community

**Join us in building the universal agent standard!**

We're building in public and would love to have you involved:

- **📢 GitHub Discussions**: [Share ideas, ask questions, show your work](https://github.com/agentflare-ai/agentml/discussions)
- **🐙 GitHub**: [github.com/agentflare-ai/agentml](https://github.com/agentflare-ai/agentml) - Stars appreciated!
- **💬 Discord**: [Join our community](#) - Real-time chat (coming soon)
- **📖 Documentation**: [docs.agentflare.ai](https://docs.agentflare.ai) - Evolving with the project
- **🔄 Transformers**: [github.com/agentflare-ai/agentml-transformers](https://github.com/agentflare-ai/agentml-transformers) - Framework converters
- **⚡ agentmlx Runtime**: [github.com/agentflare-ai/agentmlx](https://github.com/agentflare-ai/agentmlx) - WASM runtime

### Building in Public Updates

Follow our progress:
- Weekly updates in GitHub Discussions
- Roadmap changes announced in Issues
- Breaking changes clearly communicated
- RFCs for major decisions

We believe in **transparent development** - you'll see the good, the bad, and the bugs. This helps us build something that actually serves the community's needs.

---

## What to Expect

**As an early alpha project**, here's what you should know:

### ✅ What We Promise

- **Honest communication** about what works and what doesn't
- **Responsive to feedback** - your input shapes the project
- **Clear documentation** of current state vs. future vision
- **W3C standards compliance** - building on solid foundations
- **Breaking changes clearly communicated** with migration guides

### ⚠️ What to Expect

- **APIs will change** - we're still finding the best designs
- **Bugs exist** - it's alpha software, expect rough edges
- **Features incomplete** - many things are partially implemented
- **Documentation ahead of code** - we document vision and iterate toward it
- **Timelines are flexible** - priorities shift based on feedback

### 💪 How You Can Help

- **Use it** and tell us what breaks
- **Share your use case** so we can prioritize features
- **Contribute fixes** - even small PRs help
- **Spread the word** - help us find more contributors
- **Be patient** - we're building something ambitious

**This is a marathon, not a sprint.** We're committed to building AgentML right, not just fast.

## Acknowledgments

AgentML builds on the [W3C SCXML specification](https://www.w3.org/TR/scxml/) and is inspired by hierarchical state machine frameworks like XState and Statecharts.

Special thanks to the SCXML working group and the broader agent framework community (LangGraph, CrewAI, n8n, OpenAI, and others) whose work has inspired and informed AgentML's design.

