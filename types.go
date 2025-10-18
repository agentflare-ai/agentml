package agentml

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/agentflare-ai/go-xmldom"
)

const NamespaceURI = "https://xsd.agentml.dev/agentml"

// Data represents a data element defined in the SCXML document.
type Data struct {
	xmldom.Element
	ID      string // The data element identifier
	Expr    string // Optional initial value expression
	Src     string // Optional external source URI
	Content any    // Optional XML content (for XPath data model)
}

// Param represents a parameter for data operations (SCXML 5.7)
type Param struct {
	xmldom.Element
	Name     string // The name of the parameter key
	Expr     string // Optional value expression to evaluate
	Location string // Optional location expression to retrieve value from
}

// Script provides scripting capabilities (SCXML 5.8)
type Script struct {
	xmldom.Element
	Src     string // Optional URI of external script to load
	Content string // Inline script content
}

// If provides conditional execution with elseif and else branches (SCXML 4.3)
type If struct {
	xmldom.Element
	Cond        string      // Boolean condition expression
	Interpreter Interpreter // Reference to interpreter for recursive execution
}

// Foreach iterates over a collection in the data model (SCXML 4.6)
type Foreach struct {
	xmldom.Element
	Array string // Value expression that evaluates to an iterable collection
	Item  string // Variable name to store each item during iteration
	Index string // Optional variable name to store iteration index
}

// Assign changes the value of a location in the data model (SCXML 5.4)
type Assign struct {
	xmldom.Element
	Location    string        // Location expression specifying where to assign the value
	Expr        string        // Optional value expression to evaluate and assign
	AssignType  string        // Optional type attribute for XML handling modes
	InlineNodes []xmldom.Node // Inline XML/text content if no expr
	Content     string        // Text content fallback
}

// Log generates a logging or debug message (SCXML 5.11)
type Log struct {
	xmldom.Element
	Label string // Optional label for the log message
	Expr  string // Expression to evaluate and log
}

// Content represents content for data operations (SCXML 5.6)
type Content struct {
	xmldom.Element
	Expr string // Optional value expression to evaluate
	Body any    // Optional inline content body (XML, text, etc.)
}

// SendData encapsulates all data needed for a send operation
type SendData struct {
	Event    string   // Event name to send
	Target   string   // Target URI for the message
	Type     string   // I/O processor type URI
	ID       string   // Send identifier
	Delay    string   // Delay duration (CSS2 format)
	NameList []string // List of data model locations to include
	Params   []Param  // Parameter key-value pairs
	Content  *Content // Content payload (nil if not present)
}

// Send sends an event to a specified destination (SCXML 6.2)
type Send struct {
	xmldom.Element
	Event      string   // Optional event name to send
	EventExpr  string   // Optional dynamic event name expression
	Target     string   // Optional target URI
	TargetExpr string   // Optional dynamic target expression
	TypeURI    string   // Optional I/O processor type URI
	TypeExpr   string   // Optional dynamic type expression
	SendID     string   // Optional send identifier
	IdLocation string   // Optional location to store generated ID
	Delay      string   // Optional delay duration (CSS2 format)
	DelayExpr  string   // Optional dynamic delay expression
	NameList   []string // Optional list of data model locations to include
	Params     []Param  // Optional parameter key-value pairs
	Content    *Content // Optional content payload
}

// Cancel cancels a previously sent event (SCXML 6.3)
type Cancel struct {
	xmldom.Element
	SendID     string // The ID of the send element to cancel
	SendIDExpr string // Optional expression to compute the send ID
}

// Raise raises an internal event (SCXML 6.4)
type Raise struct {
	xmldom.Element
	Event     string // Event name to raise
	EventExpr string // Optional dynamic event name expression
}

const (
	EventSystemVariable        = "_event"
	SessionIDSystemVariable    = "_sessionid"
	NameSystemVariable         = "_name"
	IOProcessorsSystemVariable = "_ioprocessors"
	XSystemVariable            = "_x"
)

type Variable struct {
	Name string
}

// EventType represents the type of SCXML event
type EventType string

const (
	EventTypeInternal EventType = "internal"
	EventTypeExternal EventType = "external"
	EventTypePlatform EventType = "platform"
)

// Event represents an SCXML event as defined in the W3C specification
type Event struct {
	ID         string    `json:"id"`                   // Unique event ID using MUID
	Name       string    `json:"name"`                 // Event name for matching
	Type       EventType `json:"type"`                 // Internal, external, or platform
	Delay      string    `json:"delay,omitempty"`      // Delay for delayed events
	Data       any       `json:"data"`                 // Event data payload
	Metadata   any       `json:"metadata,omitempty"`   // Metadata for the event
	InvokeID   string    `json:"invokeid,omitempty"`   // For invoked sessions
	Timestamp  time.Time `json:"timestamp"`            // When event was created
	Origin     string    `json:"origin,omitempty"`     // Origin of external events
	OriginType string    `json:"origintype,omitempty"` // Type of origin
	SendID     string    `json:"sendid,omitempty"`     // ID from send element
	Raw        string    `json:"raw,omitempty"`        // Raw data for HTTP events
	Target     string    `json:"target,omitempty"`     // Target URI from original send
	TargetType string    `json:"targettype,omitempty"` // I/O processor type URI from send
}

