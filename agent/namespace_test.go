package agent

import (
	"context"
	"testing"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
)

func TestNamespaceURI(t *testing.T) {
	loader := Loader()
	ns, err := loader(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Failed to load namespace: %v", err)
	}

	expectedURI := "github.com/agentflare-ai/agentml/agent"
	if ns.URI() != expectedURI {
		t.Errorf("Expected URI %q, got %q", expectedURI, ns.URI())
	}
}

func TestHandle_NonAgentElement(t *testing.T) {
	loader := Loader()
	ns, err := loader(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Failed to load namespace: %v", err)
	}

	// Create a mock element with a different local name
	mockEl := &mockElement{localName: "other"}

	handled, err := ns.Handle(context.Background(), mockEl)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if handled {
		t.Error("Expected element not to be handled")
	}
}

func TestHandle_NilElement(t *testing.T) {
	loader := Loader()
	ns, err := loader(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Failed to load namespace: %v", err)
	}

	handled, err := ns.Handle(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil element")
	}
	if handled {
		t.Error("Expected element not to be handled")
	}
}

func TestHandle_TransitionWithEventSchema(t *testing.T) {
	// Create a mock interpreter
	mockInterp := &mockInterpreter{logs: make([]string, 0)}

	loader := Loader()
	ns, err := loader(context.Background(), mockInterp, nil)
	if err != nil {
		t.Fatalf("Failed to load namespace: %v", err)
	}

	// Create a mock transition element with event:schema
	mockTrans := &mockTransitionElement{
		localName:   "transition",
		eventSchema: `{"type": "object", "properties": {"text": {"type": "string"}}}`,
	}

	handled, err := ns.Handle(context.Background(), mockTrans)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if handled {
		t.Error("Expected transition not to be handled (should be processed by SCXML handler)")
	}

	// Check that the schema was logged
	if len(mockInterp.logs) == 0 {
		t.Error("Expected schema to be logged")
	}
}

// Mock element for testing
type mockElement struct {
	xmldom.Element
	localName string
}

func (m *mockElement) LocalName() xmldom.DOMString {
	return xmldom.DOMString(m.localName)
}

func (m *mockElement) Attributes() xmldom.NamedNodeMap {
	return &mockNamedNodeMap{}
}

func (m *mockElement) ChildNodes() xmldom.NodeList {
	return &mockNodeList{}
}

func (m *mockElement) GetAttribute(name xmldom.DOMString) xmldom.DOMString {
	return ""
}

// Mock NamedNodeMap
type mockNamedNodeMap struct {
	xmldom.NamedNodeMap
}

func (m *mockNamedNodeMap) Length() uint {
	return 0
}

func (m *mockNamedNodeMap) Item(index uint) xmldom.Node {
	return nil
}

// Mock NodeList
type mockNodeList struct {
	xmldom.NodeList
}

func (m *mockNodeList) Length() uint {
	return 0
}

func (m *mockNodeList) Item(index uint) xmldom.Node {
	return nil
}

// Mock transition element with event:schema
type mockTransitionElement struct {
	xmldom.Element
	localName   string
	eventSchema string
}

func (m *mockTransitionElement) LocalName() xmldom.DOMString {
	return xmldom.DOMString(m.localName)
}

func (m *mockTransitionElement) GetAttribute(name xmldom.DOMString) xmldom.DOMString {
	if string(name) == "event:schema" {
		return xmldom.DOMString(m.eventSchema)
	}
	return ""
}

// Mock interpreter for testing
type mockInterpreter struct {
	agentml.Interpreter
	logs []string
}

func (m *mockInterpreter) Log(ctx context.Context, label, message string) {
	m.logs = append(m.logs, message)
}

var _ agentml.Namespace = (*ns)(nil)
