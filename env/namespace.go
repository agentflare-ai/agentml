package env

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const NamespaceURI = "github.com/agentflare-ai/agentml/env"

type Namespace struct {
	itp agentml.Interpreter
}

func (n *Namespace) URI() string { return NamespaceURI }

func (n *Namespace) Unload(ctx context.Context) error { return nil }

func (n *Namespace) Handle(ctx context.Context, el xmldom.Element) (bool, error) {
	if el == nil {
		return false, fmt.Errorf("env: element cannot be nil")
	}
	local := strings.ToLower(string(el.LocalName()))
	switch local {
	case "get":
		return true, n.execGet(ctx, el)
	case "set":
		return true, n.execSet(ctx, el)
	default:
		return false, nil
	}
}

func (n *Namespace) execGet(ctx context.Context, el xmldom.Element) error {
	tr := otel.Tracer("env")
	ctx, span := tr.Start(ctx, "env.get")
	defer span.End()

	dm := n.itp.DataModel()
	if dm == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for env",
			Data:      map[string]any{"element": "get"},
			Cause:     fmt.Errorf("no datamodel"),
		}
	}

	// Get the name attribute (required)
	name := strings.TrimSpace(string(el.GetAttribute("name")))
	if name == "" {
		// Try nameexpr if name is not provided
		nameExpr := strings.TrimSpace(string(el.GetAttribute("nameexpr")))
		if nameExpr != "" {
			val, err := dm.EvaluateValue(ctx, nameExpr)
			if err != nil {
				return &agentml.PlatformError{
					EventName: "error.execution",
					Message:   "Failed to evaluate nameexpr",
					Data:      map[string]any{"element": "get", "nameexpr": nameExpr},
					Cause:     err,
				}
			}
			if s, ok := val.(string); ok {
				name = s
			}
		}
	}

	if name == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "env:get requires name or nameexpr attribute",
			Data:      map[string]any{"element": "get"},
			Cause:     fmt.Errorf("missing name"),
		}
	}

	span.SetAttributes(attribute.String("env.name", name))

	// Get the location attribute where we should store the result (required)
	loc := strings.TrimSpace(string(el.GetAttribute("location")))
	if loc == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "env:get requires location attribute",
			Data:      map[string]any{"element": "get", "name": name},
			Cause:     fmt.Errorf("missing location"),
		}
	}

	// Get optional default value
	defaultVal := string(el.GetAttribute("default"))

	// Read environment variable
	value, exists := os.LookupEnv(name)
	if !exists {
		if defaultVal != "" {
			value = defaultVal
			span.SetAttributes(attribute.Bool("env.used_default", true))
		}
		// If no default and variable doesn't exist, store empty string
	}

	span.SetAttributes(attribute.Bool("env.exists", exists))

	// Store the result in the data model
	if err := dm.SetVariable(ctx, loc, value); err != nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Failed to store environment variable",
			Data:      map[string]any{"element": "get", "name": name, "location": loc},
			Cause:     err,
		}
	}

	return nil
}

func (n *Namespace) execSet(ctx context.Context, el xmldom.Element) error {
	tr := otel.Tracer("env")
	ctx, span := tr.Start(ctx, "env.set")
	defer span.End()

	dm := n.itp.DataModel()
	if dm == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for env",
			Data:      map[string]any{"element": "set"},
			Cause:     fmt.Errorf("no datamodel"),
		}
	}

	// Get the name attribute (required)
	name := strings.TrimSpace(string(el.GetAttribute("name")))
	if name == "" {
		// Try nameexpr if name is not provided
		nameExpr := strings.TrimSpace(string(el.GetAttribute("nameexpr")))
		if nameExpr != "" {
			val, err := dm.EvaluateValue(ctx, nameExpr)
			if err != nil {
				return &agentml.PlatformError{
					EventName: "error.execution",
					Message:   "Failed to evaluate nameexpr",
					Data:      map[string]any{"element": "set", "nameexpr": nameExpr},
					Cause:     err,
				}
			}
			if s, ok := val.(string); ok {
				name = s
			}
		}
	}

	if name == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "env:set requires name or nameexpr attribute",
			Data:      map[string]any{"element": "set"},
			Cause:     fmt.Errorf("missing name"),
		}
	}

	span.SetAttributes(attribute.String("env.name", name))

	// Get the value - either from 'value' attribute or 'expr' attribute
	var value string
	valueAttr := string(el.GetAttribute("value"))
	exprAttr := string(el.GetAttribute("expr"))

	if valueAttr != "" && exprAttr != "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "env:set cannot have both value and expr attributes",
			Data:      map[string]any{"element": "set", "name": name},
			Cause:     fmt.Errorf("conflicting attributes"),
		}
	}

	if exprAttr != "" {
		// Evaluate the expression
		val, err := dm.EvaluateValue(ctx, exprAttr)
		if err != nil {
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   "Failed to evaluate expr",
				Data:      map[string]any{"element": "set", "name": name, "expr": exprAttr},
				Cause:     err,
			}
		}
		// Convert value to string
		value = fmt.Sprintf("%v", val)
	} else if valueAttr != "" {
		value = valueAttr
	} else {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "env:set requires value or expr attribute",
			Data:      map[string]any{"element": "set", "name": name},
			Cause:     fmt.Errorf("missing value"),
		}
	}

	// Set the environment variable
	if err := os.Setenv(name, value); err != nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Failed to set environment variable",
			Data:      map[string]any{"element": "set", "name": name, "value": value},
			Cause:     err,
		}
	}

	span.SetAttributes(attribute.String("env.value", value))

	return nil
}

var _ agentml.Namespace = (*Namespace)(nil)