// Finalize finalizes the session (SCXML 6.5)
type Finalize struct {
	xmldom.Element
}

// IOProcessor defines the interface that all I/O processors must implement
// according to W3C SCXML specification sections C.1 and C.2
type IOProcessor interface {
	// Handle processes a fully-formed event using this I/O processor
	// This is the preferred method for I/O processor implementations.
	// The event contains all pre-evaluated data from the interpreter's data model.
	// IOProcessors should focus only on transport/communication logic.
	// ctx: context for tracing and cancellation
	// event: the event to handle (all data model evaluation already completed)
	// Returns error if transport/communication fails (e.g., error.communication)
	Handle(ctx context.Context, event *Event) error

	// Location returns the location/URI that external entities can use
	// to communicate with this SCXML session via this I/O processor
	// This is used to populate the _ioprocessors system variable
	Location(ctx context.Context) (string, error)

	// Type returns the I/O processor type URI (e.g., "github.com/agentflare-ai/agentml/ioprocessor/scxml")
	Type() string

	// Shutdown cleans up resources used by this I/O processor
	Shutdown(ctx context.Context) error
}

// Executor represents any executable content that can be executed
type Executor interface {
	xmldom.Element
	// Execute runs the executable content
	Execute(ctx context.Context, interpreter Interpreter) error
}

type NamespaceLoader func(ctx context.Context, interpreter Interpreter, doc xmldom.Document) (Namespace, error)

type Namespace interface {
	URI() string
	Handle(ctx context.Context, element xmldom.Element) (bool, error)
	Unload(ctx context.Context) error
}

type DataModel interface {
	// Initialize sets up the data model with initial data elements.
	// This is called when the SCXML document is loaded and should create
	// all data elements defined in <data> elements.
	Initialize(ctx context.Context, dataElements []Data) error

	// EvaluateValue evaluates a value expression and returns the result.
	// Used for <data expr="...">, <assign expr="...">, etc.
	// Returns an error if the expression cannot be evaluated.
	EvaluateValue(ctx context.Context, expression string) (any, error)

	// EvaluateCondition evaluates a conditional expression and returns a boolean.
	// Used for <transition cond="..."> and other conditional logic.
	// Returns an error if the expression cannot be evaluated as a boolean.
	EvaluateCondition(ctx context.Context, expression string) (bool, error)

	// EvaluateLocation evaluates a location expression and returns the value at that location.
	// Used for <param location="..."> and other location-based access.
	// Returns an error if the location is invalid or cannot be accessed.
	EvaluateLocation(ctx context.Context, location string) (any, error)

	// Assign assigns a value to a location in the data model.
	// Used for <assign location="..." expr="..."> operations.
	// Returns an error if the location is invalid or the assignment fails.
	Assign(ctx context.Context, location string, value any) error

	// GetVariable retrieves the value of a data element by ID.
	// Returns an error if the data element doesn't exist.
	GetVariable(ctx context.Context, id string) (any, error)

	// SetVariable sets the value of a data element by ID.
	// Returns an error if the data element doesn't exist or the value is invalid.
	SetVariable(ctx context.Context, id string, value any) error

	// GetSystemVariable retrieves a system variable value.
	// System variables include: _event, _sessionid, _name, _ioprocessors, _x
	// Returns an error if the system variable doesn't exist.
	GetSystemVariable(ctx context.Context, name string) (any, error)

	// SetSystemVariable sets a system variable value.
	// Most system variables are read-only and will return an error if modified.
	// Returns an error if the variable doesn't exist or cannot be modified.
	SetSystemVariable(ctx context.Context, name string, value any) error

	// SetCurrentEvent sets the _event system variable to the current event.
	// This is called by the interpreter when processing events.
	SetCurrentEvent(ctx context.Context, event any) error

	// ExecuteScript executes a script in the data model's context.
	// For ECMAScript data models, this executes JavaScript code with access to all variables.
	// For other data models, this may evaluate the script as an expression.
	// Returns an error if the script cannot be executed.
	ExecuteScript(ctx context.Context, script string) error

	// Clone creates a copy of the data model for use in parallel states.
	// The clone should share system variables but have independent data elements.
	Clone(ctx context.Context) (DataModel, error)

	// ValidateExpression validates that an expression is syntactically correct
	// for this data model. Returns nil if valid, error if invalid.
	ValidateExpression(ctx context.Context, expression string, exprType ExpressionType) error
}

// ExpressionType defines the type of expression being evaluated.
type ExpressionType string

const (
	ValueExpression     ExpressionType = "value"
	ConditionExpression ExpressionType = "condition"
	LocationExpression  ExpressionType = "location"
)

