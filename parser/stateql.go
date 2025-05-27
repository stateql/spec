package parser

import (
	"bufio"
	"strings"
)

type Field struct {
	Name           string
	Type           string
	IsMany         bool
	Through        string
	IsAction       bool
	ActionType     string
	ActionArgs     map[string]string
	RequiredParams []string
	FunctionName   string    // For fields with functions (like either, sum, etc)
	FunctionArgs   []string  // Arguments for the function
}

type Entity struct {
	Name   string
	Fields []Field
}

type StateQL struct {
	Entities []Entity
}

func ParseStateQL(content string) (*StateQL, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var stateql StateQL
	var currentEntity *Entity

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check if this is an entity definition
		if !strings.HasPrefix(line, "-") {
			// Remove colon if present
			entityName := strings.TrimSuffix(line, ":")
			entity := Entity{Name: entityName}
			stateql.Entities = append(stateql.Entities, entity)
			currentEntity = &stateql.Entities[len(stateql.Entities)-1]
			continue
		}

		// Parse field definition
		if currentEntity != nil {
			field := parseField(line)
			currentEntity.Fields = append(currentEntity.Fields, field)
		}
	}

	return &stateql, nil
}

func parseField(line string) Field {
	// Remove the leading dash and space
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimSpace(line)

	parts := strings.Split(line, " is ")
	if len(parts) != 2 {
		return Field{}
	}

	name := strings.TrimSpace(parts[0])
	typeDef := strings.TrimSpace(parts[1])

	field := Field{
		Name: name,
	}

	// Check if it's an action
	if strings.HasPrefix(typeDef, "action") {
		field.IsAction = true
		actionParts := strings.Split(typeDef, " thru ")
		if len(actionParts) == 2 {
			field.ActionType = strings.TrimSpace(actionParts[1])
			// Parse action arguments if present
			if strings.Contains(field.ActionType, "(") {
				actionName := field.ActionType[:strings.Index(field.ActionType, "(")]
				args := field.ActionType[strings.Index(field.ActionType, "(")+1:strings.Index(field.ActionType, ")")]
				field.ActionType = actionName
				field.ActionArgs, field.RequiredParams = parseActionArgs(args)
			}
		}
		return field
	}

	// Check if it's a many relationship
	if strings.HasPrefix(typeDef, "many") {
		field.IsMany = true
		throughParts := strings.Split(typeDef, " thru ")
		if len(throughParts) == 2 {
			field.Through = strings.TrimSpace(throughParts[1])
		}
		return field
	}

	// Check if it has a function (thru)
	if strings.Contains(typeDef, " thru ") {
		parts := strings.Split(typeDef, " thru ")
		field.Type = strings.TrimSpace(parts[0])
		functionDef := strings.TrimSpace(parts[1])
		
		if strings.Contains(functionDef, "(") {
			field.FunctionName = functionDef[:strings.Index(functionDef, "(")]
			args := functionDef[strings.Index(functionDef, "(")+1:strings.Index(functionDef, ")")]
			field.FunctionArgs = parseFunctionArgs(args)
		}
		return field
	}

	// Regular field
	field.Type = typeDef
	return field
}

func parseActionArgs(args string) (map[string]string, []string) {
	result := make(map[string]string)
	var requiredParams []string
	
	// Handle space-separated arguments
	parts := strings.Fields(args)
	for i := 0; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		
		// Handle required parameters (with : prefix)
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			requiredParams = append(requiredParams, paramName)
			continue
		}
		
		// Handle key=value pairs
		if strings.Contains(part, "=") {
			kv := strings.Split(part, "=")
			if len(kv) == 2 {
				result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
			continue
		}
		
		// Handle positional arguments
		if i == 0 {
			result["arg1"] = part
		} else if i == 1 {
			result["arg2"] = part
		}
	}
	
	return result, requiredParams
}

func parseFunctionArgs(args string) []string {
	var result []string
	parts := strings.Fields(args)
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Remove quotes if present
		if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
			part = strings.Trim(part, "\"")
		}
		result = append(result, part)
	}
	
	return result
} 