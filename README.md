# ZenTube

A minimalist YouTube search interface built with Go, HTMX, and Templ. Designed to help you find videos without getting lost in the endless rabbit hole of recommendations.

## 💡 Motivation

YouTube's recommendation algorithm is designed to maximize engagement, which often means losing hours to suggested videos you never intended to watch. This project was born from a real frustration: wanting to quickly search for specific content without getting distracted by the endless stream of recommended videos.

ZenTube provides a clean, focused interface - just search, find what you need, and move on. No distractions, no wasted time.

Additionally, this project serves as **reference scaffolding** for developers learning the Go + HTMX + Templ stack, demonstrating clean architecture patterns and modern web development practices.

> **Note**: This is a work-in-progress. Future additions will include database integration, Docker setup, and additional features.

## ✨ Features

- 🔍 Clean, distraction-free YouTube search interface
- 🚫 No recommendations, no algorithmic rabbit holes
- ⚡ HTMX-powered SPA-like experience without JavaScript frameworks
- 🎨 Server-side rendering with type-safe Templ templates
- 🏗️ **Clean Architecture** (Hexagonal/Ports & Adapters pattern)
- 🧪 Comprehensive test coverage with mocks
- 🔥 Hot reload development workflow with Air
- 🚀 Production-ready error handling and graceful shutdown

## 🏛️ Architecture Highlights

### Clean Architecture (Hexagonal)

```
┌─────────────────────────────────────────────────┐
│             External Interfaces                 │
│  (HTTP Handlers, YouTube API Client)           │
│            /adapters/http                       │
│            /adapters/youtube                    │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│              Port Interfaces                    │
│         (Dependency Inversion)                  │
│              /ports                             │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│            Business Logic                       │
│         (Framework Agnostic)                    │
│            /usecases                            │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│           Domain Entities                       │
│         (Pure Business Objects)                 │
│            /entities                            │
└─────────────────────────────────────────────────┘
```

### Key Patterns Demonstrated

- **Dependency Injection**: All dependencies injected via constructors
- **Interface Segregation**: Small, focused port interfaces
- **Testability**: Business logic fully unit tested with mocks
- **Separation of Concerns**: Clear boundaries between layers
- **Configuration Management**: Env vars + YAML with proper injection

## 🛠️ Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Web Framework** | [Gin](https://gin-gonic.com/) | Fast HTTP router |
| **Templating** | [Templ](https://templ.guide/) | Type-safe Go templates |
| **Interactivity** | [HTMX](https://htmx.org/) | Dynamic UI without JS frameworks |
| **External API** | YouTube Data API v3 | Video search |
| **Testing** | testify/mock | Unit tests with mocks |
| **Dev Tools** | Air | Hot reload |

## 📁 Project Structure

```
zentube/
├── cmd/zentube/          # Application entry point
│   └── main.go           # Server setup, graceful shutdown
├── internal/
│   ├── adapters/         # External adapters (infrastructure)
│   │   ├── http/         # HTTP layer (Gin handlers, middleware, routes)
│   │   └── youtube/      # YouTube API client implementation
│   ├── entities/         # Domain entities (pure business objects)
│   ├── ports/            # Port interfaces (dependency inversion)
│   ├── usecases/         # Business logic (framework-agnostic)
│   └── config/           # Configuration management
├── web/
│   ├── templates/        # Templ components
│   │   ├── layouts/      # Page layouts
│   │   ├── pages/        # Full page templates
│   │   └── components/   # Reusable UI components
│   └── static/           # CSS, JS, assets
└── configs/              # Configuration files

```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- [Templ](https://templ.guide/) - `go install github.com/a-h/templ/cmd/templ@latest`
- [Air](https://github.com/air-verse/air) (optional, for hot reload)
- YouTube Data API v3 key ([Get one here](https://console.cloud.google.com/apis/credentials))

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/uiansol/zentube.git
   cd zentube
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Set up environment variables**
   
   Create a `.env` file in the project root:
   ```bash
   YOUTUBE_API_KEY=your_api_key_here
   ```

4. **Generate Templ templates**
   ```bash
   make templ
   ```

5. **Run the application**
   ```bash
   # Development mode (with hot reload)
   make dev

   # Production mode
   make run
   ```

6. **Visit** `http://localhost:8080`

## 📝 Available Commands

```bash
make help           # Show all available commands
make deps           # Install dependencies
make templ          # Generate templ files
make build          # Build the application
make run            # Build and run
make dev            # Run with hot reload (Air)
make test           # Run tests
make test-coverage  # Run tests with coverage report
make clean          # Clean build artifacts
make fmt            # Format code (Go + Templ)
```

## 🧪 Testing

The project demonstrates proper testing practices:

```bash
# Run all tests
make test

# Generate coverage report
make test-coverage
# Opens coverage.html in your browser
```

Tests use the **testify/mock** library to mock external dependencies, ensuring business logic is tested in isolation.

## 🔮 Roadmap

- [ ] Database integration (SQLite)
- [ ] Docker and Docker Compose setup
- [ ] User favorites
- [ ] Comprehensive technical article

## 📚 Learning Resources

This project is designed to teach:

- **Clean Architecture** in Go
- **HTMX** for modern, server-driven UIs
- **Templ** for type-safe templating
- **Dependency Injection** without frameworks
- **Testing** with mocks and interfaces
- **Graceful Shutdown** patterns
- **Configuration Management** best practices

## 📖 Article (Coming Soon)

A detailed technical article explaining the architecture, design decisions, and patterns used in this project will be published soon.

## 🤝 Contributing

This is primarily a learning/portfolio project, but suggestions and feedback are welcome! Feel free to open issues or submit PRs.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with inspiration from Uncle Bob's Clean Architecture
- HTMX for making server-side rendering cool again
- The Go community for excellent tooling

---

**Note**: This project is under active development. The current implementation focuses on demonstrating clean architecture patterns. Database features and Docker configuration will be added in future iterations.

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
