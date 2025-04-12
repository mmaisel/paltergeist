package generator

import (
	"log/slog"

	"cloud.google.com/go/vertexai/genai"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

// IntoSchema any T into a OAPI schema struct for genai.Schema.
func IntoSchema[T any](t T) genai.Schema {
	schema, err := openapi3gen.NewSchemaRefForValue(&t, nil)
	if err != nil {
		panic(err)
	}
	convertedSchema := convert(schema.Value)
	slog.Debug("converted response schema", "schema", convertedSchema)
	return convertedSchema
}

// convert converts an openapi3.Schema to a genai.Schema.
func convert(in *openapi3.Schema) genai.Schema {
	slog.Debug("convert", "type", in.Type.Slice(), "in", in)
	out := genai.Schema{}
	out.Description = in.Description
	out.Format = in.Format
	out.Title = in.Title
	out.Nullable = in.Nullable
	out.MinItems = int64(in.MinItems)
	if in.MaxItems != nil {
		out.MaxItems = int64(*in.MaxItems)
	}
	if len(in.Enum) > 0 {
		values := make([]string, 0)
		for _, value := range in.Enum {
			values = append(values, value.(string))
		}
		out.Enum = values
	}
	out.Required = in.Required
	out.MinProperties = int64(in.MinProps)
	if in.MaxProps != nil {
		out.MaxProperties = int64(*in.MaxProps)
	}
	if in.Min != nil {
		out.Minimum = *in.Min
	}
	if in.Max != nil {
		out.Maximum = *in.Max
	}
	if in.MaxLength != nil {
		out.MaxLength = int64(*in.MaxLength)
	}
	out.MinLength = int64(in.MinLength)
	// Set type specific fields
	if in.Type.Is("object") {
		out.Type = genai.TypeObject
		out.Properties = make(map[string]*genai.Schema)
		for key, value := range in.Properties {
			propertySchema := convert(value.Value)
			out.Properties[key] = &propertySchema
			// Make the object key required.
			slog.Debug("setting required", "key", key)
			out.Required = append(out.Required, key)
		}
		out.MinProperties = int64(len(out.Properties))
	} else if in.Type.Is("string") {
		out.Type = genai.TypeString
		out.MinLength = 1
	} else {
		slog.Error("not supported", "in", in)
		panic("not supported")
	}
	return out
}
