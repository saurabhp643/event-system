# Event Ingestion System

A production-ready multi-tenant event ingestion system with real-time dashboard.

## Features

### Core Features
- **Tenant Management**: Create and manage multiple isolated tenants
- **Event Ingestion API**: RESTful API to ingest events with metadata
- **Real-Time Streaming**: WebSocket endpoint for live event streaming
- **Persistence**: SQLite (local) / PostgreSQL (production) via GORM
- **Authentication**: JWT tokens and API key authentication
- **Rate Limiting**: Per-tenant rate limiting to prevent abuse

### Frontend Features
- **Tenant Selector**: Easy dropdown to switch between tenants
- **Live Event Feed**: Real-time event display with WebSocket
- **Event Filtering**: Search and filter events by type or metadata
- **Event Statistics**: Visual breakdown of events by type
- **Auto-scroll**: Optional auto-scroll for live feed
- **Dark/Light Theme**: Clean, modern UI with TailwindCSS
- **Toast Notifications**: User-friendly error and success messages
- **Loading States**: Smooth UX with spinners and skeleton loaders

### Bonus Features
- **Webhook Delivery**: Configurable webhooks for specific event types
- **Event Search**: RAG-style search over event metadata
- **Event Statistics**: Real-time statistics and charts

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (React + TypeScript)       │
│  ┌─────────────┐  ┌─────────────────┐  ┌─────────────────┐│
│  │ Tenant      │  │ Live Event Feed │  │ Event Search    ││
│  │ Selector    │  │ (WebSocket)     │  │ & Filters       ││
│  └─────────────┘  └─────────────────┘  └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Backend (Golang)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐│
│  │ Auth        │  │ REST API    │  │ WebSocket Server    ││
│  │(JWT/API Key)│  │   (Gin)     │  │ (gorilla/websocket) ││
│  └─────────────┘  └─────────────┘  └─────────────────────┘│
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐│
│  │ Rate Limiter│  │ Event Store │  │ Nginx Proxy         ││
│  │ (In-Memory) │  │(GORM)       │  │ (Production)        ││
│  └─────────────┘  └─────────────┘  └─────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Tech Stack

- **Backend**: Go 1.21, Gin, GORM, SQLite/PostgreSQL, gorilla/websocket
- **Frontend**: React 18, TypeScript, Vite, TailwindCSS
- **Authentication**: JWT, API Keys
- **Docker**: Multi-stage builds, docker-compose
- **Nginx**: Reverse proxy for production
- **CI/CD**: GitHub Actions

## Database Support

### Local Development (SQLite)
Default configuration uses SQLite for simplicity:
```yaml
database:
  driver: "sqlite"
  host: "./data/events.db"
```

### Production Deployment (PostgreSQL)
For production, use PostgreSQL for better reliability:
```yaml
database:
  driver: "postgres"
  host: "your-db-host.internal"
```

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (for local development)

### Local Development

1. **Clone the repository**
```bash
git clone <repository-url>
cd event-ingestion-system
```

2. **Start backend**
```bash
cd backend
go run main.go
```

3. **Start frontend** (in another terminal)
```bash
cd frontend
npm install
npm run dev
```

4. **Access the application**
- Frontend: http://localhost:3000
- API: http://localhost:8080
- Health Check: http://localhost:8080/health

### Docker Deployment (Unified Full-Stack)

```bash
# Build and run all services
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build

# Access the application at http://localhost:8080
```

### Environment Variables

Create a `.env` file in the root directory:

```env
# For SQLite (local development)
DB_DRIVER=sqlite
DATABASE_PATH=./data/events.db

# For PostgreSQL (production)
DB_DRIVER=postgres
DB_HOST=your-render-db-host.internal
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=render

JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

## API Documentation

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tenants` | Create a new tenant |
| GET | `/api/v1/tenants` | List all tenants |
| GET | `/api/v1/tenants/:id` | Get tenant details |
| POST | `/api/v1/events` | Ingest an event |
| GET | `/api/v1/events` | List events (with filtering) |
| GET | `/api/v1/events/stats` | Get event statistics |
| GET | `/api/v1/ws` | WebSocket endpoint |
| GET | `/health` | Health check |

### Event Ingestion Request

```json
{
  "tenant_id": "uuid",
  "event_type": "page_view",
  "timestamp": "2024-01-15T10:30:00Z",
  "metadata": {
    "page": "/home",
    "user_agent": "Mozilla/5.0"
  }
}
```

### Authentication

Include API key in header:
```
X-API-Key: your-api-key-here
```

Or JWT token in Authorization header:
```
Authorization: Bearer your-jwt-token
```

## Deployment Guide

### Recommended: Render + Vercel

#### Backend on Render (with PostgreSQL)

