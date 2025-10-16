# AgentML

> **🚧 Early Alpha - Building in Public**
>
> AgentML is in early alpha. The vision is ambitious and the foundation is solid, but many features are still in development. Join us in shaping the future of agent standards.

**AgentML is a universal language for AI agents, designed to outlive the churn of frameworks.**

---

## The Vision: HTML for Agents

A new agent framework is released every week. LangGraph, CrewAI, Autogen, n8n—the landscape is fragmented and constantly changing. Choosing the wrong one leads to vendor lock-in and costly rewrites.

AgentML solves this by separating agent *behavior* from the *runtime*.

```
AgentML : Agent Frameworks  =  HTML : Web Browsers
```

You write your agent's logic once in AgentML. Then, you can:

1.  **Run it natively** with **`agentmlx`**, our high-performance Go/WASM runtime. This is the recommended approach for production.
2.  **Transform it** to other popular frameworks. (This is a planned feature).

When a framework becomes obsolete, your agent logic remains portable. No more rewrites.

## How It Works

AgentML is built on [W3C SCXML](https://www.w3.org/TR/scxml/), a battle-tested standard for state machines, and extends it for the AI era.

-   **Define Behavior**: You model your agent as a state machine in an `.aml` file. This provides a deterministic, predictable, and auditable structure.
-   **Guide LLMs with Schemas**: Instead of unpredictable outputs, LLMs generate structured events that are validated against JSON schemas. This makes agent behavior reliable.
-   **Extend with WebAssembly**: Import functionality written in any language (Rust, Go, Python) that compiles to WASM. This enables true polyglot development and secure, sandboxed execution.

### Execution: `agentmlx` Runtime

The primary way to run AgentML is with `agentmlx`, the native reference runtime.

```bash
# Run your agent natively
agentmlx run customer-support.aml
```

`agentmlx` is designed to be a lightweight, high-performance, and portable runtime that can be deployed anywhere—from servers to edge devices.

### Transformation (Planned)

To provide an escape hatch and integrate with existing ecosystems, we are planning to build transformers that convert AgentML to other formats.

```bash
# PLANNED: Transform AgentML to other frameworks
agentmlx transform customer-support.aml --target langgraph --output agent.py
agentmlx transform customer-support.aml --target crewai --output agent.py
```

**This functionality is not yet implemented.** It is a core part of our forward-looking roadmap to ensure AgentML remains a truly universal standard.

---

## Current Status & Roadmap

AgentML is actively being built in public. Here's a snapshot of our progress:

**What's working now:**
- ✅ Core SCXML interpreter (Go implementation)
- ✅ Gemini LLM integration namespace
- ✅ Basic event-driven agent workflows
- ✅ Datamodel and state machine semantics

**What's in active development:**
- 🚧 **`agentmlx` runtime** - Native Go/WASM execution (primary focus)
- 🚧 Memory namespace (vector search, graph database)
- 🚧 Event schema validation and external schema loading
- 🚧 WASM namespace loading via `agentml.wit`

**What's planned:**
- 🔮 **Framework transformers** (LangGraph, CrewAI, n8n, Autogen, etc.)
- 🔮 Visual editor and debugger
- 🔮 Additional LLM provider namespaces (Anthropic, OpenAI, etc.)
- 🔮 Distributed agent communication (IOProcessors)

---

## Get Involved

The vision for a universal agent standard can only succeed with community collaboration.

- **🗣️ Share your use cases** in [GitHub Discussions](https://github.com/agentflare-ai/agentml/discussions).
- **💡 Propose features** via [GitHub Issues](https://github.com/agentflare-ai/agentml/issues).
- **🔧 Contribute code** through pull requests.

---

## A Quick Example

Here is a conceptual look at a simple AgentML file. It defines states, transitions, and uses a namespace (`gemini`) to call an LLM.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       import:gemini="github.com/agentflare-ai/agentml/gemini">

  <datamodel>
    <data id="user_input" expr="''" />
  </datamodel>

  <state id="main">
    <initial>
      <transition target="awaiting_input" />
    </initial>

    <state id="awaiting_input">
      <!-- In a real agent, this would come from an I/O processor -->
      <onentry>
        <assign location="user_input" expr="'Hello, world!'" />
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
      <transition event="action.response" target="responding" />
    </state>

    <state id="responding">
      <onentry>
        <log expr="'LLM responded with: ' + _event.data.message" />
      </onentry>
      <transition target="awaiting_input" />
    </state>
  </state>
</agent>
```