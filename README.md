# zentube

A minimalist YouTube search interface built with Go, HTMX, and Templ. Designed to reduce distractions while searching for videos.

## Features

- 🔍 Clean YouTube search interface
- ⚡ Fast HTMX-powered interactions (no page reloads)
- 🎨 Server-side rendered with Templ
- 🏗️ Clean architecture (Ports & Adapters pattern)
- 🔥 Hot reload development with Air

## Tech Stack

- **Backend**: Go + Gin
- **Frontend**: HTMX + Templ
- **API**: YouTube Data API v3
- **Architecture**: Clean Architecture (Hexagonal)

## Project Structure

```
zentube/
├── cmd/zentube/          # Application entry point
├── internal/
│   ├── adapters/         # External adapters
│   │   ├── http/         # HTTP handlers, routes, middleware
│   │   └── youtube/      # YouTube API client
│   ├── entities/         # Domain entities
│   ├── ports/            # Port interfaces
│   ├── usecases/         # Business logic
│   └── config/           # Configuration
├── web/
│   ├── templates/        # Templ components
│   │   ├── layouts/
│   │   ├── pages/
│   │   └── components/
│   └── static/           # Static assets (CSS, JS)
└── configs/              # Config files

```

## Prerequisites

- Go 1.21+
- [Templ](https://templ.guide/) - `go install github.com/a-h/templ/cmd/templ@latest`
- [Air](https://github.com/air-verse/air) (optional, for hot reload)
- YouTube Data API v3 key

## Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/uiansol/zentube.git
   cd zentube
   ```

2. **Install dependencies**
   ```bash
   make deps
   # or manually:
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env and add your YouTube API key
   ```

4. **Generate templates**
   ```bash
   make templ
   ```

5. **Run the application**
   ```bash
   # Development (with hot reload)
   make dev

   # Production
   make build
   ./zentube
   ```

## Available Commands

```bash
make help           # Show all available commands
make deps           # Install dependencies
make templ          # Generate templ files
make build          # Build the application
make run            # Build and run
make dev            # Run with hot reload
make test           # Run tests
make test-coverage  # Run tests with coverage
make clean          # Clean build artifacts
make fmt            # Format code
```

## Configuration

Configuration is managed through `configs/config.yaml` and environment variables:

```yaml
app:
  port: 8080

youtube:
  api_key: ${YOUTUBE_API_KEY}
  max_results: 10
```

Environment variables override config file values.

## Development

The project uses Air for hot reloading during development:

```bash
air
```

Any changes to `.go`, `.templ`, or `.yaml` files will automatically rebuild and restart the server.

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

## Architecture

This project follows **Clean Architecture** principles:

- **Entities**: Core business objects (`Video`)
- **Use Cases**: Business logic (`SearchVideos`)
- **Ports**: Interfaces for external systems (`YouTubeClient`)
- **Adapters**: Implementations of ports (HTTP handlers, YouTube API client)

Benefits:
- ✅ Testable (easy to mock dependencies)
- ✅ Independent of frameworks
- ✅ Flexible and maintainable

## License

MIT License - see [LICENSE](LICENSE) file

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
