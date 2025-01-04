package inference

// ToolInputHelper converts a generic interface{} schema into a ToolInputSchema
func ToolInputHelper(schema interface{}) ToolInputSchema {
	if schemaMap, ok := schema.(map[string]interface{}); ok {
		return convertToToolInputSchema(schemaMap)
	}
	// Return empty schema if input is not a map
	return ToolInputSchema{}
}

// convertToToolInputSchema converts a map[string]interface{} to ToolInputSchema
func convertToToolInputSchema(schemaMap map[string]interface{}) ToolInputSchema {
	result := ToolInputSchema{
		Properties: make(map[string]Property),
	}

	// Handle type field
	if typeVal, ok := schemaMap["type"].(string); ok {
		result.Type = typeVal
	}

	// Handle required field
	if required, ok := schemaMap["required"].([]string); ok {
		result.Required = required
	} else if requiredInterface, ok := schemaMap["required"].([]interface{}); ok {
		// Convert []interface{} to []string
		required := make([]string, len(requiredInterface))
		for i, v := range requiredInterface {
			if str, ok := v.(string); ok {
				required[i] = str
			}
		}
		result.Required = required
	}

	// Handle properties field
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for key, value := range props {
			if propMap, ok := value.(map[string]interface{}); ok {
				property := convertToProperty(propMap)
				result.Properties[key] = property
			}
		}
	}

	return result
}

// convertToProperty converts a map[string]interface{} to Property
func convertToProperty(propertyMap map[string]interface{}) Property {
	property := Property{
		Items: make(map[string]interface{}),
	}

	// Handle type field
	if typeVal, ok := propertyMap["type"].(string); ok {
		property.Type = typeVal
	}

	// Handle description field
	if desc, ok := propertyMap["description"].(string); ok {
		property.Description = desc
	}

	// Handle items field for array types
	if items, ok := propertyMap["items"].(map[string]interface{}); ok {
		property.Items = items
	}

	return property
}
