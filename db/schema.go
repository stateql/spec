package db

import (
	"fmt"
	"strings"

	"stateql/parser"

	"gorm.io/gorm"
)

type SchemaGenerator struct {
	db *gorm.DB
}

func NewSchemaGenerator(db *gorm.DB) *SchemaGenerator {
	return &SchemaGenerator{db: db}
}

func (sg *SchemaGenerator) GenerateSchema(stateql *parser.StateQL) error {
	// Create tables for each entity
	for _, entity := range stateql.Entities {
		if err := sg.createTable(entity); err != nil {
			return err
		}
	}

	// Create relationship tables
	for _, entity := range stateql.Entities {
		if err := sg.createRelationships(entity); err != nil {
			return err
		}
	}

	return nil
}

func (sg *SchemaGenerator) createTable(entity parser.Entity) error {
	// Create base table
	tableName := strings.ToLower(entity.Name)
	
	// Start building the CREATE TABLE statement
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName)
	
	// Add primary key
	createSQL += "id SERIAL PRIMARY KEY,"
	
	// Add regular fields
	for _, field := range entity.Fields {
		if !field.IsMany && !field.IsAction {
			columnType := mapTypeToPostgres(field.Type)
			createSQL += fmt.Sprintf("%s %s,", strings.ToLower(field.Name), columnType)
		}
	}
	
	// Remove trailing comma and close the statement
	createSQL = strings.TrimSuffix(createSQL, ",") + ")"
	
	return sg.db.Exec(createSQL).Error
}

func (sg *SchemaGenerator) createRelationships(entity parser.Entity) error {
	for _, field := range entity.Fields {
		if field.IsMany {
			// Create junction table for many-to-many relationships
			tableName := fmt.Sprintf("%s_%s", strings.ToLower(entity.Name), strings.ToLower(field.Name))
			createSQL := fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					%s_id INTEGER REFERENCES %s(id),
					%s_id INTEGER REFERENCES %s(id),
					PRIMARY KEY (%s_id, %s_id)
				)`, 
				tableName,
				strings.ToLower(entity.Name), strings.ToLower(entity.Name),
				strings.ToLower(field.Through), strings.ToLower(field.Through),
				strings.ToLower(entity.Name), strings.ToLower(field.Through))
			
			if err := sg.db.Exec(createSQL).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func mapTypeToPostgres(stateqlType string) string {
	switch strings.ToLower(stateqlType) {
	case "text":
		return "TEXT"
	case "number":
		return "NUMERIC"
	case "switch":
		return "BOOLEAN"
	case "date":
		return "DATE"
	case "timestamp":
		return "TIMESTAMP"
	case "seconds":
		return "INTEGER"
	default:
		return "TEXT"
	}
} 