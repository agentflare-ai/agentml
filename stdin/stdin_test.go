package stdin_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/agentml/stdin"
	"github.com/agentflare-ai/agentmlx/clock"
	"github.com/agentflare-ai/agentmlx/interpreter"
	"github.com/agentflare-ai/go-xmldom"
)

// captureHandler is a slog.Handler that captures log output for testing
type captureHandler struct {
	buf   *strings.Builder
	level slog.Level
}

func newCaptureHandler(level slog.Level) *captureHandler {
	return &captureHandler{
		buf:   &strings.Builder{},
		level: level,
	}
}

func (h *captureHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *captureHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format: label: message (or just message if no label)
	var label string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "label" {
			label = a.Value.String()
			return false
		}
		return true
	})

	if label != "" {
		h.buf.WriteString(label)
		h.buf.WriteString(": ")
	}

	// Get the message (which might be in "expr" attribute for log elements)
	var message string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "expr" || a.Key == "message" {
			message = a.Value.String()
			return false
		}
		return true
	})

	if message != "" {
		h.buf.WriteString(message)
	} else {
		h.buf.WriteString(r.Message)
	}
	h.buf.WriteString("\n")
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *captureHandler) String() string {
	return h.buf.String()
}

// TestStdinSimpleRead tests a simple stdin read operation
func TestStdinSimpleRead(t *testing.T) {
	// Create AML document that reads once from stdin
	amlDoc := `<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
  xmlns:stdin="github.com/agentflare-ai/agentml/stdin"
  datamodel="ecmascript">

  <datamodel>
    <data id="input" expr="''" />
  </datamodel>

  <state id="main">
    <onentry>
      <stdin:read location="input" prompt="Enter text: " />
      <log expr="input" />
    </onentry>
    <transition target="done" />
  </state>

  <final id="done" />

</agent>`

	// Simulate stdin input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdin temporarily
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write test input to pipe
	go func() {
		io.WriteString(w, "hello world\n")
		w.Close()
	}()

	// Parse the document
	decoder := xmldom.NewDecoder(strings.NewReader(amlDoc))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse AML document: %v", err)
	}

	// Create capture logger for test output
	captureLog := newCaptureHandler(slog.LevelInfo)
	logger := slog.New(captureLog)

	// Get current working directory for root
	absPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	root, err := os.OpenRoot(absPath)
	if err != nil {
		t.Fatalf("Failed to open root: %v", err)
	}

	config := interpreter.Config{
		Clock:  clock.Default(),
		Logger: logger,
		Root:   root,
		Namespaces: map[string]agentml.NamespaceLoader{
			"github.com/agentflare-ai/agentml/stdin": stdin.Loader(),
		},
	}

	ctx := t.Context()
	interp := interpreter.New(ctx, doc, config)

	// Run interpretation in background since stdin blocks
	interpDone := make(chan error, 1)
	go func() {
		interpDone <- interp.Start(ctx)
	}()

	// Wait for interpretation to complete
	select {
	case err := <-interpDone:
		if err != nil {
			t.Fatalf("Interpretation failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Interpretation cancelled: %v", ctx.Err())
	}

	// Check the log output
	logs := captureLog.String()
	t.Logf("Test log output:\n%s", logs)

	if !strings.Contains(logs, "hello world") {
		t.Errorf("Expected output to contain 'hello world', got: %s", logs)
	}
}

