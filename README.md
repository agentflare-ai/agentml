# AgentML

**AgentML** is an agentic workflow framework built on [W3C SCXML](https://www.w3.org/TR/scxml/) that provides deterministic, event-driven state machines for building LLM-powered agents. By combining SCXML's formal state machine semantics with modern LLM capabilities, AgentML enables highly composable, maintainable, and predictable agent behaviors.

**The Universal Agent Standard**: Define your agent once in AgentML and deploy anywhere. Transform to LangGraph, CrewAI, n8n, OpenAI Agent Builder, or run natively with the agentmlx WASM runtime. AgentML unifies the fractured agentic ecosystem, establishing the W3C standard language for agents.

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

- **Deterministic Behavior**: State machines provide predictable, auditable agent behavior
- **Event-Driven Architecture**: LLMs generate structured events validated by JSON schemas
- **Efficient Token Usage**: Runtime snapshots and model caching minimize token consumption
- **Composability**: Invoked services enable modular, reusable agent components
- **Namespace Extensibility**: Plug in custom functionality through SCXML namespaces

### Key Features

- 🎯 **Schema-Guided Events**: `event:schema` attributes validate LLM-generated events
- 📝 **Description-Driven**: Schema `description` fields guide LLMs to generate correct events ⭐
- 🔄 **Runtime Snapshots**: Current state + datamodel provided to LLM for context
- 📦 **Modular Design**: Decompose complex agents into reusable services
- 🔌 **Extensible**: Custom namespaces for memory, LLM providers, I/O, and more
- 📊 **Observable**: Built-in OpenTelemetry tracing and logging
- 🌐 **Universal Standard**: Write once, deploy anywhere - transform to any framework or runtime ⭐
- 🔗 **Remote Communication**: Built-in distributed agent communication via IOProcessors ⭐

> **💡 Critical Success Factor**: Always include detailed `description` fields in your `event:schema` JSON schemas. These descriptions are the primary mechanism for guiding LLMs to generate correct, well-structured events. Both schema-level and property-level descriptions are essential.

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

### JSON Pointer References

AgentML supports **JSON Pointer (RFC 6901)** references for event schemas, enabling schema reuse and cleaner documents:

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
       use:spec="./schemas/events.json"
       use:gemini="github.com/agentflare-ai/agentml/gemini">

  <state id="classify_intent">
    <onentry>
      <gemini:generate location="_event" promptexpr="'Classify: ' + user_input" />
    </onentry>
    
    <!-- Reference schemas using JSON pointers -->
    <transition event="intent.flight"
                event:schema="#/components/schemas/FlightRequest"
                target="handle_flight" />
    
    <transition event="intent.hotel"
                event:schema="#/components/schemas/HotelRequest"
                target="handle_hotel" />
  </state>
  
  <state id="confirm_action">
    <transition event="confirmation.accepted"
                event:schema="#/components/schemas/ConfirmationAccepted"
                target="execute" />
  </state>
</agent>
```

#### Benefits of JSON Pointers

1. **DRY Principle**: Define schemas once, reference everywhere
2. **Maintainability**: Update schema in one place, affects all references
3. **Readability**: Keep `.aml` files clean and focused on behavior
4. **Reusability**: Share schemas across multiple agents
5. **Standards Compliance**: Use OpenAPI 3.x specifications directly
6. **Validation**: External schemas can be validated independently
7. **Versioning**: Version control schemas separately from agents

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
<agent use:spec="https://api.example.com/openapi.yaml">
  <state id="handle_task">
    <transition event="task.request"
                event:schema="#/components/schemas/TaskRequest"
                target="process_task" />
  </state>
</agent>
```

#### Multiple Specification Files

Load multiple specs and reference them:

```xml
<agent use:spec1="./events.json"
       use:spec2="./openapi.yaml"
       use:spec3="https://external-api.com/schema.json">
  
  <!-- Reference schemas from different specs -->
  <transition event="local.event"
              event:schema="spec1#/components/schemas/LocalEvent"
              target="..." />
  
  <transition event="api.event"
              event:schema="spec2#/components/schemas/ApiEvent"
              target="..." />
  
  <transition event="external.event"
              event:schema="spec3#/definitions/ExternalEvent"
              target="..." />
</agent>
```

#### JSON Pointer Syntax

AgentML supports standard JSON Pointer syntax (RFC 6901):

- **`#/components/schemas/MySchema`** - OpenAPI 3.x style
- **`#/definitions/MySchema`** - JSON Schema / OpenAPI 2.0 style
- **`#/properties/myField`** - Direct property reference
- **`spec1#/components/schemas/MySchema`** - Named spec reference

#### Remote Specifications

Load specifications from HTTP(S) URLs:

```xml
<agent use:spec="https://api.github.com/openapi.yaml">
  <!-- Schemas are fetched and cached at agent startup -->
  <transition event="github.webhook"
              event:schema="#/components/schemas/WebhookEvent"
              target="handle_webhook" />
</agent>
```

**Caching Behavior:**
- Specifications loaded once at agent initialization
- Cached for agent lifetime
- Support for ETags and conditional requests
- Optional refresh intervals for long-running agents

#### Schema Composition

Combine inline schemas with references:

```xml
<transition event="custom.event"
            event:schema='{
              "allOf": [
                {"$ref": "#/components/schemas/BaseEvent"},
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
```

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

## Write Once, Deploy Anywhere

AgentML addresses a critical problem in the agentic ecosystem: **framework fragmentation**. Today, agents built for LangGraph won't run on CrewAI, n8n workflows can't be used in OpenAI's Agent Builder, and each framework has its own learning curve and limitations.

### The AgentML Solution

**Define once, deploy everywhere:**

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
agentml transform customer-support.aml --target langgraph --output customer-support.py

# Transform to n8n workflow
agentml transform customer-support.aml --target n8n --output customer-support.json

# Transform to OpenAI Agent Builder
agentml transform customer-support.aml --target openai --output customer-support-config.json
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

## Remote Agent Communication

AgentML includes **built-in support for distributed agent communication** using the W3C SCXML IOProcessor interface. This enables agents to communicate across processes, machines, and networks using standard protocols.

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

### Installation

```bash
go get github.com/agentflare-ai/agentml
```

### Basic Agent Example

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
3. **Use JSON pointers**: Reference schemas from `use:spec` for reusability
4. **Consistent data structure**: Use the same schema across similar events
5. **Const for literals**: Use `{"const": "value"}` for specific values
6. **Version your schemas**: Keep event schemas in version-controlled files

### Datamodel Management

1. **Minimal state**: Store only what's needed in datamodel
2. **Clear naming**: `conversation_history` vs `data1`
3. **Initialize properly**: Set default values in `<datamodel>` section
4. **Type consistency**: ECMAScript datamodel allows flexible types

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
       use:spec="./schemas/events.json"
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
  
  <transition event="intent.hotel"
              event:schema='{
                "type": "object",
                "description": "User intent to search, book, update, or cancel a hotel",
                "properties": {
                  "category": {
                    "const": "hotel",
                    "description": "Category identifier for hotel requests"
                  },
                  "action": {
                    "type": "string",
                    "description": "Action type: search, book, update, or cancel"
                  }
                },
                "required": ["category", "action"]
              }'
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
2. **Use JSON pointers for reusability** ⭐ — Load schemas with `use:spec` and reference with `event:schema="#/path/to/schema"`
3. **Describe every property** — Even simple properties benefit from clear descriptions
4. **Keep prompts minimal** — context comes from runtime snapshot
5. **Use hierarchical states** for agent-lifetime services
6. **Handle errors** — include fallback transitions for unexpected events
7. **Document event flows** — use XML comments to explain complex transitions
8. **Test event schemas** — ensure LLM can generate valid events