1. **Create Render Account**
   - Go to [Render Dashboard](https://dashboard.render.com)
   - Sign up with GitHub

2. **Create PostgreSQL Database**
   - Click "New +" → "PostgreSQL"
   - Name: `event-ingestion-db`
   - Select free plan
   - Click "Create Database"
   - **Important**: Copy the "Internal Database URL" (format: `postgres://user:password@host:5432/db`)

3. **Create Backend Web Service**
   - Click "New +" → "Web Service"
   - Connect your GitHub repository
   - Configure:
     - **Name**: `event-ingestion-backend`
     - **Branch**: `main`
     - **Build Command**: `cd backend && go build -o bin/server .`
     - **Start Command**: `./bin/server`
     - **Environment Variables**: Click "Add" and add:
       | Key | Value |
       |-----|-------|
       | `DB_DRIVER` | `postgres` |
       | `DB_HOST` | Host from PostgreSQL URL (before `:`) |
       | `DB_PORT` | `5432` |
       | `DB_USER` | User from PostgreSQL URL |
       | `DB_PASSWORD` | Password from PostgreSQL URL |
       | `DB_NAME` | `render` (or as specified) |
       | `GIN_MODE` | `release` |
       | `PORT` | `10000` |
       | `JWT_SECRET` | (generate a secure random string) |
   - Click "Create Web Service"

4. **Get Backend URL**
   - Wait for deployment to complete
   - Copy the service URL (e.g., `https://event-ingestion-backend.onrender.com`)

#### Frontend on Vercel

1. **Create Vercel Account**
   - Go to [Vercel Dashboard](https://vercel.com)
   - Sign up with GitHub

2. **Import Project**
   - Click "Add New Project"
   - Import your GitHub repository
   - Configure:
     - **Framework Preset**: `Vite` (or Other)
     - **Build Command**: `npm run build`
     - **Output Directory**: `dist`
     - **Environment Variables**: Add:
       | Key | Value |
       |-----|-------|
       | `VITE_API_URL` | Your Render backend URL |
   - Click "Deploy"

3. **Get Frontend URL**
   - Wait for deployment to complete
   - Copy the deployed URL

#### Test the Deployment

1. Open your Vercel frontend URL
2. Select a tenant from the dropdown (or create one)
3. View live events in the feed
4. Test creating events and verify they appear in real-time

---

### Alternative: Render Full-Stack Docker

Deploy both frontend and backend as a single Docker container:

1. **Push to GitHub**
   ```bash
   git add .
   git commit -m "Initial commit"
   git push origin main
   ```

2. **Deploy to Render**
   - Click "New +" → "Web Service"
   - Connect your GitHub repo
   - Configure:
     - **Environment**: `Docker`
     - **Build Command**: (leave empty)
     - **Start Command**: (leave empty)
   - Add environment variables:
     | Key | Value |
     |-----|-------|
     | `JWT_SECRET` | (generate) |
     | `DB_DRIVER` | `sqlite` (or `postgres`) |
     | `DATABASE_PATH` | `/app/data/events.db` |
   - Add Disk (if using SQLite):
     - Mount Path: `/app/data`
     - Size: 1 GB
   - Click "Create Web Service"

---

### Alternative: Docker + Railway

Deploy both services together:

1. **Create Railway Account**: [railway.app](https://railway.app)
2. **Deploy Backend with Docker**
   ```bash
   npm i -g railway
   railway login
   railway init
   railway add postgresql
   railway up
   ```
3. **Set environment variables** in Railway dashboard

---

## Database Setup Options

| Platform | Free Tier | Best For |
|----------|-----------|----------|
| **Render PostgreSQL** | 1GB storage | Production on Render |
| **Supabase** | 500MB | PostgreSQL with API |
| **Neon** | 600MB | Serverless PostgreSQL |
| **ElephantSQL** | 20MB | Small projects |

## Trade-offs and Assumptions

1. **Database Choice**: Default to SQLite for local development, PostgreSQL for production. GORM makes switching seamless.

2. **Rate Limiting**: In-memory rate limiting for simplicity. For distributed systems, use Redis.

3. **No Webhook Retry Queue**: Webhooks delivered synchronously. For production, use RabbitMQ/Kafka.

4. **No Event Deduplication**: Assuming events are idempotent.

5. **Single-Instance Deployment**: For horizontal scaling, add distributed caching and message queues.

6. **Basic Search**: LIKE-based search. For production, consider Elasticsearch.

7. **No Metrics Export**: Could add Prometheus for monitoring.

## Project Structure

```
event-ingestion-system/
├── backend/
│   ├── main.go
│   ├── config.yaml
│   ├── internal/
│   │   ├── auth/
│   │   ├── config/
│   │   ├── database/
│   │   ├── errors/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── models/
│   │   └── websocket/
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── lib/
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   └── Dockerfile
├── Dockerfile
├── docker-compose.yml
├── render.yaml
└── README.md
```

## Error Handling Best Practices

### Backend (Go)
- Structured error codes (Validation, Authentication, Not Found, Rate Limit, Server)
- Panic recovery middleware
- Security headers (XSS protection, HSTS)
- Request logging with timing

### Frontend (React + TypeScript)
- Error categorization (network, authentication, validation)
- User-friendly error messages
- Toast notifications for feedback
- Retry logic with exponential backoff
- Loading states for better UX

## CI/CD Pipeline

The project includes GitHub Actions workflow (`.github/workflows/ci.yml`) that:
1. Lints Go code
2. Runs Go tests
3. Lints TypeScript code
4. Builds frontend
5. Runs frontend tests

## License

MIT License
