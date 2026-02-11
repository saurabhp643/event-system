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
│  │ Rate Limiter│  │ Event Store │  │ Nginx Proxy         │  │
│  │ (In-Memory) │  │(GORM/SQLite)│  │ (API + WebSocket)   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Tech Stack

- **Backend**: Go 1.21, Gin, GORM, SQLite, gorilla/websocket
- **Frontend**: React 18, TypeScript, Vite, TailwindCSS
- **Authentication**: JWT, API Keys
- **Docker**: Multi-stage builds, docker-compose
- **Nginx**: Reverse proxy for API and WebSocket
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
JWT_SECRET=your-super-secret-jwt-key-change-in-production
DATABASE_PATH=/app/data/events.db
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
| GET | `/ws` | WebSocket endpoint |
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

### Option 3: Render Full-Stack Docker (Recommended for simplicity)

Deploy both frontend and backend as a single Docker container with Nginx reverse proxy.

---

#### **Step 1: Create GitHub Repository**

1. **Create a new GitHub repository**
   - Go to [GitHub](https://github.com)
   - Click "New repository"
   - Name: `event-ingestion-system`
   - Make it **Public** or **Private**
   - Click "Create repository"

2. **Push your code to GitHub**
   ```bash
   cd event-ingestion-system
   
   # Initialize git if not already done
   git init
   git add .
   git commit -m "Initial commit: multi-tenant event ingestion system"
   
   # Add remote (replace with your username)
   git remote add origin https://github.com/YOURUSERNAME/event-ingestion-system.git
   
   # Push to GitHub
   git push -u origin main
   ```

---

#### **Step 2: Create Render Account**

1. **Go to [Render Dashboard](https://dashboard.render.com)**
2. **Sign up with GitHub**
   - Click "Sign Up"
   - Choose "Sign up with GitHub"
   - Authorize Render

---

#### **Step 3: Deploy with Blueprint (Easiest)**

1. **Click "New +" → "Blueprint"**
   ![Blueprint Setup](https://render.com/images/docs/blueprints/create-blueprint.png)

2. **Connect your GitHub repository**
   - Select your `event-ingestion-system` repository
   - Click "Connect"

3. **Configure the service**
   Render will automatically detect the `render.yaml` file and configure:

   | Setting | Value |
   |---------|-------|
   | Service Type | Web Service |
   | Environment | Docker |
   | Plan | Free |
   | Disk | 1 GB (auto-configured) |

4. **Review environment variables**
   - `JWT_SECRET`: Auto-generated (secure)
   - `DATABASE_PATH`: `/app/data/events.db`
   - `GIN_MODE`: `release`
   - `PORT`: `80`

5. **Click "Apply Blueprint"**

6. **Wait for deployment**
   - Build time: ~3-5 minutes
   - Watch the logs for progress

---

#### **Step 4: Get Your Live URL**

1. **After deployment completes**, you'll see:
   ```
   https://event-ingestion.onrender.com
   ```

2. **Test the deployment**
   - Open the URL in your browser
   - You should see the Event Dashboard
   - Create a tenant
   - Ingest an event
   - Watch it appear in real-time

---

#### **Alternative: Manual Web Service Creation**

If Blueprint doesn't work:

1. **Click "New +" → "Web Service"**
2. **Connect your GitHub repository**
3. **Configure:**
   | Setting | Value |
   |---------|-------|
   | Environment | Docker |
   | Plan | Free |
   | Build Command | (leave empty - uses Dockerfile) |
   | Start Command | (leave empty - uses ENTRYPOINT) |
4. **Add Disk:**
   - Click "Add Disk"
   - Mount Path: `/app/data`
   - Size: 1 GB
   - Name: `events-db`
5. **Add Environment Variables:**
   | Key | Value |
   |-----|-------|
   | JWT_SECRET | (click "Generate" or enter your own) |
   | DATABASE_PATH | `/app/data/events.db` |
   | GIN_MODE | `release` |
   | PORT | `80` |
6. **Click "Create Web Service"**

---

#### **Step 5: Verify Your Deployment**

**Test these endpoints:**

```bash
# Replace with your URL
export BASE_URL="https://your-app.onrender.com"

# Health check
curl $BASE_URL/health

# List tenants (create one first via UI)
curl $BASE_URL/api/v1/tenants

# Create a tenant
curl -X POST $BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Tenant"}'

# Ingest an event (replace TENANT_ID and API_KEY)
curl -X POST $BASE_URL/api/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "tenant_id": "YOUR_TENANT_ID",
    "event_type": "page_view",
    "timestamp": "2024-01-15T10:30:00Z",
    "metadata": {"page": "/home"}
  }'
```

---

#### **Step 6: View Logs (Troubleshooting)**

1. **Go to your Render dashboard**
2. **Click on your service**
3. **Click "Logs" tab**
4. **Filter by:**
   - `All logs` - see everything
   - `Errors` - see only errors
   - `Events` - see deployment events

**Common log messages:**
- `Starting Event Ingestion System...` - App is starting
- `Starting Go backend server...` - Backend starting
- `Backend server started (PID: ...)` - Backend running
- `Starting Nginx...` - Nginx starting
- `All services started. Application is ready.` - All good!

---

#### **Step 7: Update Your Deployment**

1. **Make changes to code**
2. **Commit and push to GitHub:**
   ```bash
   git add .
   git commit -m "Your commit message"
   git push origin main
   ```
3. **Render auto-deploys**
   - Watch the "Events" tab for progress
   - Deployment takes ~2-3 minutes

---

### Alternative Deployment Options

#### **Option 1: Render (Backend) + Vercel (Frontend)**

See detailed instructions above in the original README.

#### **Option 2: Docker + Railway**

1. **Create Railway Account**: [railway.app](https://railway.app)
2. **Deploy with Docker:**
   ```bash
   npm i -g railway
   railway login
   railway init
   railway add postgresql
   railway up
   ```

#### **Option 3: Google Cloud Run**

```bash
# Build and push to Google Container Registry
gcloud builds submit --tag gcr.io/PROJECT_ID/event-ingestion

# Deploy to Cloud Run
gcloud run deploy event-ingestion \
  --image gcr.io/PROJECT_ID/event-ingestion \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

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
├── Dockerfile              # Unified full-stack Dockerfile
├── render.yaml             # Render Blueprint configuration
├── docker-compose.yml      # Local Docker deployment
├── nginx.conf              # Nginx reverse proxy config
├── entrypoint.sh           # Container startup script
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
- User-friendly error messages (no technical notifications for feedback
 jargon)
- Toast- Retry logic with exponential backoff
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