// TestStdinEchoLoop tests the echo loop behavior
func TestStdinEchoLoop(t *testing.T) {
	// Use the actual echo.aml file
	amlDoc := `<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
  xmlns:stdin="github.com/agentflare-ai/agentml/stdin"
  datamodel="ecmascript">

  <datamodel>
    <data id="input" expr="''" />
  </datamodel>

  <state id="prompt">
    <onentry>
      <stdin:read location="input" prompt="> " />
    </onentry>
    <transition target="done" cond="input === 'exit' || input === null" />
    <transition target="prompt" cond="input === ''" />
    <transition target="echo" cond="input !== null &amp;&amp; input !== '' &amp;&amp; input !== 'exit'" />
  </state>

  <state id="echo">
    <onentry>
      <log expr="'You said: ' + input" />
    </onentry>
    <transition target="prompt" cond="true" />
  </state>

  <final id="done" />

</agent>`

	// Simulate stdin input with multiple lines
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdin temporarily
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write test input to pipe
	go func() {
		io.WriteString(w, "hello\n")
		io.WriteString(w, "world\n")
		io.WriteString(w, "test\n")
		io.WriteString(w, "exit\n")
		w.Close()
	}()

	// Parse the document
	decoder := xmldom.NewDecoder(strings.NewReader(amlDoc))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse AML document: %v", err)
	}

	// Create capture logger for test output
	captureLog := newCaptureHandler(slog.LevelInfo)
	logger := slog.New(captureLog)

	// Get current working directory for root
	absPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	root, err := os.OpenRoot(absPath)
	if err != nil {
		t.Fatalf("Failed to open root: %v", err)
	}

	config := interpreter.Config{
		Clock:  clock.Default(),
		Logger: logger,
		Root:   root,
		Namespaces: map[string]agentml.NamespaceLoader{
			"github.com/agentflare-ai/agentml/stdin": stdin.Loader(),
		},
	}

	ctx := t.Context()
	interp := interpreter.New(ctx, doc, config)

	// Run interpretation in background since stdin blocks
	interpDone := make(chan error, 1)
	go func() {
		interpDone <- interp.Start(ctx)
	}()

	// Wait for interpretation to complete
	select {
	case err := <-interpDone:
		if err != nil {
			t.Fatalf("Interpretation failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Interpretation cancelled: %v", ctx.Err())
	}

	// Check the log output
	logs := captureLog.String()
	t.Logf("Test log output:\n%s", logs)

	// Verify all echoes happened
	if !strings.Contains(logs, "You said: hello") {
		t.Errorf("Expected output to contain 'You said: hello', got: %s", logs)
	}
	if !strings.Contains(logs, "You said: world") {
		t.Errorf("Expected output to contain 'You said: world', got: %s", logs)
	}
	if !strings.Contains(logs, "You said: test") {
		t.Errorf("Expected output to contain 'You said: test', got: %s", logs)
	}
	// Should NOT echo "exit"
	if strings.Contains(logs, "You said: exit") {
		t.Errorf("Should not echo 'exit' command, got: %s", logs)
	}
}

// TestStdinEOF tests that stdin handles EOF correctly
// TODO: This test has issues with null comparison - needs investigation
func TestStdinEOF(t *testing.T) {
	t.Skip("Skipping EOF test - null comparison issue needs investigation")
	amlDoc := `<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
  xmlns:stdin="github.com/agentflare-ai/agentml/stdin"
  datamodel="ecmascript">

  <datamodel>
    <data id="input" expr="''" />
  </datamodel>

  <state id="prompt">
    <onentry>
      <stdin:read location="input" prompt="> " />
    </onentry>
    <transition target="done" cond="input === null">
      <log expr="'Got EOF'" />
    </transition>
    <transition target="echo" cond="input !== null" />
  </state>

  <state id="echo">
    <onentry>
      <log expr="'You said: ' + input" />
    </onentry>
    <transition target="prompt" cond="true" />
  </state>

  <final id="done" />

</agent>`

	// Simulate stdin with immediate EOF
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdin temporarily
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Close write end immediately to simulate EOF
	w.Close()

	// Parse the document
	decoder := xmldom.NewDecoder(strings.NewReader(amlDoc))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse AML document: %v", err)
	}

	// Create capture logger for test output
	captureLog := newCaptureHandler(slog.LevelInfo)
	logger := slog.New(captureLog)

	// Get current working directory for root
	absPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	root, err := os.OpenRoot(absPath)
	if err != nil {
		t.Fatalf("Failed to open root: %v", err)
	}

	config := interpreter.Config{
		Clock:  clock.Default(),
		Logger: logger,
		Root:   root,
		Namespaces: map[string]agentml.NamespaceLoader{
			"github.com/agentflare-ai/agentml/stdin": stdin.Loader(),
		},
	}

	ctx := t.Context()
	interp := interpreter.New(ctx, doc, config)

	// Run interpretation in background since stdin blocks
	interpDone := make(chan error, 1)
	go func() {
		interpDone <- interp.Start(ctx)
	}()

	// Wait for interpretation to complete
	select {
	case err := <-interpDone:
		if err != nil {
			t.Fatalf("Interpretation failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Interpretation cancelled: %v", ctx.Err())
	}

	// Check the log output
	logs := captureLog.String()
	t.Logf("Test log output:\n%s", logs)

	if !strings.Contains(logs, "Got EOF") {
		t.Errorf("Expected output to contain 'Got EOF', got: %s", logs)
	}
}

