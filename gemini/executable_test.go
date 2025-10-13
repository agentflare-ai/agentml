package gemini

import (
	"context"
	"strings"
	"testing"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
)

// mockElement provides a minimal implementation of xmldom.Element for testing
type mockElement struct {
	tagName    string
	attributes map[string]string
}

// Element interface methods
func (m *mockElement) TagName() xmldom.DOMString { return xmldom.DOMString(m.tagName) }
func (m *mockElement) GetAttribute(name xmldom.DOMString) xmldom.DOMString {
	return xmldom.DOMString(m.attributes[string(name)])
}
func (m *mockElement) SetAttribute(name, value xmldom.DOMString) error {
	m.attributes[string(name)] = string(value)
	return nil
}
func (m *mockElement) RemoveAttribute(name xmldom.DOMString) error                  { return nil }
func (m *mockElement) GetAttributeNode(name xmldom.DOMString) xmldom.Attr           { return nil }
func (m *mockElement) SetAttributeNode(newAttr xmldom.Attr) (xmldom.Attr, error)    { return nil, nil }
func (m *mockElement) RemoveAttributeNode(oldAttr xmldom.Attr) (xmldom.Attr, error) { return nil, nil }
func (m *mockElement) GetElementsByTagName(name xmldom.DOMString) xmldom.NodeList   { return nil }
func (m *mockElement) GetAttributeNS(namespaceURI, localName xmldom.DOMString) xmldom.DOMString {
	return ""
}
func (m *mockElement) SetAttributeNS(namespaceURI, qualifiedName, value xmldom.DOMString) error {
	return nil
}
func (m *mockElement) RemoveAttributeNS(namespaceURI, localName xmldom.DOMString) error { return nil }
func (m *mockElement) GetAttributeNodeNS(namespaceURI, localName xmldom.DOMString) xmldom.Attr {
	return nil
}
func (m *mockElement) SetAttributeNodeNS(newAttr xmldom.Attr) (xmldom.Attr, error) { return nil, nil }
func (m *mockElement) GetElementsByTagNameNS(namespaceURI, localName xmldom.DOMString) xmldom.NodeList {
	return nil
}
func (m *mockElement) HasAttribute(name xmldom.DOMString) bool                      { return false }
func (m *mockElement) HasAttributeNS(namespaceURI, localName xmldom.DOMString) bool { return false }
func (m *mockElement) ToggleAttribute(name xmldom.DOMString, force ...bool) bool    { return false }
func (m *mockElement) Before(nodes ...xmldom.Node) error                            { return nil }
func (m *mockElement) After(nodes ...xmldom.Node) error                             { return nil }
func (m *mockElement) ReplaceWith(nodes ...xmldom.Node) error                       { return nil }
func (m *mockElement) Remove()                                                      {}
func (m *mockElement) Prepend(nodes ...xmldom.Node) error                           { return nil }
func (m *mockElement) Append(nodes ...xmldom.Node) error                            { return nil }
func (m *mockElement) ChildElementCount() uint32                                    { return 0 }
func (m *mockElement) FirstElementChild() xmldom.Element                            { return nil }
func (m *mockElement) LastElementChild() xmldom.Element                             { return nil }
func (m *mockElement) NextElementSibling() xmldom.Element                           { return nil }
func (m *mockElement) PreviousElementSibling() xmldom.Element                       { return nil }
func (m *mockElement) Children() xmldom.ElementList                                 { return nil }
func (m *mockElement) Contains(other xmldom.Node) bool                              { return false }
func (m *mockElement) GetRootNode() xmldom.Node                                     { return nil }
func (m *mockElement) IsConnected() bool                                            { return false }

