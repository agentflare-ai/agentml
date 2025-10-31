# Customer Support Bot

A simple customer support agent that classifies and handles different types of support requests.

## What It Does

This agent:
1. Receives customer messages
2. Classifies them into categories (billing, technical, account, or general) using AI
3. Routes to the appropriate handler
4. Generates helpful responses using OpenAI's GPT-4
5. Tracks conversation history

## Prerequisites

**REQUIRED:** You must have an OpenAI API key to run this example.

Get your API key from: https://platform.openai.com/api-keys

## How to Run

**Step 1: Set your OpenAI API key**

```bash
export OPENAI_API_KEY=your_api_key_here
```

**Step 2: Run with a customer message**

```bash
agentmlx run support_bot.aml --initial-event '{
  "type": "customer.message",
  "data": {
    "message": "I was charged twice this month"
  }
}'
```

## Example Interactions

### Billing Question

```bash
agentmlx run support_bot.aml --initial-event '{
  "type": "customer.message",
  "data": {
    "message": "I was charged twice this month"
  }
}'
```

### Technical Issue

```bash
agentmlx run support_bot.aml --initial-event '{
  "type": "customer.message",
  "data": {
    "message": "The app keeps crashing when I try to login"
  }
}'
```

### Account Question

```bash
agentmlx run support_bot.aml --initial-event '{
  "type": "customer.message",
  "data": {
    "message": "How do I reset my password?"
  }
}'
```

## How It Works

### State Machine Flow

```
idle → classify_issue → route_to_handler → [billing/technical/account/general] → send_response
  ↑                                                                                    |
  └────────────────────────────────────────────────────────────────────────────────────┘
```

### States Explained

- **idle**: Waiting for customer messages
- **classify_issue**: Uses AI to categorize the support request
- **route_to_handler**: Directs to appropriate handler based on category
- **handle_billing/technical/account/general**: Specialized handlers for each category
- **send_response**: Sends response and waits for follow-up
- **resolved**: Final state when ticket is closed

## Extending This Example

Want to add more features? Try:

1. **Add more categories**: Edit the classification prompt and add new handler states
2. **Add sentiment analysis**: Track if customer is frustrated and escalate
3. **Integrate with ticketing system**: Use I/O processors to send/receive from external APIs
4. **Add canned responses**: Use datamodel to store common responses
5. **Track metrics**: Count response times, resolution rates, etc.

## Testing

Validate the agent:

```bash
agentmlx validate support_bot.aml
```

## Learn More

- [Quick Start Guide](../../docs/quick-start.mdx)
- [State Machines](../../docs/concepts/state-machines.mdx)
- [OpenAI Integration](https://docs.agentml.dev/extensions/openai)