// ExecutionError represents an error that occurred during agentml execution
type ExecutionError struct {
	Message string
	Element xmldom.Element
}

func (e *ExecutionError) Error() string {
	line, column, _ := e.Element.Position()
	return fmt.Sprintf("Execution error: %s in %s at %d:%d", e.Message, e.Element.TagName(), line, column)
}

var _ error = (*ExecutionError)(nil)

// PlatformError represents an error that should generate a platform error event
type PlatformError struct {
	EventName string         // The error event name (e.g., "error.execution")
	Message   string         // Error message
	Data      map[string]any // Additional error data (element, line, etc.)
	Cause     error          // Wrapped underlying error
}

func (e *PlatformError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *PlatformError) Unwrap() error {
	return e.Cause
}

var _ error = (*PlatformError)(nil)

// Clock provides an abstraction over time for testing and simulation
type Clock interface {
	// Now returns the current time
	Now() time.Time

	// Since returns the duration since the given time
	Since(t time.Time) time.Duration

	// Sleep pauses the current goroutine for at least the given duration
	Sleep(ctx context.Context, d time.Duration) error

	// After returns a channel that receives the current time after the given duration
	After(d time.Duration) <-chan time.Time

	// NewTimer creates a new timer that will send the current time after the given duration
	NewTimer(d time.Duration) Timer

	// NewTicker creates a new ticker that will send the current time every given duration
	NewTicker(d time.Duration) Ticker

	// TimeScale returns the current time scale (1.0 = real-time, 2.0 = 2x speed, etc.)
	TimeScale() float64

	// SetTimeScale sets the time scale for simulation (only applies to simulation clocks)
	SetTimeScale(scale float64)

	// Advance manually advances time by the given duration (only applies to mock clocks)
	Advance(d time.Duration)

	// Pause pauses time advancement (only applies to mock clocks)
	Pause()

	// Resume resumes time advancement (only applies to mock clocks)
	Resume()

	// IsPaused returns true if the clock is paused
	IsPaused() bool
}

// Timer interface abstracts time.Timer
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

// Ticker interface abstracts time.Ticker
type Ticker interface {
	C() <-chan time.Time
	Stop()
	Reset(d time.Duration)
}

type IOProcessorLoader func(ctx context.Context, interpreter Interpreter) (IOProcessor, error)

type DataModelLoader func(ctx context.Context, interpreter Interpreter) (DataModel, error)

// SnapshotConfig controls what the snapshot excludes when embedding into the document.
// By default, all sections are included. Use Exclude fields to opt-out of specific sections.
type SnapshotConfig struct {
	// ExcludeAll acts as a master switch: when true, disables all sections
	ExcludeAll bool
	// Specific sections to exclude (opt-out pattern)
	ExcludeConfiguration bool // exclude state configuration
	ExcludeData          bool // exclude datamodel values
	ExcludeQueue         bool // exclude internal/external queues
	ExcludeServices      bool // exclude invoked child services recursively
	ExcludeRaise         bool // exclude available raise (internal) transitions
	ExcludeSend          bool // exclude available send (external) transitions
	ExcludeCancel        bool // exclude cancelable delayed events
}

// Interpreter interface for SCXML interpretation
type Interpreter interface {
	IOProcessor
	SessionID() string
	Configuration() []string
	In(ctx context.Context, stateId string) bool
	Raise(ctx context.Context, event *Event)
	Send(ctx context.Context, event *Event) error
	Cancel(ctx context.Context, sendId string) error
	Log(ctx context.Context, label, message string)
	Context() context.Context
	Clock() Clock
	DataModel() DataModel
	ExecuteElement(ctx context.Context, element xmldom.Element) error
	SendMessage(ctx context.Context, data SendData) error
	ScheduleMessage(ctx context.Context, data SendData) (string, error)
	InvokedSessions() map[string]Interpreter
	Tracer() Tracer
	Snapshot(ctx context.Context, maybeConfig ...SnapshotConfig) (xmldom.Document, error)
}

// Position contains source position information for a diagnostic
type Position struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Offset int64  `json:"offset"`
}

// Trace describes an issue found during validation or runtime execution
type Trace struct {
	Level     slog.Level `json:"level"`
	Code      string     `json:"code"`
	Message   string     `json:"message"`
	Position  Position   `json:"position"`
	Tag       string     `json:"tag,omitempty"`
	Attribute string     `json:"attribute,omitempty"`
	Hints     []string   `json:"hints,omitempty"`
}

// Option for adding extra context to diagnostics

type Option func(*Trace)

// Tracer interface for collecting diagnostics
type Tracer interface {
	Error(code, message string, element xmldom.Element, opts ...Option)
	Warn(code, message string, element xmldom.Element, opts ...Option)
	Info(code, message string, element xmldom.Element, opts ...Option)

	Diagnostics() []Trace
	HasErrors() bool
	Clear()
}
