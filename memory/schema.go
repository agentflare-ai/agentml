package memory

import (
	"github.com/agentflare-ai/go-jsonschema"
)

// NewInsertSchema creates an INSERT schema with the given columns
func NewInsertSchema(columnsSchema *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Insert a new row into the table",
		Properties: map[string]*jsonschema.Schema{
			"columns": columnsSchema,
		},
		Required: []string{"columns"},
	}
}

// NewUpdateSchema creates an UPDATE schema with the given columns
func NewUpdateSchema(columnsSchema *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Update existing rows in the table",
		Properties: map[string]*jsonschema.Schema{
			"set":   columnsSchema,
			"where": {Type: "string", Description: "The where clause"},
		},
		Required: []string{"set", "where"},
	}
}

func NewSelectSchema(columnsSchema *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Select rows from the table",
		Properties: map[string]*jsonschema.Schema{
			"columns": columnsSchema,
			"where":   {Type: "string", Description: "The where clause"},
		},
		Required: []string{"columns"},
	}
}

// NewDeleteSchema creates a DELETE schema with the given where conditions
func NewDeleteSchema(whereSchema *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Delete rows from the table",
		Properties: map[string]*jsonschema.Schema{
			"where": {Type: "string", Description: "The where clause"},
		},
		Required: []string{"where"},
	}
}

// NewBatchOperationSchema creates a schema for batch operations on a table
func NewBatchOperationSchema(columnsSchema *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Batch table operations",
		Properties: map[string]*jsonschema.Schema{
			"inserts": {
				Type:        "array",
				Description: "New rows to insert",
				Items:       NewInsertSchema(columnsSchema),
			},
			"updates": {
				Type:        "array",
				Description: "Existing rows to update",
				Items:       NewUpdateSchema(columnsSchema),
			},
			"deletes": {
				Type:        "array",
				Description: "Rows to delete",
				Items:       NewDeleteSchema(columnsSchema),
			},
		},
	}
}

type InsertOperation struct {
	Columns map[string]any `json:"columns"`
}

type UpdateOperation struct {
	Set   []map[string]any `json:"set"`
	Where string           `json:"where"`
}

type DeleteOperation struct {
	Where string
}

type BatchOperations struct {
	Inserts []InsertOperation
	Updates []UpdateOperation
	Deletes []DeleteOperation
}

var FunctionResponseSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"will_continue": {
			Type:        "boolean",
			Description: "Signals that function call continues, and more responses will be returned, turning the function call into a generator",
		},
		"id": {
			Type:        "string",
			Description: "The ID of the function call this response is for",
		},
		"name": {
			Type:        "string",
			Description: "The name of the function to call",
		},
		"response": {
			Type:                 "object",
			Description:          "The function response in JSON object format",
			AdditionalProperties: true,
		},
	},
	Required:    []string{"name", "response"},
	Description: "A function response",
}

var FunctionCallSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"id": {
			Type:        "string",
			Description: "The unique ID of the function call",
		},
		"args": {
			Type:                 "object",
			Description:          "The function parameters and values in JSON object format",
			AdditionalProperties: true,
		},
		"name": {
			Type:        "string",
			Description: "The name of the function to call",
		},
	},
	Required:    []string{"name"},
	Description: "A function call",
}

// PartSchema is a schema for a part of a message genai.Part
var PartSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"text": {
			Type:        "string",
			Description: "The text of the part",
		},
		"thought": {
			Type:        "boolean",
			Description: "Whether the part is thought from the model",
		},
		"thought_signature": {
			Type:        "string",
			Description: "The signature of the thought",
		},
		"function_call":     FunctionCallSchema,
		"function_response": FunctionResponseSchema,
	},
}
