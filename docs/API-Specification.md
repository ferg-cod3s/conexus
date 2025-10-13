# API Specification

## Overview

This document specifies the API interfaces for the Context Engine system. The system provides two primary interfaces:

1. **MCP Server (Primary)**: Model Context Protocol compliant interface for AI assistants and IDE integrations
2. **REST API (Secondary)**: HTTP-based API for web clients, CI/CD pipelines, and integrations that don't support MCP

### Key Design Principles

- **Security-first**: Input validation, authentication, and authorization at all layers
- **Performance**: Target <1s query latency for search operations
- **Consistency**: Standardized error handling, pagination, and response formats
- **Versioning**: URI-based versioning with backward compatibility guarantees

## 1. REST API Specification (v1)

**Base URL:** `/api/v1`
**Supported Content Types:** `application/json`
**Authentication:** JWT Bearer tokens (multi-user) or API keys (service accounts)
**Rate Limiting:** Token bucket algorithm with configurable tiers

### 1.1 Authentication

#### JWT Authentication
- **Header:** `Authorization: Bearer <jwt_token>`
- **Token Format:**
  ```json
  {
    "iss": "context-engine",
    "sub": "user-uuid",
    "exp": 1640995200,
    "iat": 1640991600,
    "aud": "api",
    "scope": "search:read index:write",
    "tenant_id": "org-uuid"
  }
  ```
- **Scopes:**
  - `search:read` - Perform searches
  - `index:read` - Read index status
  - `index:write` - Control indexing operations
  - `connectors:read` - List connectors
  - `connectors:write` - Manage connectors
  - `config:read` - Read configuration
  - `config:write` - Modify configuration

#### API Key Authentication
- **Header:** `X-API-Key: <api_key>`
- **Usage:** For service accounts and CI/CD integrations
- **Key Format:** 32-character hexadecimal string

#### Token Refresh
- **Endpoint:** `POST /auth/refresh`
- **Request:**
  ```json
  {
    "refresh_token": "refresh-jwt-token"
  }
  ```
- **Response:**
  ```json
  {
    "access_token": "new-access-jwt",
    "refresh_token": "new-refresh-jwt",
    "expires_in": 900
  }
  ```

### 1.2 Error Handling

#### Standard Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "detail": "The 'query' field is required and cannot be empty",
    "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
    "retryable": false,
    "docs_url": "https://docs.context-engine.com/errors/validation"
  }
}
```

#### Error Codes
- `VALIDATION_ERROR` (400): Invalid request parameters
- `AUTHENTICATION_ERROR` (401): Invalid or missing credentials
- `AUTHORIZATION_ERROR` (403): Insufficient permissions
- `NOT_FOUND_ERROR` (404): Resource not found
- `RATE_LIMIT_ERROR` (429): Rate limit exceeded
- `INTERNAL_ERROR` (500): Server error
- `SERVICE_UNAVAILABLE` (503): Service temporarily unavailable

#### HTTP Status Code Mappings
- `200 OK`: Success
- `201 Created`: Resource created
- `400 Bad Request`: Validation error
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Authorization failed
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict
- `422 Unprocessable Entity`: Semantic error
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service down

### 1.3 Rate Limiting

#### Rate Limit Headers
- `X-RateLimit-Limit`: Maximum requests per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when limit resets
- `Retry-After`: Seconds to wait before retrying (when exceeded)

#### Rate Limit Rules
- **Search endpoints**: 100 requests/minute per user
- **Index control**: 10 requests/minute per user
- **Connector management**: 20 requests/minute per user
- **Configuration**: 5 requests/minute per user
- **Health checks**: Unlimited

#### Rate Limit Exceeded Response (429)
```json
{
  "error": {
    "code": "RATE_LIMIT_ERROR",
    "message": "Rate limit exceeded",
    "detail": "Too many requests. Try again in 60 seconds.",
    "correlation_id": "550e8400-e29b-41d4-a716-446655440001",
    "retryable": true,
    "retry_after": 60
  }
}
```

### 1.4 Endpoints

#### Health & Status

##### `GET /health`
**Description:** System health check
**Authentication:** None
**Rate Limit:** Unlimited
**Response (200 OK):**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2025-10-12T12:00:00Z",
  "services": {
    "index": "healthy",
    "search": "healthy",
    "connectors": "healthy"
  }
}
```

##### `GET /status`
**Description:** Detailed system status
**Authentication:** Bearer token with `index:read` scope
**Rate Limit:** 10/minute
**Response (200 OK):**
```json
{
  "indexing": {
    "status": "active",
    "last_run": "2025-10-12T11:45:00Z",
    "documents_indexed": 15420,
    "connectors_active": 5
  },
  "performance": {
    "avg_query_latency_ms": 450,
    "queries_per_minute": 120,
    "error_rate": 0.01
  }
}
```