// Node interface methods
func (m *mockElement) Position() (line, column int, offset int64) { return 0, 0, 0 }
func (m *mockElement) NodeType() xmldom.NodeType                  { return xmldom.ELEMENT_NODE }
func (m *mockElement) NodeName() xmldom.DOMString                 { return xmldom.DOMString(m.tagName) }
func (m *mockElement) NodeValue() xmldom.DOMString                { return "" }
func (m *mockElement) SetNodeValue(value xmldom.DOMString) error  { return nil }
func (m *mockElement) ParentNode() xmldom.Node                    { return nil }
func (m *mockElement) ChildNodes() xmldom.NodeList                { return nil }
func (m *mockElement) FirstChild() xmldom.Node                    { return nil }
func (m *mockElement) LastChild() xmldom.Node                     { return nil }
func (m *mockElement) PreviousSibling() xmldom.Node               { return nil }
func (m *mockElement) NextSibling() xmldom.Node                   { return nil }
func (m *mockElement) Attributes() xmldom.NamedNodeMap            { return nil }
func (m *mockElement) OwnerDocument() xmldom.Document             { return nil }
func (m *mockElement) InsertBefore(newChild, refChild xmldom.Node) (xmldom.Node, error) {
	return nil, nil
}
func (m *mockElement) ReplaceChild(newChild, oldChild xmldom.Node) (xmldom.Node, error) {
	return nil, nil
}
func (m *mockElement) RemoveChild(oldChild xmldom.Node) (xmldom.Node, error) { return nil, nil }
func (m *mockElement) AppendChild(newChild xmldom.Node) (xmldom.Node, error) { return nil, nil }
func (m *mockElement) HasChildNodes() bool                                   { return false }
func (m *mockElement) CloneNode(deep bool) xmldom.Node                       { return nil }
func (m *mockElement) Normalize()                                            {}
func (m *mockElement) IsSupported(feature, version xmldom.DOMString) bool    { return false }
func (m *mockElement) NamespaceURI() xmldom.DOMString                        { return "" }
func (m *mockElement) Prefix() xmldom.DOMString                              { return "" }
func (m *mockElement) SetPrefix(prefix xmldom.DOMString) error               { return nil }
func (m *mockElement) LocalName() xmldom.DOMString                           { return xmldom.DOMString(m.tagName) }
func (m *mockElement) HasAttributes() bool                                   { return false }
func (m *mockElement) BaseURI() xmldom.DOMString                             { return "" }
func (m *mockElement) CompareDocumentPosition(other xmldom.Node) xmldom.DocumentPositionType {
	return 0
}
func (m *mockElement) TextContent() xmldom.DOMString                               { return "" }
func (m *mockElement) SetTextContent(textContent xmldom.DOMString)                 {}
func (m *mockElement) IsSameNode(other xmldom.Node) bool                           { return false }
func (m *mockElement) LookupPrefix(namespaceURI xmldom.DOMString) xmldom.DOMString { return "" }
func (m *mockElement) IsDefaultNamespace(namespaceURI xmldom.DOMString) bool       { return false }
func (m *mockElement) LookupNamespaceURI(prefix xmldom.DOMString) xmldom.DOMString { return "" }
func (m *mockElement) IsEqualNode(arg xmldom.Node) bool                            { return false }

func newMockElement(tagName string, attributes map[string]string) *mockElement {
	if attributes == nil {
		attributes = make(map[string]string)
	}
	return &mockElement{
		tagName:    tagName,
		attributes: attributes,
	}
}

