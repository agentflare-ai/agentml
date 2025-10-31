# Flight Booking Agent

A conversational agent that helps users search for and book flights through a natural dialogue.

## What It Does

This agent guides users through the flight booking process:
1. Collects travel details (from, to, dates, passengers, class)
2. Searches for flights using OpenAI's GPT-4
3. Presents flight options
4. Confirms booking with confirmation number

## Prerequisites

**REQUIRED:** You must have an OpenAI API key to run this example.

Get your API key from: https://platform.openai.com/api-keys

## How to Run

**Step 1: Set your OpenAI API key**

```bash
export OPENAI_API_KEY=your_api_key_here
```

**Step 2: Run the agent**

### Interactive Mode (Recommended)

For a conversational experience, use the stdin module:

```bash
agentmlx run booking_agent.aml --initial-event '{"type":"user.input","data":{"message":"start"}}'
```

Then follow the prompts!

### Single Request Mode

Book a flight with one command:

```bash
agentmlx run booking_agent.aml --initial-event '{
  "type": "user.input",
  "data": {
    "message": "New York"
  }
}'
```

## Example Conversation

```
Agent: Welcome! I can help you book a flight. Where would you like to fly from?
User: New York
Agent: Great! Where would you like to fly to?
User: Los Angeles
Agent: When would you like to depart? (e.g., 2025-12-15)
User: 2025-12-15
Agent: When would you like to return? (e.g., 2025-12-20 or say one-way)
User: 2025-12-20
Agent: How many passengers? (1-9)
User: 2
Agent: Which class would you prefer? (economy, business, first)
User: economy
Agent: Searching for flights from New York to Los Angeles...
Agent: Here are your flight options:
[Flight options displayed]
Agent: Would you like to book one of these flights? (yes/no)
User: yes
Agent: [Booking confirmation with confirmation number]
```

## How It Works

### State Machine Flow

```
welcome → get_destination → get_departure_date → get_return_date
   ↓
get_passengers → get_class → search_flights → present_options
                                                       ↓
                              modify_search ← (no) ← confirm_booking → booking_complete
                                   ↓
                              welcome (start over)
```

### States Explained

- **welcome**: Greets user and asks for departure city
- **get_destination**: Collects arrival city
- **get_departure_date**: Gets departure date
- **get_return_date**: Gets return date (or one-way)
- **get_passengers**: Collects number of passengers
- **get_class**: Gets flight class preference
- **search_flights**: Uses AI to search for flight options
- **present_options**: Shows available flights
- **confirm_booking**: Creates booking confirmation
- **booking_complete**: Final state with confirmation details
- **modify_search**: Allows user to start over
- **search_error**: Handles search failures

## Key Features

### Data Collection
The agent systematically collects all required booking information through a conversational flow.

### AI-Powered Search
Uses Gemini AI to generate realistic flight options based on search criteria.

### Error Handling
Includes error states to handle search failures gracefully.

### Flexible Flow
Users can modify their search or start over at any point.

## Extending This Example

Want to add more features? Try:

1. **Real flight API integration**: Connect to actual flight search APIs using I/O processors
2. **Price comparison**: Compare prices across multiple airlines
3. **Seat selection**: Add seat map and selection
4. **Payment processing**: Integrate payment gateway
5. **Email confirmation**: Send booking confirmation via email
6. **Calendar integration**: Add to user's calendar
7. **Multi-city trips**: Support complex itineraries

## Testing

Validate the agent:

```bash
agentmlx validate booking_agent.aml
```

Test with different scenarios:

```bash
# One-way flight
agentmlx run booking_agent.aml --initial-event '{"type":"user.input","data":{"message":"Boston"}}'

# Business class
agentmlx run booking_agent.aml --initial-event '{"type":"user.input","data":{"message":"San Francisco"}}'

# Multiple passengers
agentmlx run booking_agent.aml --initial-event '{"type":"user.input","data":{"message":"Chicago"}}'
```

## Learn More

- [Quick Start Guide](../../docs/quick-start.mdx)
- [State Machines](../../docs/concepts/state-machines.mdx)
- [OpenAI Integration](https://docs.agentml.dev/extensions/openai)
- [I/O Processors](../../docs/architecture/io-processors.mdx) - For real API integration
