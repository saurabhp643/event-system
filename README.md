# Event Ingestion System

A production-ready multi-tenant event ingestion system with real-time dashboard.

## Features

### Core Features
- **Tenant Management**: Create and manage multiple isolated tenants
- **Event Ingestion API**: RESTful API to ingest events with metadata
- **Real-Time Streaming**: WebSocket endpoint for live event streaming
- **Persistence**: SQLite database for reliable storage
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
│                        Frontend (React + TS)                │
│  ┌─────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Tenant      │  │ Live Event Feed │  │ Event Search    │  │
│  │ Selector    │  │ (WebSocket)     │  │ & Filters       │  │
│  └─────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Backend (Golang)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Auth        │  │ REST API    │  │ WebSocket Server    │  │
│  │(JWT/API Key)│  │   (Gin)     │  │ (gorilla/websocket) │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Rate Limiter│  │ Event Store │  │ Redis Pub/Sub       │  │
│  │ (Redis)     │  │(GORM/SQLite)│  │ (Real-time)         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Tech Stack

- **Backend**: Go 1.21, Gin, GORM, SQLite, gorilla/websocket
- **Frontend**: React 18, TypeScript, Vite, TailwindCSS
- **Authentication**: JWT, API Keys
- **Docker**: Multi-stage builds, docker-compose
- **CI/CD**: GitHub Actions

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (for local development)

### Development

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

### Docker Deployment

```bash
# Build and run all services
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build
```

### Environment Variables

Create a `.env` file in the root directory:

```env
JWT_SECRET=your-super-secret-jwt-key-change-in-production
DATABASE_PATH=/app/data/events.db
REDIS_HOST=localhost:6379
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

### Recommended: Render (Backend) + Vercel (Frontend)

This is the easiest and most reliable combination with free tiers.

---

#### Step 1: Push to GitHub

```bash
cd event-ingestion-system
git init
git add .
git commit -m "Initial commit: multi-tenant event ingestion system"
git remote add origin https://github.com/yourusername/event-ingestion-system.git
git push -u origin main
```

---

#### Step 2: Deploy Backend to Render

1. **Create Render Account**
   - Go to [Render Dashboard](https://dashboard.render.com)
   - Sign up with GitHub

2. **Create PostgreSQL Database**
   - Click "New +" → "PostgreSQL"
   - Name: `event-ingestion-db`
   - Select free plan
   - Click "Create Database"
   - **Note**: Copy the "Internal Database URL" (you'll need it later)

3. **Create Backend Web Service**
   - Click "New +" → "Web Service"
   - Connect your GitHub repository
   - Configure:
     - **Name**: `event-ingestion-backend`
     - **Branch**: `main`
     - **Build Command**: `cd backend && go build -o bin/server .`
     - **Start Command**: `./bin/server`
     - **Environment Variables**: Click "Add" and add:
       - `DB_HOST`: From PostgreSQL connection string (host part)
       - `DB_PORT`: `5432`
       - `DB_USER`: From PostgreSQL connection string (user part)
       - `DB_PASSWORD`: From PostgreSQL connection string (password part)
       - `DB_NAME`: `render` (or as specified)
       - `GIN_MODE`: `release`
       - `PORT`: `10000`
   - Click "Create Web Service"

4. **Get Backend URL**
   - Wait for deployment to complete
   - Copy the service URL (e.g., `https://event-ingestion-backend.onrender.com`)

---

#### Step 3: Deploy Frontend to Vercel

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
       - `VITE_API_URL`: Your Render backend URL (e.g., `https://event-ingestion-backend.onrender.com`)
   - Click "Deploy"

3. **Get Frontend URL**
   - Wait for deployment to complete
   - Copy the deployed URL (e.g., `https://event-ingestion-system.vercel.app`)

---

#### Step 4: Test the Deployment

1. Open your Vercel frontend URL
2. Select a tenant from the dropdown
3. View live events in the feed
4. Test creating a new tenant
5. Verify WebSocket connection is working

---

### Alternative: Docker + Railway

Deploy both services together using Docker:

1. **Create Railway Account**
   - Go to [Railway](https://railway.app)
   - Sign up with GitHub

2. **Deploy Backend with Docker**
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Connect your repository
   - Add PostgreSQL plugin
   - Set environment variables from `.env`
   - Deploy using Dockerfile or docker-compose.yml

3. **Deploy Frontend with Docker**
   - Create a new Railway service
   - Connect your repository
   - Configure build and start commands
   - Set `VITE_API_URL` to your backend URL

---

### Alternative: Render (Full Stack with Docker)

Deploy both frontend and backend as a single Docker service:

1. Use the existing `Dockerfile` in root directory
2. Connect to Render
3. Configure:
   - Build Command: `docker build -t event-ingestion .`
   - Start Command: `docker run -p 10000:8080 event-ingestion`
4. Set all environment variables
5. Deploy

---

### Alternative: Google Cloud Run

1. **Create Google Cloud Account**
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project

2. **Enable Required APIs**
   - Cloud Run API
   - Cloud SQL Admin API

3. **Deploy Backend**
   ```bash
   gcloud run deploy event-ingestion-backend \
     --source backend \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

4. **Deploy Frontend**
   - Build static files: `cd frontend && npm run build`
   - Deploy to Cloud Run:
   ```bash
   gcloud run deploy event-ingestion-frontend \
     --image gcr.io/PROJECT_ID/frontend \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

---

### Database Setup (if not using managed service)

| Platform | Free Tier | Connection String Format |
|----------|-----------|-------------------------|
| **Supabase** | 500MB | `postgres://user:password@host:5432/db` |
| **Neon** | 600MB | `postgres://user:password@host:5432/db` |
| **ElephantSQL** | 20MB | `postgres://user:password@host:5432/db` |
| **Render PostgreSQL** | Free | Provided in dashboard |

---

## Trade-offs and Assumptions

1. **Database Choice**: Used SQLite for simplicity and portability. In production, PostgreSQL would be preferred for better concurrency and reliability.

2. **Rate Limiting**: Implemented in-memory rate limiting. For distributed systems, Redis-based rate limiting would be implemented.

3. **No Webhook Retry Queue**: Webhooks are delivered synchronously. A message queue (RabbitMQ/Kafka) would be better for reliable delivery.

4. **No Event Deduplication**: Assuming events are idempotent. Adding deduplication would increase complexity.

5. **Single-Instance Deployment**: The current design assumes single-instance deployment. For horizontal scaling, we'd need distributed caching and message queues.

6. **Basic Search**: Implemented LIKE-based search. For production, Elasticsearch or Algolia would provide better search capabilities.

7. **No Metrics Export**: Prometheus metrics endpoint could be added for monitoring.

8. **Error Handling**: Implemented structured error codes with user-friendly messages. In production, consider adding more detailed error tracking (Sentry, DataDog).

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
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## Error Handling Best Practices

### Backend (Go)
- Structured error codes for different error types (Validation, Authentication, Not Found, Rate Limit, Server)
- Panic recovery middleware prevents crashes
- Security headers (XSS protection, HSTS)
- Request logging with timing

### Frontend (React + TypeScript)
- Error categorization (network, authentication, validation)
- User-friendly error messages (no technical jargon)
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
