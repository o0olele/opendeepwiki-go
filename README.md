# OpenDeepWiki

OpenDeepWiki is an AI-driven documentation system that automatically generates comprehensive documentation for code repositories.

## Features

- Submit Git repositories for automatic documentation generation
- Asynchronous processing of documentation tasks
- RESTful API for task submission and status checking
- YAML-based configuration
- Modular API structure
- SQLite database for persistent storage
- Automatic task recovery after server restart

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/o0olele/opendeepwiki-go.git
cd opendeepwiki-go
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the project:
```bash
go build -o opendeepwiki cmd/server/main.go
```

### Configuration

The application uses a YAML configuration file. By default, it looks for `config.yaml` in the current directory. You can specify a different configuration file using the `CONFIG_FILE` environment variable.

Example configuration file:
```yaml
# OpenDeepWiki Configuration

# Server settings
server:
  address: ":8080"
  
# Repository settings
repository:
  dir: "./repos"

# Database settings
database:
  path: "./data/opendeepwiki.db"
```

### Running the Server

```bash
./opendeepwiki
```

The server will start on port 8080 by default, or as specified in the configuration file.

## Task Recovery

OpenDeepWiki automatically recovers unfinished tasks when the server restarts. This ensures that:

1. No tasks are lost due to server crashes or restarts
2. Long-running tasks can be resumed from where they left off
3. System resources are not wasted on duplicate processing

When the server starts, it:
1. Queries the database for all tasks with status other than "completed" or "failed"
2. For each pending task, it checks if the repository has already been cloned
3. If the repository exists and is valid, it continues from the next step
4. If the repository doesn't exist or is invalid, it restarts the task

## Database Structure

OpenDeepWiki uses SQLite for persistent storage with the following tables:

- **repositories**: Stores information about Git repositories
- **tasks**: Tracks documentation generation tasks
- **documents**: Stores generated documentation
- **code_analyses**: Stores code analysis results

The database file is created automatically at the path specified in the configuration file.

## API Usage

### Submit a Repository

```bash
curl -X POST http://localhost:8080/api/warehouse/repos \
  -H "Content-Type: application/json" \
  -d '{"git_url": "https://github.com/username/repo.git"}'
```

Response:
```json
{
  "message": "Repository submitted for processing",
  "task_id": "task_1234567890"
}
```

If the repository is already being processed:
```json
{
  "existing": true,
  "message": "Repository already being processed",
  "status": "cloned",
  "task_id": "task_1234567890"
}
```

### Check Task Status

```bash
curl http://localhost:8080/api/warehouse/tasks/task_1234567890
```

Response:
```json
{
  "task_id": "task_1234567890",
  "git_url": "https://github.com/username/repo.git",
  "status": "cloned",
  "created_at": "2023-06-15T10:30:45Z"
}
```

## API Structure

The API is organized into modules:

- `/api/warehouse/*` - Repository management and task status endpoints

## API Testing Tools

The project includes several tools to test the API in the `tests/api` directory:

### HTTP Request File

Use the `tests/api/api_tests.http` file with REST Client extensions in VS Code or JetBrains IDEs to test the API directly from your editor.

### Shell Script

Run the `tests/api/test_api.sh` script (Linux/macOS) to test the API using curl commands:

```bash
cd tests/api
chmod +x test_api.sh
./test_api.sh
```

### PowerShell Script

For Windows users, run the `tests/api/test_api.ps1` script to test the API:

```powershell
cd tests/api
.\test_api.ps1
```

### HTML Test Page

Open the `tests/api/api_test.html` file in a web browser to test the API using a simple web interface.

### Task Recovery Testing

To test the task recovery functionality:

```bash
# Linux/macOS
cd tests/api
chmod +x test_recovery.sh
./test_recovery.sh

# Windows
cd tests/api
.\test_recovery.ps1
```

For more details about the testing tools, see the [tests/api/README.md](tests/api/README.md) file.

## Environment Variables

- `CONFIG_FILE`: Path to the configuration file (default: `config.yaml`)

## License

This project is licensed under the MIT License - see the LICENSE file for details. 