#### Search Endpoints

##### `POST /search`
**Description:** Performs a hybrid search across all indexed sources
**Authentication:** Bearer token with `search:read` scope
**Rate Limit:** 100/minute
**Request Body:**
```json
{
  "query": "How is user authentication handled?",
  "work_context": {
    "active_file": "src/api/routes/users.js",
    "git_branch": "feat/PROJ-123-new-feature",
    "open_ticket_ids": ["PROJ-123"]
  },
  "top_k": 20,
  "filters": {
    "source_types": ["file", "slack"],
    "date_range": {
      "from": "2025-01-01T00:00:00Z",
      "to": "2025-12-31T23:59:59Z"
    }
  },
  "pagination": {
    "cursor": "eyJwYWdlIjoxLCJsaW1pdCI6MjB9",
    "limit": 20
  }
}
```
**Success Response (200 OK):**
```json
{
  "results": [
    {
      "source": "file://src/auth/jwt.js",
      "content": "function signToken(user) { ... }",
      "score": 0.98,
      "metadata": {
        "type": "function",
        "lines": "15-30",
        "last_modified": "2025-10-10T14:30:00Z"
      }
    }
  ],
  "pagination": {
    "has_more": true,
    "next_cursor": "eyJwYWdlIjoyLCJsaW1pdCI6MjB9",
    "total_results": 150
  },
  "query_metadata": {
    "execution_time_ms": 450,
    "sources_searched": 5
  }
}
```

##### `GET /search/suggestions`
**Description:** Get search query suggestions
**Authentication:** Bearer token with `search:read` scope
**Rate Limit:** 50/minute
**Query Parameters:**
- `q`: Partial query string
- `limit`: Number of suggestions (default: 5)
**Response (200 OK):**
```json
{
  "suggestions": [
    "user authentication jwt",
    "authentication middleware",
    "login flow implementation"
  ]
}
```

#### Index Control Endpoints

##### `POST /index/start`
**Description:** Start indexing process
**Authentication:** Bearer token with `index:write` scope
**Rate Limit:** 10/minute
**Request Body:**
```json
{
  "connectors": ["github", "slack"],
  "force_full_reindex": false
}
```
**Response (202 Accepted):**
```json
{
  "job_id": "index-550e8400-e29b-41d4-a716-446655440002",
  "status": "queued",
  "estimated_duration_minutes": 30
}
```

##### `POST /index/stop`
**Description:** Stop ongoing indexing
**Authentication:** Bearer token with `index:write` scope
**Rate Limit:** 10/minute
**Response (200 OK):**
```json
{
  "status": "stopping",
  "message": "Indexing process will stop gracefully"
}
```

##### `GET /index/status`
**Description:** Get indexing status
**Authentication:** Bearer token with `index:read` scope
**Rate Limit:** 30/minute
**Response (200 OK):**
```json
{
  "status": "running",
  "current_job": {
    "id": "index-550e8400-e29b-41d4-a716-446655440002",
    "started_at": "2025-10-12T11:00:00Z",
    "progress": {
      "completed": 75,
      "total": 100,
      "current_connector": "github"
    }
  },
  "last_completed": {
    "finished_at": "2025-10-12T10:30:00Z",
    "duration_minutes": 25,
    "documents_processed": 15420
  }
}
```

##### `POST /index/force-reindex`
**Description:** Force full reindex of all connectors
**Authentication:** Bearer token with `index:write` scope
**Rate Limit:** 5/minute
**Request Body:**
```json
{
  "confirm_destructive": true,
  "connectors": ["all"]
}
```
**Response (202 Accepted):**
```json
{
  "job_id": "reindex-550e8400-e29b-41d4-a716-446655440003",
  "status": "queued",
  "warning": "This will clear existing index and may take several hours"
}
```

#### Connector Management Endpoints

##### `GET /connectors`
**Description:** List all configured connectors
**Authentication:** Bearer token with `connectors:read` scope
**Rate Limit:** 20/minute
**Response (200 OK):**
```json
{
  "connectors": [
    {
      "id": "github-repo-1",
      "type": "github",
      "name": "Main Repository",
      "status": "active",
      "last_sync": "2025-10-12T11:00:00Z",
      "config": {
        "repository": "org/repo",
        "include_issues": true,
        "include_prs": true
      }
    }
  ]
}
```

