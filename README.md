# TaxiFleet Backend

Go + Gin backend API for the TaxiFleet Management System.

## Features

- RESTful API with Gin framework
- PostgreSQL database with proper migrations (golang-migrate)
- JWT authentication with refresh tokens
- Multi-tenant architecture
- Role-based access control (Owner, Manager, Mechanic, Driver)
- Weekly report management
- Expense tracking
- Bank deposit tracking
- CSV export functionality
- Graceful shutdown
- Structured logging with logrus
- Connection pooling and database health checks

## Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose (optional)

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. Start PostgreSQL (using Docker):
```bash
docker-compose up -d postgres
```

4. Run the application:
```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

### Using Docker Compose

```bash
docker-compose up
```

This will start both PostgreSQL and the API server.

## Configuration

All configuration is loaded from environment variables or a `.env` file. See `.env.example` for all available options.

Key configuration sections:
- **Server**: Port, host, timeouts, environment
- **Database**: Connection details, pool settings, migration path
- **JWT**: Secret, expiration times
- **Security**: BCrypt cost, rate limiting, CORS
- **Logging**: Level, format, output

## Database Migrations

Migrations are automatically run on application startup using `golang-migrate`.

Migration files are located in the `migrations/` directory:
- `001_initial_schema.up.sql` - Creates all tables
- `001_initial_schema.down.sql` - Drops all tables
- `002_seed_default_tenant.up.sql` - Placeholder for seed data (use seed script instead)

To create a new migration:
```bash
migrate create -ext sql -dir migrations -seq <migration_name>
```

## Seed Data

To create default test data (tenant, users, and taxis), run the seed script:

```bash
go run cmd/seed/main.go
```

This will create:
- **Tenant**: Gnakpa Transport
- **Users**:
  - Owner: tanguy.gnakpa@gnakpa-transport.com / `owner123`
  - Manager: manager@gnakpa-transport.com / `manager123`
  - Mechanic: mechanic@gnakpa-transport.com / `mechanic123`
  - Driver: driver@gnakpa-transport.com / `driver123`
- **Test Taxis**: 3 sample taxis for testing

See `SEED_DATA.md` for more details.

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Get current user

### Taxis
- `GET /api/v1/taxis` - List all taxis
- `POST /api/v1/taxis` - Create taxi
- `GET /api/v1/taxis/:id` - Get taxi by ID
- `PUT /api/v1/taxis/:id` - Update taxi
- `DELETE /api/v1/taxis/:id` - Delete taxi

### Reports
- `GET /api/v1/reports` - List reports
- `POST /api/v1/reports` - Create report
- `GET /api/v1/reports/:id` - Get report by ID
- `PUT /api/v1/reports/:id` - Update report
- `POST /api/v1/reports/:id/submit` - Submit report
- `POST /api/v1/reports/:id/approve` - Approve report
- `POST /api/v1/reports/:id/reject` - Reject report

### Deposits
- `GET /api/v1/deposits` - List deposits
- `POST /api/v1/deposits` - Create deposit
- `GET /api/v1/deposits/:id` - Get deposit by ID
- `PUT /api/v1/deposits/:id` - Update deposit
- `DELETE /api/v1/deposits/:id` - Delete deposit

### Expenses
- `GET /api/v1/expenses` - List expenses
- `POST /api/v1/expenses` - Create expense
- `GET /api/v1/expenses/:id` - Get expense by ID
- `PUT /api/v1/expenses/:id` - Update expense
- `DELETE /api/v1/expenses/:id` - Delete expense

### Export
- `GET /api/v1/export/reports?format=csv` - Export reports
- `GET /api/v1/export/expenses?format=csv` - Export expenses

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── database/             # Database connection and migrations
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # Middleware (auth, CORS)
│   ├── repository/          # Database access layer
│   └── service/             # Business logic layer
├── migrations/              # SQL migration files
├── docker-compose.yml       # Docker setup
├── Dockerfile              # Docker image
└── go.mod                  # Go dependencies
```

## Database Schema

The application uses PostgreSQL with proper migrations. The schema includes:
- Tenants (multi-tenant support)
- Users (with roles)
- Sessions (JWT token management)
- Taxis (vehicle management)
- Weekly Reports (driver reports)
- Expenses (expense tracking)
- Bank Deposits (deposit records)
- Maintenance Logs (vehicle maintenance)

## Security

- JWT tokens with configurable expiration
- Refresh tokens with longer expiration
- Password hashing with bcrypt (configurable cost)
- Tenant isolation for all queries
- Role-based access control
- CORS configuration

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/api ./cmd/api
```

### Logging

The application uses structured logging with logrus. Logs are output in JSON format by default and can be configured via environment variables.

## Production Considerations

1. Set a strong `JWT_SECRET` in production
2. Configure proper CORS origins
3. Set `ENVIRONMENT=production`
4. Configure database connection pooling appropriately
5. Set up proper log rotation
6. Use HTTPS in production
7. Configure rate limiting appropriately