### Token Optimization

To minimize token usage:

1. **Store conversation history in datamodel** instead of passing in prompts
2. **Reference datamodel variables** via `promptexpr` instead of copying values
3. **Let runtime generate event lists** — don't list them in prompts
4. **Use schema descriptions** ⭐ — Put ALL documentation in `event:schema` descriptions, not prompts. The runtime sends schemas to the LLM automatically.

**Why Descriptions Matter:**

The runtime automatically provides available events and their schemas to the LLM. By putting comprehensive descriptions in the schema, you:
- Guide the LLM without inflating your prompt
- Ensure consistent event generation across different prompts
- Make schemas self-documenting for both LLMs and humans
- Leverage prompt caching (schemas change rarely)

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

### Current (v0.x)

- ✅ Core SCXML interpreter
- ✅ Agent namespace with event schemas
- ✅ Gemini, Ollama, Memory, Stdin namespaces
- ✅ Runtime snapshots for LLM context
- ✅ Go implementation

### Phase 1: Transformers (Q2 2025)

- 🚧 LangGraph transformer
- 🚧 CrewAI transformer
- 🚧 n8n transformer
- 🚧 OpenAI Agent Builder transformer
- 🚧 CLI tooling for transformations

### Phase 2: Native Runtime (Q3 2025)

- 🔮 agentmlx WASM runtime
- 🔮 Browser support
- 🔮 Edge deployment capabilities
- 🔮 Performance optimizations

### Phase 3: Ecosystem (Q4 2025)

- 🔮 Visual AgentML editor
- 🔮 Agent marketplace
- 🔮 Framework integration SDKs
- 🔮 Enterprise features

### Phase 4: Additional Transformers (2026)

- 🔮 Langflow, Flowise, Dify
- 🔮 Haystack, Semantic Kernel
- 🔮 Custom transformer SDK

## License

AgentML is part of the AgentFlare ecosystem.

## Contributing

Contributions welcome! Please see CONTRIBUTING.md for guidelines.

We're particularly interested in:
- Framework transformer implementations
- Additional namespace providers
- Runtime optimizations
- Documentation improvements
- Example agents and use cases

## Community

- **GitHub**: [github.com/agentflare-ai/agentml](https://github.com/agentflare-ai/agentml)
- **Discord**: [Join our community](#)
- **Documentation**: [docs.agentflare.ai](https://docs.agentflare.ai)
- **Transformers**: [github.com/agentflare-ai/agentml-transformers](https://github.com/agentflare-ai/agentml-transformers)
- **agentmlx Runtime**: [github.com/agentflare-ai/agentmlx](https://github.com/agentflare-ai/agentmlx)

## Acknowledgments

AgentML builds on the [W3C SCXML specification](https://www.w3.org/TR/scxml/) and is inspired by hierarchical state machine frameworks like XState and Statecharts.

Special thanks to the SCXML working group and the broader agent framework community (LangGraph, CrewAI, n8n, OpenAI, and others) whose work has inspired and informed AgentML's design.