##### `POST /connectors`
**Description:** Add new connector
**Authentication:** Bearer token with `connectors:write` scope
**Rate Limit:** 10/minute
**Request Body:**
```json
{
  "type": "slack",
  "name": "Dev Team Channel",
  "config": {
    "workspace": "myworkspace",
    "channels": ["dev", "engineering"],
    "token": "xoxb-..."
  }
}
```
**Response (201 Created):**
```json
{
  "id": "slack-dev-1",
  "type": "slack",
  "name": "Dev Team Channel",
  "status": "configuring",
  "created_at": "2025-10-12T12:00:00Z"
}
```

##### `GET /connectors/{id}`
**Description:** Get connector details
**Authentication:** Bearer token with `connectors:read` scope
**Rate Limit:** 20/minute
**Response (200 OK):**
```json
{
  "id": "github-repo-1",
  "type": "github",
  "name": "Main Repository",
  "status": "active",
  "last_sync": "2025-10-12T11:00:00Z",
  "config": {
    "repository": "org/repo",
    "include_issues": true,
    "include_prs": true
  },
  "stats": {
    "documents_indexed": 5420,
    "last_error": null
  }
}
```

##### `PUT /connectors/{id}`
**Description:** Update connector configuration
**Authentication:** Bearer token with `connectors:write` scope
**Rate Limit:** 10/minute
**Request Body:**
```json
{
  "name": "Updated Repository Name",
  "config": {
    "repository": "org/repo",
    "include_issues": true,
    "include_prs": false
  }
}
```
**Response (200 OK):**
```json
{
  "id": "github-repo-1",
  "status": "reconfiguring",
  "message": "Configuration updated, reindexing will start shortly"
}
```

##### `DELETE /connectors/{id}`
**Description:** Remove connector
**Authentication:** Bearer token with `connectors:write` scope
**Rate Limit:** 5/minute
**Response (204 No Content):** (Empty body)

#### Configuration Management Endpoints

##### `GET /config`
**Description:** Get current configuration
**Authentication:** Bearer token with `config:read` scope
**Rate Limit:** 5/minute
**Response (200 OK):**
```json
{
  "search": {
    "max_results": 100,
    "default_top_k": 20,
    "enable_reranking": true
  },
  "indexing": {
    "batch_size": 100,
    "parallel_workers": 4,
    "retry_attempts": 3
  },
  "rate_limits": {
    "search_per_minute": 100,
    "index_per_minute": 10
  }
}
```

##### `PATCH /config`
**Description:** Update configuration
**Authentication:** Bearer token with `config:write` scope
**Rate Limit:** 2/minute
**Request Body:**
```json
{
  "search": {
    "max_results": 50
  }
}
```
**Response (200 OK):**
```json
{
  "updated": ["search.max_results"],
  "requires_restart": false
}
```

## 2. MCP Server Specification (v1)

This is the primary interface, compliant with the Model Context Protocol. It exposes tools and resources via stdio and HTTP/SSE transports.

### 2.1. Exposed Tools

**Tool: `context.search`**
* **Description:** Performs a comprehensive search using the user's query and current working context to find the most relevant code, discussions, and documents.
* **Input Schema (JSON Schema):**
  ```json
  {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "The user's natural language query."
      },
      "work_context": {
        "type": "object",
        "properties": {
          "active_file": {"type": "string"},
          "git_branch": {"type": "string"},
          "open_ticket_ids": {"type": "array", "items": {"type": "string"}}
        }
      },
      "top_k": {
        "type": "integer",
        "default": 20,
        "maximum": 100
      },
      "filters": {
        "type": "object",
        "properties": {
          "source_types": {
            "type": "array",
            "items": {"type": "string", "enum": ["file", "slack", "github", "jira"]}
          },
          "date_range": {
            "type": "object",
            "properties": {
              "from": {"type": "string", "format": "date-time"},
              "to": {"type": "string", "format": "date-time"}
            }
          }
        }
      }
    },
    "required": ["query"]
  }
  ```
* **Output:** A formatted string containing the top results, suitable for direct injection into an LLM context window.

**Tool: `context.get_related_info`**
* **Description:** Finds information directly related to the user's active file or ticket. Use this when the user asks a vague question like "what's the history of this file?"
* **Input Schema (JSON Schema):**
  ```json
  {
    "type": "object",
    "properties": {
      "file_path": {
        "type": "string",
        "description": "Path to the file to get related info for"
      },
      "ticket_id": {
        "type": "string",
        "description": "Ticket ID to get related info for"
      }
    }
  }
  ```
* **Output:** A summary of related PRs, tickets, and Slack discussions for the specified file or ticket.

**Tool: `context.index_control`**
* **Description:** Control indexing operations
* **Input Schema (JSON Schema):**
  ```json
  {
    "type": "object",
    "properties": {
      "action": {
        "type": "string",
        "enum": ["start", "stop", "status", "force_reindex"]
      },
      "connectors": {
        "type": "array",
        "items": {"type": "string"}
      }
    },
    "required": ["action"]
  }
  ```
