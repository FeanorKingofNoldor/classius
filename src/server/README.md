# Classius Server API

A comprehensive Go-based REST API server for the Classius classical education platform.

## ✨ Features

- **User Authentication**: JWT-based auth with refresh tokens
- **Book Management**: Upload, organize, and manage classical texts
- **Annotation Sync**: Bidirectional synchronization between devices
- **Reading Progress**: Track and sync reading progress across devices
- **AI Sage Integration**: Classical education AI assistant
- **Community Features**: Discussions and note sharing
- **Multi-device Support**: Session management across devices

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis (optional, for caching)
- Docker & Docker Compose

### Development Setup

1. **Start Services**
   ```bash
   # Start PostgreSQL and Redis via Docker
   docker-compose -f ../docker/docker-compose.dev.yml up -d
   ```

2. **Configure Environment**
   ```bash
   # Copy config and set environment variables
   cp config.yaml config.local.yaml
   export DATABASE_PASSWORD=your_password
   export JWT_SECRET=your_jwt_secret
   export OPENAI_API_KEY=your_openai_key
   ```

3. **Run Server**
   ```bash
   # Install dependencies
   go mod tidy
   
   # Run migrations and start server
   go run cmd/server/main.go
   ```

Server will start at `http://localhost:8080`

## 📚 API Documentation

### Authentication

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "scholar123",
  "email": "scholar@example.com",
  "password": "securepassword123",
  "full_name": "Classical Scholar"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "scholar@example.com", 
  "password": "securepassword123",
  "device_id": "device-123",
  "device_name": "My E-Reader"
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}
```

### User Management

#### Get Profile
```http
GET /api/v1/user/profile
Authorization: Bearer {access_token}
```

#### Update Profile  
```http
PUT /api/v1/user/profile
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "full_name": "Updated Name",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

### Reading Progress

#### Get Progress
```http
GET /api/v1/user/progress?book_id={uuid}&page=1&limit=20
Authorization: Bearer {access_token}
```

#### Save Progress
```http
POST /api/v1/user/progress
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "book_id": "book-uuid",
  "current_page": 42,
  "total_pages": 200,
  "percentage": 0.21,
  "time_spent_minutes": 30
}
```

### Annotations

#### Get Annotations
```http
GET /api/v1/annotations?book_id={uuid}&type=highlight&page=1&limit=50
Authorization: Bearer {access_token}
```

#### Create Annotation
```http
POST /api/v1/annotations
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "book_id": "book-uuid",
  "type": "highlight",
  "page_number": 42,
  "start_position": 100,
  "end_position": 200,
  "selected_text": "Selected passage",
  "color": "#ffff00",
  "is_private": true
}
```

#### Update Annotation
```http
PUT /api/v1/annotations/{id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "content": "Updated note content",
  "color": "#ff0000",
  "tags": ["important", "philosophy"]
}
```

#### Sync Annotations
```http
POST /api/v1/annotations/sync
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "last_sync_time": "2024-01-01T00:00:00Z",
  "annotations": [
    {
      "book_id": "book-uuid",
      "type": "note",
      "page_number": 10,
      "content": "Important insight",
      "is_private": true
    }
  ]
}
```

### Books

#### Get Books
```http
GET /api/v1/books?page=1&limit=20&search=plato
Authorization: Bearer {access_token}
```

#### Upload Book
```http
POST /api/v1/books/upload
Authorization: Bearer {access_token}
Content-Type: multipart/form-data

file=@book.epub
title=The Republic
author=Plato
```

### AI Sage

#### Ask Sage
```http
POST /api/v1/sage/ask
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "question": "What is the main theme of Plato's Republic?",
  "book_id": "book-uuid",
  "passage_text": "Context from the book..."
}
```

#### Get Sage History
```http
GET /api/v1/sage/history?page=1&limit=20
Authorization: Bearer {access_token}
```

## 🏗️ Architecture

```
├── cmd/
│   └── server/           # Main application entry point
├── internal/
│   ├── db/              # Database connection & migrations
│   ├── handlers/        # HTTP request handlers
│   │   ├── auth.go      # Authentication endpoints
│   │   ├── user.go      # User management
│   │   ├── annotations.go # Annotation sync
│   │   └── books.go     # Book management
│   ├── middleware/      # HTTP middleware
│   ├── models/         # Database models (GORM)
│   └── services/       # Business logic services
├── config.yaml         # Configuration file
└── go.mod             # Dependencies
```

## 🔒 Security

- JWT tokens with configurable expiration
- Password hashing with bcrypt
- Rate limiting (Redis-based)
- CORS protection
- SQL injection prevention (GORM)
- Input validation and sanitization

## 📊 Database Schema

The server uses PostgreSQL with the following key tables:
- `users` - User accounts and profiles
- `books` - Book catalog and metadata  
- `user_books` - User's book library
- `reading_progress` - Reading progress tracking
- `annotations` - Highlights, notes, bookmarks
- `sage_conversations` - AI interaction history
- `user_sessions` - Multi-device session management

## 🔧 Configuration

Configuration via `config.yaml` or environment variables:

- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection (optional)
- `JWT_SECRET` - JWT signing secret
- `OPENAI_API_KEY` - OpenAI API key for Sage
- `PORT` - Server port (default: 8080)

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Integration tests (requires running services)
go test -tags=integration ./...
```

## 🚢 Deployment

### Docker
```bash
# Build image
docker build -t classius-server .

# Run container
docker run -p 8080:8080 -e DATABASE_URL=... classius-server
```

### Production Checklist
- [ ] Set strong JWT secret
- [ ] Configure production database
- [ ] Set up SSL/TLS termination
- [ ] Configure log aggregation  
- [ ] Set up monitoring & alerts
- [ ] Configure backup strategy

## 🤝 API Integration

The server is designed to work seamlessly with:
- **Qt/QML Device App**: Native e-reader interface
- **Web Dashboard**: Browser-based management
- **Mobile Apps**: iOS/Android companions  
- **CLI Tools**: Automation and bulk operations

## 📈 Performance

- Database connection pooling
- Redis caching layer
- Pagination for large datasets
- Optimized queries with proper indexes
- Rate limiting for API protection

## 🛟 Support

For issues and questions:
- Check the logs at `./logs/classius.log`
- Review configuration in `config.yaml`
- Verify database connectivity
- Check JWT token validity

## 🔄 Development

The server supports hot reloading in development:
```bash
# Install air for hot reloading
go install github.com/cosmtrek/air@latest

# Start with hot reload
air
```

---

**Built with Go, Gin, GORM, and PostgreSQL** 🚀