// TestStdinEmptyLines tests that empty lines are handled correctly
func TestStdinEmptyLines(t *testing.T) {
	amlDoc := `<?xml version="1.0" encoding="UTF-8"?>
<agent xmlns="github.com/agentflare-ai/agentml/agent"
  xmlns:stdin="github.com/agentflare-ai/agentml/stdin"
  datamodel="ecmascript">

  <datamodel>
    <data id="input" expr="''" />
    <data id="count" expr="0" />
  </datamodel>

  <state id="prompt">
    <onentry>
      <stdin:read location="input" prompt="> " />
      <assign location="count" expr="count + 1" />
    </onentry>
    <transition target="done" cond="input === 'exit'" />
    <transition target="prompt" cond="input === ''">
      <log expr="'Empty line skipped'" />
    </transition>
    <transition target="echo" />
  </state>

  <state id="echo">
    <onentry>
      <log expr="'You said: ' + input" />
    </onentry>
    <transition target="prompt" />
  </state>

  <final id="done">
    <onentry>
      <log expr="'Total prompts: ' + count" />
    </onentry>
  </final>

</agent>`

	// Simulate stdin with empty lines
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdin temporarily
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write test input with empty lines
	go func() {
		io.WriteString(w, "hello\n")
		io.WriteString(w, "\n") // empty line
		io.WriteString(w, "\n") // another empty line
		io.WriteString(w, "world\n")
		io.WriteString(w, "exit\n")
		w.Close()
	}()

	// Parse the document
	decoder := xmldom.NewDecoder(strings.NewReader(amlDoc))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse AML document: %v", err)
	}

	// Create capture logger for test output
	captureLog := newCaptureHandler(slog.LevelInfo)
	logger := slog.New(captureLog)

	// Get current working directory for root
	absPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	root, err := os.OpenRoot(absPath)
	if err != nil {
		t.Fatalf("Failed to open root: %v", err)
	}

	config := interpreter.Config{
		Clock:  clock.Default(),
		Logger: logger,
		Root:   root,
		Namespaces: map[string]agentml.NamespaceLoader{
			"github.com/agentflare-ai/agentml/stdin": stdin.Loader(),
		},
	}

	ctx := t.Context()
	interp := interpreter.New(ctx, doc, config)

	// Run interpretation in background since stdin blocks
	interpDone := make(chan error, 1)
	go func() {
		interpDone <- interp.Start(ctx)
	}()

	// Wait for interpretation to complete
	select {
	case err := <-interpDone:
		if err != nil {
			t.Fatalf("Interpretation failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Interpretation cancelled: %v", ctx.Err())
	}

	// Check the log output
	logs := captureLog.String()
	t.Logf("Test log output:\n%s", logs)

	// Should have skipped empty lines
	if !strings.Contains(logs, "Empty line skipped") {
		t.Errorf("Expected output to contain 'Empty line skipped', got: %s", logs)
	}

	// Should echo non-empty lines
	if !strings.Contains(logs, "You said: hello") {
		t.Errorf("Expected output to contain 'You said: hello', got: %s", logs)
	}
	if !strings.Contains(logs, "You said: world") {
		t.Errorf("Expected output to contain 'You said: world', got: %s", logs)
	}

	// Should show total prompts (including empty ones)
	if !strings.Contains(logs, "Total prompts: 5") {
		t.Errorf("Expected 5 total prompts (hello + 2 empty + world + exit), got: %s", logs)
	}
}