// Test NewGenerate function
func TestNewGenerate(t *testing.T) {
	t.Run("WithNilElement", func(t *testing.T) {
		ctx := context.Background()
		result, err := NewGenerate(ctx, nil)

		if err == nil {
			t.Errorf("Expected error with nil element")
		}
		if result != nil {
			t.Errorf("Expected nil result with error")
		}
		if err.Error() != "generate element cannot be nil" {
			t.Errorf("Expected specific error message, got: %s", err.Error())
		}
	})

	t.Run("WithValidElement", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "gemini-1.5-flash",
			"prompt":   "test prompt",
			"location": "result",
		})

		result, err := NewGenerate(ctx, element)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Errorf("Expected non-nil result")
		}

		// Type assert to access fields
		generate, ok := result.(*Generate)
		if !ok {
			t.Errorf("Expected *Generate type, got: %T", result)
		}
		if generate.Model != "gemini-1.5-flash" {
			t.Errorf("Expected model 'gemini-1.5-flash', got: %s", generate.Model)
		}
		if generate.Prompt != "test prompt" {
			t.Errorf("Expected prompt 'test prompt', got: %s", generate.Prompt)
		}
		if generate.Location != "result" {
			t.Errorf("Expected location 'result', got: %s", generate.Location)
		}
	})

	t.Run("WithMissingModel", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"prompt":   "test prompt",
			"location": "result",
		})

		result, err := NewGenerate(ctx, element)

		if err == nil {
			t.Errorf("Expected error with missing model")
		}
		if result != nil {
			t.Errorf("Expected nil result with error")
		}
		if !strings.Contains(err.Error(), "model") {
			t.Errorf("Expected error message to mention model, got: %s", err.Error())
		}
	})

	t.Run("WithMissingPrompt", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "gemini-1.5-flash",
			"location": "result",
		})

		result, err := NewGenerate(ctx, element)

		// Prompt is allowed to be empty (can come from child elements)
		if err != nil {
			t.Errorf("Unexpected error with missing prompt: %v", err)
		}
		if result == nil {
			t.Errorf("Expected non-nil result")
		}

		generate, ok := result.(*Generate)
		if !ok {
			t.Errorf("Expected *Generate type")
		} else if generate.Prompt != "" {
			t.Errorf("Expected empty prompt, got: %s", generate.Prompt)
		}
	})

	t.Run("WithMissingLocation", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":  "gemini-1.5-flash",
			"prompt": "test prompt",
		})

		result, err := NewGenerate(ctx, element)

		if err == nil {
			t.Errorf("Expected error with missing location")
		}
		if result != nil {
			t.Errorf("Expected nil result with error")
		}
		if !strings.Contains(err.Error(), "location") {
			t.Errorf("Expected error message to mention location, got: %s", err.Error())
		}
	})
}

// Test Generate.SetClient method
func TestGenerate_SetClient(t *testing.T) {
	generate := &Generate{}
	client := &Client{}

	// Initially no client
	if generate.client != nil {
		t.Errorf("Expected nil client initially")
	}

	// Set client
	generate.SetClient(client)

	// Verify client was set
	if generate.client != client {
		t.Errorf("Expected client to be set")
	}

	// Test setting nil client
	generate.SetClient(nil)
	if generate.client != nil {
		t.Errorf("Expected client to be nil after setting nil")
	}
}

// Test Generate type implements scxml.Executable interface
func TestGenerate_ImplementsInterface(t *testing.T) {
	var _ agentml.Executor = (*Generate)(nil)
	// If this compiles, the interface is implemented correctly
}

// Test Generate struct fields are accessible
func TestGenerate_FieldAccess(t *testing.T) {
	generate := &Generate{
		Model:    "gemini-1.5-flash",
		Prompt:   "test prompt",
		Location: "result",
	}

	if generate.Model != "gemini-1.5-flash" {
		t.Errorf("Expected model 'gemini-1.5-flash', got: %s", generate.Model)
	}
	if generate.Prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got: %s", generate.Prompt)
	}
	if generate.Location != "result" {
		t.Errorf("Expected location 'result', got: %s", generate.Location)
	}
}

// Test that basic template processing doesn't crash
func TestGenerate_processTemplate(t *testing.T) {
	generate := &Generate{}

	t.Run("PlainText", func(t *testing.T) {
		result, err := generate.processTemplate("Hello, World!", nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got: %s", result)
		}
	})

	t.Run("SimpleTemplate", func(t *testing.T) {
		data := map[string]any{"name": "Alice"}
		result, err := generate.processTemplate("Hello, {{.name}}!", data)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "Hello, Alice!" {
			t.Errorf("Expected 'Hello, Alice!', got: %s", result)
		}
	})

	t.Run("NoTemplate", func(t *testing.T) {
		result, err := generate.processTemplate("No template here", nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "No template here" {
			t.Errorf("Expected unchanged text, got: %s", result)
		}
	})
}

// Test error conditions without requiring full SCXML setup
func TestGenerate_ErrorConditions(t *testing.T) {
	t.Run("EmptyModel", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "",
			"prompt":   "test prompt",
			"location": "result",
		})

		result, err := NewGenerate(ctx, element)

		if err == nil {
			t.Errorf("Expected error with empty model")
		}
		if result != nil {
			t.Errorf("Expected nil result with error")
		}
	})

	t.Run("EmptyPrompt", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "gemini-1.5-flash",
			"prompt":   "",
			"location": "result",
		})

		result, err := NewGenerate(ctx, element)

		// Empty prompt should be allowed (can come from child elements)
		if err != nil {
			t.Errorf("Unexpected error with empty prompt: %v", err)
		}
		if result == nil {
			t.Errorf("Expected non-nil result")
		}
	})

	t.Run("EmptyLocation", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "gemini-1.5-flash",
			"prompt":   "test prompt",
			"location": "",
		})

		result, err := NewGenerate(ctx, element)

		if err == nil {
			t.Errorf("Expected error with empty location")
		}
		if result != nil {
			t.Errorf("Expected nil result with error")
		}
	})
}

