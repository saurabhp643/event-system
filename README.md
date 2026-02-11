# Multi-Tenant Event Ingestion System

A production-ready event ingestion platform supporting multiple isolated tenants with real-time streaming capabilities. Built with Go (backend) and React + TypeScript (frontend).

## Architecture Overview

The system follows a clean layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend Layer                         │
│  ┌─────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Tenant      │  │ Live Event Feed │  │ Event Dashboard │ │
│  │ Management  │  │ (WebSocket)     │  │ & Statistics   │ │
│  └─────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway Layer                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Gin Router with CORS, Rate Limiting, Auth Middleware │ │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Auth        │  │ Event       │  │ WebSocket Hub       │  │
│  │ Service     │  │ Service     │  │ (Real-time Push)    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Persistence Layer                         │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  GORM ORM (SQLite for dev, PostgreSQL for prod)     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Multi-Tenancy Model
- **Database-per-tenant isolation** is achieved via `tenant_id` foreign keys on all event records
- API key authentication ensures tenants can only access their own data
- Each tenant receives a unique API key upon creation

### 2. Real-Time Architecture
- **WebSocket-based streaming** using `gorilla/websocket` for bi-directional communication
- **Publish-Subscribe pattern** via a centralized hub that manages all active connections
- Events are broadcast immediately upon ingestion to all connected clients of the respective tenant
- Automatic reconnection with exponential backoff on the frontend

### 3. API Design
- RESTful endpoints following standard HTTP semantics
- Consistent JSON response formats across all endpoints
- Proper HTTP status codes (201 for creation, 429 for rate limits, etc.)
- Health check endpoint for load balancer integration

### 4. Database Strategy
- **GORM ORM** provides abstraction layer enabling SQLite (local) and PostgreSQL (production)
- Connection pooling with configurable max open/idle connections
- Connection lifetime management to prevent stale connections

## Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Backend | Go 1.21 + Gin | High performance, low memory footprint, excellent concurrency model |
| ORM | GORM | Mature ORM with auto-migration support, database-agnostic design |
| Database | SQLite (dev) / PostgreSQL (prod) | SQLite for zero-config development, PostgreSQL for production reliability |
| Real-time | gorilla/websocket | Battle-tested WebSocket implementation with fallback support |
| Frontend | React 18 + TypeScript | Component-based UI with type safety for maintainability |
| Build Tool | Vite | Fast HMR, optimized production builds |
| Styling | TailwindCSS | Utility-first CSS for rapid UI development |

## API Endpoints

### Tenant Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tenants` | Create a new tenant with auto-generated API key |
| GET | `/api/v1/tenants` | List all tenants (public endpoint) |

### Event Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/events` | Ingest a new event (requires API key) |
| GET | `/api/v1/events` | Retrieve events with filtering support |
| GET | `/api/v1/events/stats` | Get aggregated event statistics |

### Real-Time
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/ws` | WebSocket connection (query param: `api_key`) |

## Features Implemented

### Core Requirements
- ✅ Tenant creation and management with unique API keys
- ✅ Event ingestion with structured metadata support
- ✅ Real-time WebSocket streaming per tenant
- ✅ PostgreSQL persistence with GORM
- ✅ API key authentication middleware
- ✅ Per-tenant rate limiting

### Frontend Features
- **Tenant Selector**: Dropdown to switch between tenants with visual indicators
- **Live Event Feed**: Real-time updates with WebSocket, auto-scroll toggle
- **Event Statistics**: Visual breakdown of events by type with progress bars
- **Event Filtering**: Search by event type or metadata content
- **Toast Notifications**: User feedback for success/error states
- **Connection Status**: Visual indicator for WebSocket connectivity


## Trade-offs & Assumptions

### Deliberate Simplifications
1. **In-Memory Rate Limiting**: Used fixed-window counter for simplicity. Would use Redis for distributed deployments.

2. **Synchronous Webhook Delivery**: Webhooks fire immediately on event ingestion. Production systems should use message queues (RabbitMQ/Kafka) for reliability and retry handling.

3. **No Event Deduplication**: Assumes events are idempotent. Production would require deduplication logic using event IDs.

4. **Single-Instance Architecture**: WebSocket hub is in-memory, limiting to single-node deployments. Would use Redis Pub/Sub for horizontal scaling.

5. **Basic LIKE-based Search**: Metadata search uses SQL LIKE clauses. Production would benefit from Elasticsearch or PostgreSQL full-text search.

6. **No Metrics Export**: Application lacks Prometheus/Graphite endpoints. Would add for production monitoring.

### Key Assumptions
- Events are append-only and immutable
- Tenant API keys are not rotated frequently (no key versioning)
- WebSocket connections are relatively short-lived
- Metadata JSON size is bounded (< 1MB per event)
- Rate limits are per-minute windows with reasonable thresholds

## Project Structure

```
event-ingestion-system/
├── backend/
│   ├── main.go                          # Application entry point
│   ├── config.yaml                      # Configuration file
│   └── internal/
│       ├── auth/                        # Authentication middleware
│       ├── config/                      # Configuration loading
│       ├── database/                    # GORM database layer
│       ├── handlers/                    # HTTP request handlers
│       ├── middleware/                  # Gin middleware (CORS, rate limiting)
│       ├── models/                      # Data models (Tenant, Event)
│       └── websocket/                    # WebSocket hub implementation
├── frontend/
│   ├── src/
│   │   ├── components/                  # React components
│   │   ├── App.tsx                      # Main application component
│   │   └── main.tsx                     # Application entry point
│   ├── vite.config.ts                   # Vite configuration
│   └── Dockerfile                       # Frontend containerization
├── Dockerfile                           # Backend containerization
├── docker-compose.yml                   # Local development orchestration
└── README.md                            # This file
```

## Authentication Model

The system implements a dual authentication mechanism:

1. **API Key Authentication**: Primary method for event ingestion and retrieval. Each tenant receives a unique API key upon creation, passed via the `X-API-Key` header.

2. **JWT Token Authentication**: Optional JWT-based authentication for longer-lived sessions. Tokens are issued via the `/tenants/:id/token` endpoint.

Both methods enforce tenant isolation - API keys and tokens are tenant-scoped.

## Error Handling Strategy

### Backend (Go)
- Structured error codes: `ErrValidation`, `ErrAuthentication`, `ErrNotFound`, `ErrRateLimit`, `ErrServer`
- Panic recovery middleware prevents crashes
- Security headers (X-XSS-Protection, HSTS)
- Request logging with timing for debugging

### Frontend (React)
- Error categorization for appropriate user feedback
- Toast notifications for non-blocking errors
- Retry logic with exponential backoff for network failures
- Loading states throughout for improved UX

## Running Locally

```bash
# Backend (port 8080)
cd backend && go run main.go

# Frontend (port 5173)
cd frontend && npm run dev
```

Access the dashboard at `http://localhost:5173`

## License

MIT License
