# Echo Example for AgentML

This example demonstrates an interactive echo agent that uses the `@agentml/stdin` module.

## Files

- `echo.aml` - The main echo agent that loops and echoes user input
- `simple_test.aml` - A simpler test that reads once and exits
- `test_echo.sh` - Test script for manual testing

## Usage

### Interactive Mode

To run the echo agent interactively:

```bash
amlx run echo.aml
```

The agent will:
1. Display `>` prompt
2. Wait for your input
3. Echo back `You said: <your input>`
4. Repeat until you type `exit`

### How It Works

The echo agent uses:
- `import:stdin` attribute to load the stdin namespace module
- `<stdin:read>` element to read from standard input
- Conditional transitions to check for the "exit" command
- A loop structure to continuously prompt for input

## Implementation Details

### Key Components

1. **Datamodel**: Stores the user input in a variable called `input`
2. **Prompt State**: Uses `<stdin:read>` to get input and transitions based on the value
3. **Echo State**: Logs the input using `<log>` element
4. **Exit Condition**: Transitions to final state when input is "exit" or null (EOF)

### State Machine Flow

```
prompt → (check input) → done (if "exit" or EOF)
  ↓                       ↓
  → echo → prompt        (back to prompt if empty)
```

## Validation

The agent passes validation with minor warnings about potential deadlocks, which are acceptable for this use case:

```bash
amlx validate echo.aml
```

## Known Issues

**Current Issue**: The stdin module appears to have a problem where `bufio.Scanner.Scan()` returns immediately without blocking for input when used in the built binary. This needs investigation in the stdin namespace implementation.

### Testing Workaround

For now, you can validate the structure and logic, but interactive testing may not work as expected until the stdin module issue is resolved.

### Next Steps

The stdin module at `/shared/agentflare/agentml/stdin/namespace.go` may need debugging to understand why:
- The scanner returns EOF immediately instead of blocking
- Piped input isn't being read correctly
- The program exits without waiting for user input

This could be related to how os.Stdin is being handled in the generated binary or timing/buffering issues with the scanner.