// Test streaming functionality
func TestGenerate_StreamingAttributes(t *testing.T) {
	t.Run("StreamingEnabled", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "gemini-1.5-flash",
			"prompt":   "test prompt",
			"location": "result",
			"stream":   "true",
			"onchunk":  "chunk",
		})

		result, err := NewGenerate(ctx, element)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Errorf("Expected non-nil result")
		}

		generate, ok := result.(*Generate)
		if !ok {
			t.Errorf("Expected *Generate type")
		}

		if !generate.Stream {
			t.Errorf("Expected Stream to be true")
		}
		if generate.OnChunk != "chunk" {
			t.Errorf("Expected OnChunk to be 'chunk', got: %s", generate.OnChunk)
		}
	})

	t.Run("StreamingDisabled", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "gemini-1.5-flash",
			"prompt":   "test prompt",
			"location": "result",
			"stream":   "false",
		})

		result, err := NewGenerate(ctx, element)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		generate, ok := result.(*Generate)
		if !ok {
			t.Errorf("Expected *Generate type")
		}

		if generate.Stream {
			t.Errorf("Expected Stream to be false")
		}
	})

	t.Run("StreamingDefault", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"model":    "gemini-1.5-flash",
			"prompt":   "test prompt",
			"location": "result",
		})

		result, err := NewGenerate(ctx, element)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		generate, ok := result.(*Generate)
		if !ok {
			t.Errorf("Expected *Generate type")
		}

		if generate.Stream {
			t.Errorf("Expected Stream to default to false")
		}
	})
}

// Test auto-selection functionality
func TestGenerate_AutoSelectAttributes(t *testing.T) {
	t.Run("AutoSelectEnabled", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"prompt":     "test prompt",
			"location":   "result",
			"autoselect": "true",
			"complexity": "moderate",
		})

		result, err := NewGenerate(ctx, element)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		generate, ok := result.(*Generate)
		if !ok {
			t.Errorf("Expected *Generate type")
		}

		if !generate.AutoSelect {
			t.Errorf("Expected AutoSelect to be true")
		}
		if generate.ComplexityHint != "moderate" {
			t.Errorf("Expected ComplexityHint to be 'moderate', got: %s", generate.ComplexityHint)
		}
	})

	t.Run("AutoSelectWithInvalidComplexity", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"prompt":     "test prompt",
			"location":   "result",
			"autoselect": "true",
			"complexity": "invalid",
		})

		result, err := NewGenerate(ctx, element)
		if err == nil {
			t.Errorf("Expected error with invalid complexity hint")
		}
		if result != nil {
			t.Errorf("Expected nil result with error")
		}
		if !strings.Contains(err.Error(), "invalid complexity hint") {
			t.Errorf("Expected complexity hint error, got: %s", err.Error())
		}
	})

	t.Run("AutoSelectWithoutModel", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"prompt":     "test prompt",
			"location":   "result",
			"autoselect": "true",
		})

		result, err := NewGenerate(ctx, element)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		generate, ok := result.(*Generate)
		if !ok {
			t.Errorf("Expected *Generate type")
		}

		if generate.Model != "" {
			t.Errorf("Expected empty model with autoselect")
		}
		if !generate.AutoSelect {
			t.Errorf("Expected AutoSelect to be true")
		}
	})

	t.Run("NoModelNoAutoSelect", func(t *testing.T) {
		ctx := context.Background()
		element := newMockElement("gemini:generate", map[string]string{
			"prompt":   "test prompt",
			"location": "result",
		})

		result, err := NewGenerate(ctx, element)
		if err == nil {
			t.Errorf("Expected error without model or autoselect")
		}
		if result != nil {
			t.Errorf("Expected nil result with error")
		}
	})
}
