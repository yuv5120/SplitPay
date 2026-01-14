# Split-It Backend (Golang)

A RESTful API backend built with Go (Golang) for the Split-It expense tracking application.

## Tech Stack

- **Go 1.21+** - Programming language
- **Fiber v2** - Fast HTTP web framework
- **MongoDB** - Database
- **Firebase Admin SDK** - Authentication
- **godotenv** - Environment variable management

## Features

- Firebase Authentication integration
- MongoDB database with Mongoose-like operations
- RESTful API endpoints for users and groups
- Expense tracking and management
- JWT token verification
- CORS support
- Rate limiting
- Security headers with Helmet

## Prerequisites

- Go 1.21 or higher
- MongoDB (local or Atlas)
- Firebase project with service account key

## Installation

1. **Clone the repository** (if not already done)

2. **Navigate to backend directory**
   ```bash
   cd backend
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Set up environment variables**
   
   Copy `.env.example` to `.env` and fill in your values:
   ```bash
   cp .env.example .env
   ```

   Required environment variables:
   - `MONGODB_URI` - Your MongoDB connection string
   - `FIREBASE_SERVICE_ACCOUNT_PATH` - Path to Firebase service account JSON file
   - `PORT` - Server port (default: 5000)
   - `CLIENT_URL` - Frontend URL for CORS (default: http://localhost:3000)

5. **Add Firebase Service Account**
   
   Download your Firebase service account key from Firebase Console and save it as `firebase-service-account.json` in the backend directory.

## Running the Server

### Development Mode
```bash
go run main.go
```

### Build and Run
```bash
# Build the application
go build -o split-it-backend

# Run the built binary
./split-it-backend
```

### With Air (Live Reload)
Install Air for live reloading during development:
```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with Air
air
```

## API Endpoints

### Health Check
- `GET /health` - Server health status

### User Routes
- `POST /api/users/profile` - Get or create user profile (requires auth)
- `PUT /api/users/profile` - Update user profile (requires auth)

### Group Routes
- `GET /api/groups` - Get all groups for user (requires auth)
- `GET /api/groups/:groupId` - Get single group (requires auth)
- `POST /api/groups` - Create new group (requires auth)
- `PUT /api/groups/:groupId` - Update group (requires auth)
- `DELETE /api/groups/:groupId` - Delete group (requires auth)

### Expense Routes
- `POST /api/groups/:groupId/expenses` - Add expense to group (requires auth)
- `DELETE /api/groups/:groupId/expenses/:expenseId` - Delete expense (requires auth)

## Authentication

All protected routes require a Firebase ID token in the Authorization header:
```
Authorization: Bearer <firebase-id-token>
```

## Project Structure

```
backend/
├── main.go                 # Application entry point
├── config/
│   ├── database.go        # MongoDB connection
│   └── firebase.go        # Firebase Admin SDK setup
├── models/
│   ├── user.go           # User model
│   └── group.go          # Group model
├── middleware/
│   └── auth.go           # Authentication middleware
├── routes/
│   ├── users.go          # User routes
│   └── groups.go         # Group routes
├── go.mod                # Go module file
├── go.sum                # Go dependencies checksum
├── .env                  # Environment variables (not in git)
└── .env.example          # Example environment variables
```

## Database Schema

### Users Collection
```json
{
  "_id": "ObjectId",
  "firebaseUid": "string",
  "email": "string",
  "name": "string",
  "createdAt": "Date",
  "updatedAt": "Date"
}
```

### Groups Collection
```json
{
  "_id": "ObjectId",
  "id": "string",
  "name": "string",
  "userId": "string",
  "members": [
    {
      "id": "string",
      "name": "string"
    }
  ],
  "expenses": [
    {
      "id": "string",
      "description": "string",
      "amount": "number",
      "paidBy": "string",
      "participants": ["string"],
      "date": "Date"
    }
  ],
  "createdAt": "Date",
  "updatedAt": "Date"
}
```

## Development

### Code Formatting
```bash
go fmt ./...
```

### Run Tests
```bash
go test ./...
```

### Build for Production
```bash
# Build for current platform
go build -o split-it-backend

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o split-it-backend-linux

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o split-it-backend.exe
```

## Deployment

### Docker (Optional)
Create a `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/firebase-service-account.json .
EXPOSE 5000
CMD ["./main"]
```

Build and run:
```bash
docker build -t split-it-backend .
docker run -p 5000:5000 split-it-backend
```

## Differences from Node.js Version

- Uses Go's native goroutines for concurrency
- Fiber framework instead of Express
- MongoDB driver with BSON instead of Mongoose
- More type-safe with Go's static typing
- Better performance and lower memory footprint
- Compiled binary for deployment

## Troubleshooting

### MongoDB Connection Issues
- Verify `MONGODB_URI` is correct
- Check if MongoDB is running
- For Atlas, ensure IP whitelist is configured

### Firebase Authentication Issues
- Verify service account JSON file exists
- Check `FIREBASE_SERVICE_ACCOUNT_PATH` in `.env`
- Ensure Firebase project is properly configured

### CORS Issues
- Verify `CLIENT_URL` matches your frontend URL
- Check if frontend is sending credentials correctly

## License

MIT