* **Output:** Status information about the indexing operation.

**Tool: `context.connector_management`**
* **Description:** Manage data source connectors
* **Input Schema (JSON Schema):**
  ```json
  {
    "type": "object",
    "properties": {
      "action": {
        "type": "string",
        "enum": ["list", "add", "update", "remove"]
      },
      "connector_id": {"type": "string"},
      "connector_config": {"type": "object"}
    },
    "required": ["action"]
  }
  ```
* **Output:** Connector management results.

### 2.2. Exposed Resources

The MCP server exposes the entire indexed project file system as browsable resources.
**URI Scheme:** `engine://files/path/to/file.js`
**Supported Methods:** `resources/read`, `resources/list` (for directories). This allows clients like Claude Code to let users `@-mention` specific files to be included in the context.

## 3. OpenAPI Specification

```yaml
openapi: 3.0.3
info:
  title: Context Engine API
  version: 1.0.0
  description: API for intelligent code and document search
servers:
  - url: /api/v1
security:
  - bearerAuth: []
  - apiKeyAuth: []

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
    apiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key

  schemas:
    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
            detail:
              type: string
            correlation_id:
              type: string
            retryable:
              type: boolean
            docs_url:
              type: string

    SearchRequest:
      type: object
      required:
        - query
      properties:
        query:
          type: string
        work_context:
          type: object
          properties:
            active_file:
              type: string
            git_branch:
              type: string
            open_ticket_ids:
              type: array
              items:
                type: string
        top_k:
          type: integer
          default: 20
          maximum: 100
        filters:
          type: object
          properties:
            source_types:
              type: array
              items:
                type: string
                enum: [file, slack, github, jira]
            date_range:
              type: object
              properties:
                from:
                  type: string
                  format: date-time
                to:
                  type: string
                  format: date-time
        pagination:
          type: object
          properties:
            cursor:
              type: string
            limit:
              type: integer
              default: 20
              maximum: 100

    SearchResult:
      type: object
      properties:
        source:
          type: string
        content:
          type: string
        score:
          type: number
        metadata:
          type: object

    SearchResponse:
      type: object
      properties:
        results:
          type: array
          items:
            $ref: '#/components/schemas/SearchResult'
        pagination:
          type: object
          properties:
            has_more:
              type: boolean
            next_cursor:
              type: string
            total_results:
              type: integer
        query_metadata:
          type: object
          properties:
            execution_time_ms:
              type: integer
            sources_searched:
              type: integer

paths:
  /health:
    get:
      summary: Health check
      security: []
      responses:
        '200':
          description: System is healthy
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                  version:
                    type: string
                  timestamp:
                    type: string
                    format: date-time

  /search:
    post:
      summary: Perform search
      security:
        - bearerAuth: [search:read]
        - apiKeyAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SearchRequest'
      responses:
        '200':
          description: Search results
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SearchResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimitExceeded'

  /index/start:
    post:
      summary: Start indexing
      security:
        - bearerAuth: [index:write]
      responses:
        '202':
          description: Indexing started
        '401':
          $ref: '#/components/responses/Unauthorized'

  /connectors:
    get:
      summary: List connectors
      security:
        - bearerAuth: [connectors:read]
      responses:
        '200':
          description: List of connectors
          content:
            application/json:
              schema:
                type: object
                properties:
                  connectors:
                    type: array
                    items:
                      type: object
    post:
      summary: Add connector
      security:
        - bearerAuth: [connectors:write]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                type:
                  type: string
                name:
                  type: string
                config:
                  type: object
      responses:
        '201':
          description: Connector created
        '400':
          $ref: '#/components/responses/BadRequest'

responses:
  BadRequest:
    description: Bad request
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
  Unauthorized:
    description: Authentication required
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
  RateLimitExceeded:
    description: Rate limit exceeded
    headers:
      Retry-After:
        schema:
          type: integer
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
```

## 4. Migration and Compatibility

### Versioning Strategy
- URI-based versioning: `/api/v1`, `/api/v2`
- Backward compatibility maintained within major versions
- Deprecation notices provided 90 days before breaking changes
- Sunset periods of 180 days for deprecated endpoints

### Client Migration Guide
1. Update base URL to include version: `/api/v1`
2. Implement JWT authentication for multi-user scenarios
3. Handle new error response format
4. Update to cursor-based pagination for large result sets
5. Monitor rate limit headers and implement backoff logic

## 5. Security Considerations

- All endpoints require authentication except `/health`
- JWT tokens expire in 15 minutes, refresh tokens in 30 days
- API keys should be rotated regularly
- Input validation on all parameters
- Rate limiting prevents abuse
- HTTPS only in production
- Sensitive data encrypted at rest and in transit
