# Cobutler

NOTE: This is all vibe code... dont use it please! Just experimenting

Cobutler is a Markov chain-based text generation system that learns from text input and can generate contextually relevant replies.

## Project Structure

The project follows standard Go project layout:

```
cobutler/
├── cmd/
│   └── cobutler/          # Main application executable
│       └── main.go        # Entry point for HTTP server
├── pkg/
│   └── cobutler/          # Core packages for reuse in other projects
│       ├── api/           # HTTP API handlers and server
│       │   ├── handlers.go
│       │   └── server.go
│       ├── db/            # Database interaction
│       │   └── graph.go
│       └── models/        # Domain models
│           ├── brain.go
│           └── tokenizer.go
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
└── README.md              # This file
```

## Usage

### Running the Server

```bash
# Set the brain database path (optional, defaults to "brain.db")
export COBUTLER_DB=/path/to/brain.db

# Set the port (optional, defaults to 8080)
export PORT=8080

# Run the server
go run cmd/cobutler/main.go
```

### API Endpoints

#### Learn from Text

```
POST /learn
Content-Type: application/json

{
  "text": "The text to learn from"
}
```

#### Generate a Reply

```
POST /predict
Content-Type: application/json

{
  "text": "Text to generate a contextual reply for"
}
```

Response:

```json
{
  "reply": "Generated reply based on learned patterns"
}
```

## Using as a Library

You can use Cobutler in your own Go projects:

```go
import (
    "github.com/kirkegaard/cobutler/pkg/cobutler/models"
)

func main() {
    // Initialize a brain
    brain, err := models.NewBrain("brain.db")
    if err != nil {
        panic(err)
    }
    defer brain.Close()

    // Learn from text
    err = brain.Learn("Text to learn from")
    if err != nil {
        panic(err)
    }

    // Generate a reply
    reply, err := brain.Reply("Input text")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(reply)
}
```

## License

[Add your license information here] 
