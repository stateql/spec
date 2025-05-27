# StateQL Microservice

A Go-based microservice that converts StateQL state definitions into a PostgreSQL database schema.

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Docker (optional)

## Setup

1. Install dependencies:

```bash
go mod download
```

2. Set up PostgreSQL:

```bash
# Using Docker
docker run --name stateql-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=stateql -p 5432:5432 -d postgres:latest
```

3. Configure environment variables (optional):

```bash
export PORT=8080
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=stateql
export DB_PORT=5432
```

## Running the Service

```bash
go run main.go
```

The service will start on port 8080 by default.

## API Endpoints

### POST /schema

Converts a StateQL definition into a PostgreSQL schema.

Request body:

```json
{
  "content": "User:\n- id is text\n- name is text\n..."
}
```

Response:

```json
{
  "message": "Schema generated successfully"
}
```

## Example StateQL Definition

```stateql
User:
- id is text
- name is text
- friends is many User thru befriendedBy
- befriendedBy is many User thru friends

Task:
- id is text
- title is text
- content is text
- author is User thru authoredTasks
```

## Development

To add new features or modify the service:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT
