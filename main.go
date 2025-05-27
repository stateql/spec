package main

import (
	"log"
	"os"

	"stateql/db"
	"stateql/parser"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize database connection
	dsn := "host=localhost user=postgres password=postgres dbname=stateql port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize schema generator
	schemaGen := db.NewSchemaGenerator(database)

	// Initialize Gin router
	r := gin.Default()

	// Endpoint to parse and apply StateQL schema
	r.POST("/schema", func(c *gin.Context) {
		var schemaContent struct {
			Content string `json:"content"`
		}

		if err := c.BindJSON(&schemaContent); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		// Parse StateQL content
		stateql, err := parser.ParseStateQL(schemaContent.Content)
		if err != nil {
			c.JSON(400, gin.H{"error": "Failed to parse StateQL: " + err.Error()})
			return
		}

		// Generate database schema
		if err := schemaGen.GenerateSchema(stateql); err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate schema: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Schema generated successfully"})
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 