# Mancala Game Services

A distributed Mancala game implementation built with Go and gRPC, featuring microservices architecture with engine, games, and matchmaking services, Redis state management, and event streaming.

## Architecture

The system consists of eight main components:

- **Engine Service** (`cmd/engine`): Core game logic and move processing
- **Games Service** (`cmd/games`): Game session management, player validation, and Redis integration
- **Matchmaking Service** (`cmd/matchmaking`): Player queue management and automatic game matching
- **Bot Service** (`cmd/bot`): AI opponents with three difficulty levels (Easy, Medium, Hard)
- **Auth Service** (`cmd/auth`): User authentication and JWT token management
- **Notifications Service** (`cmd/notifications`): Real-time event notifications via Server-Sent Events
- **API Gateway** (`cmd/gateway`): HTTP REST API gateway providing unified client access
- **CLI Client** (`cmd/mancala`): Command-line client for playing games
- **Redis**: Persistent storage for game sessions and event streaming via Redis Streams

```
┌─────────────┐     HTTP      ┌─────────────┐     gRPC     ┌─────────────┐
│ CLI Client  │ ────────────► │ API Gateway │ ────────────► │ Auth Service│
│  (mancala)  │               │   (HTTP)    │               │   (JWT)     │
└─────────────┘               └─────────────┘               └─────────────┘
                                     │
                                     │ gRPC
                                     ▼
                              ┌─────────────┐     gRPC     ┌─────────────┐
                              │   Games     │ ────────────► │   Engine    │
                              │  Service    │               │  Service    │
                              └─────────────┘               └─────────────┘
                                     │                             │
                                     │                             │
                              ┌─────────────┐               ┌─────────────┐
                              │Matchmaking  │◄─────────────►│ Bot Service │
                              │  Service    │     gRPC      │ (AI Engine) │
                              └─────────────┘               └─────────────┘
                                     │                             │
                                     │                       ┌─────────────┐
                                     │                       │Notifications│
                                     │                       │  Service    │
                                     │                       └─────────────┘
                                     │ Redis                       │ SSE
                                     ▼                             ▼
                              ┌─────────────┐               ┌─────────────┐
                              │    Redis    │               │   Events    │
                              │ + Streams   │◄──────────────│   Stream    │
                              └─────────────┘               └─────────────┘
```

## Features

- **Stateless Engine**: Pure game logic with move validation and processing
- **Stateful Games Service**: Session management with Redis persistence
- **Intelligent Matchmaking**: FIFO queue-based player matching with automatic game creation
- **AI Bot Opponents**: Three difficulty levels with sophisticated game AI
  - **Easy**: Random valid moves, perfect for beginners
  - **Medium**: Strategic play with captures and extra turns
  - **Hard**: Advanced minimax algorithm with alpha-beta pruning
- **Event Streaming**: Redis Streams for real-time game events and notifications
- **Player Authentication**: Validates players belong to games and turns
- **Automatic Cleanup**: Removes finished games from storage
- **gRPC Interface**: High-performance protocol buffer communication
- **Comprehensive Testing**: Unit and integration tests with concurrent access validation
- **Containerized**: Docker images for easy deployment
- **Kubernetes Ready**: Complete K8s manifests with private registry support
- **JWT Authentication**: Secure user authentication with token-based authorization
- **Real-time Notifications**: Server-Sent Events for live game updates
- **HTTP REST API**: Gateway providing unified access to all services
- **CLI Client**: Full-featured command-line interface for gameplay

## Quick Start

### Prerequisites

- Go 1.24+
- Docker
- Kubernetes cluster (optional)
- Redis (for games service)

### CLI Client (Recommended)

The easiest way to play Mancala is using the command-line client:

1. **Build the client**:
   ```bash
   go build -o mancala cmd/mancala/main.go
   ```

2. **Connect to server**:
   ```bash
   ./mancala connect <server-ip>
   ```

3. **Create account**:
   ```bash
   ./mancala register
   ```

4. **Play against other players**:
   ```bash
   ./mancala play
   ```

5. **Play against AI bots**:
   ```bash
   ./mancala bot easy     # Beginner-friendly
   ./mancala bot medium   # Balanced challenge
   ./mancala bot hard     # Advanced AI
   ```

📚 **Full CLI documentation**: [docs/CLI_CLIENT.md](docs/CLI_CLIENT.md)

### Local Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/laerson/mancala
   cd mancala
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Generate protobuf code** (if modified):
   ```bash
   protoc --proto_path=. \
     --go_out=. --go_opt=paths=source_relative \
     --go-grpc_out=. --go-grpc_opt=paths=source_relative \
     proto/engine/engine.proto proto/games/games.proto proto/matchmaking/matchmaking.proto
   ```

4. **Run tests**:
   ```bash
   go test ./...
   ```

5. **Build services**:
   ```bash
   go build ./cmd/engine
   go build ./cmd/games
   go build ./cmd/matchmaking
   go build ./cmd/bot
   go build ./cmd/auth
   go build ./cmd/notifications
   go build ./cmd/gateway
   go build ./cmd/mancala
   ```

### Local Services Setup

1. **Start Redis**:
   ```bash
   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

2. **Start Engine Service**:
   ```bash
   ./engine
   # Runs on port 50051
   ```

3. **Start Games Service**:
   ```bash
   ./games
   # Runs on port 50052, connects to Redis on localhost:6379 and Engine on localhost:50051
   ```

4. **Start Bot Service**:
   ```bash
   ./bot
   # Runs on port 50057, provides AI opponents with three difficulty levels
   ```

5. **Start Matchmaking Service**:
   ```bash
   ./matchmaking
   # Runs on port 50054, connects to Redis on localhost:6379, Games on localhost:50052, and Bot on localhost:50057
   ```

## Infrastructure as Code Deployment

### Prerequisites

- **OpenStack Cloud Access** with RC file sourced
- **Required Tools**: Terraform, Ansible, kubectl
- **SSH Key Pair**: For VM access

### Quick Deployment

1. **Navigate to IaC directory**:
   ```bash
   cd iac
   ```

2. **Configure variables**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your OpenStack settings
   ```

3. **Source OpenStack credentials**:
   ```bash
   source your-openstack-rc.sh
   ```

4. **Deploy complete infrastructure**:
   ```bash
   ./deploy.sh deploy
   ```

This will:
- ✅ Create OpenStack VM with Ubuntu 22.04
- ✅ Configure security groups and networking
- ✅ Install and configure single-node Kubernetes cluster
- ✅ Deploy Redis, Engine, Games, Matchmaking, Bot, Auth, and Gateway services
- ✅ Expose API Gateway via NodePort (30080) for unified client access
- ✅ Configure private GitHub Container Registry access

### IaC Configuration Options

**terraform.tfvars** example:
```hcl
cluster_name          = "mancala-k8s"
environment          = "production"
image_name           = "Ubuntu 22.04"
flavor_name          = "m1.large"
network_name         = "private"
external_network_name = "public"
public_key_path      = "~/.ssh/id_rsa.pub"
private_key_path     = "~/.ssh/id_rsa"
ssh_user            = "ubuntu"
```

### Deployment Commands

```bash
# Preview changes
./deploy.sh plan

# Deploy everything
./deploy.sh deploy

# Deploy only Kubernetes cluster (assumes VM exists)
./deploy.sh k8s-only

# Deploy only application (assumes cluster exists)
./deploy.sh app-only

# Destroy all infrastructure
./deploy.sh destroy
```

### Access Your Deployed Services

After deployment:
```bash
# Get cluster info
cd terraform && terraform output

# SSH to master node
ssh -i ~/.ssh/id_rsa ubuntu@<master-ip>

# Copy kubectl config locally
scp -i ~/.ssh/id_rsa ubuntu@<master-ip>:~/.kube/config ~/.kube/config

# Access services
# API Gateway - External: <master-ip>:30080 (HTTP REST API)
# Internal access: kubectl port-forward -n mancala svc/gateway 8080:8080
```

## API Reference

### Engine Service (port 50051)

**Move Request**:
```protobuf
service Engine {
  rpc Move(MoveRequest) returns (MoveResponse);
}

message MoveRequest {
  GameState game_state = 1;
  uint32 pit_index = 2;
}
```

### Games Service (port 50052)

**Create Game**:
```protobuf
service Games {
  rpc Create(CreateGameRequest) returns (CreateGameResponse);
  rpc Move(MakeGameMoveRequest) returns (MakeGameMoveResponse);
}

message CreateGameRequest {
  string player1_id = 1;
  string player2_id = 2;
}
```

**Make Move**:
```protobuf
message MakeGameMoveRequest {
  string player_id = 1;
  string game_id = 2;
  uint32 pit_index = 3;
}
```

### Matchmaking Service (port 50054)

**Enqueue Player**:
```protobuf
service Matchmaking {
  rpc Enqueue(EnqueueRequest) returns (EnqueueResponse);
  rpc BotMatch(BotMatchRequest) returns (BotMatchResponse);
  rpc CancelQueue(CancelQueueRequest) returns (CancelQueueResponse);
  rpc GetQueueStatus(GetQueueStatusRequest) returns (GetQueueStatusResponse);
  rpc StreamUpdates(StreamUpdatesRequest) returns (stream UpdateEvent);
}

message EnqueueRequest {
  Player player = 1;
}
```

**Bot Match Creation**:
```protobuf
message BotMatchRequest {
  Player player = 1;
  string bot_difficulty = 2; // "easy", "medium", "hard"
}

message BotMatchResponse {
  bool success = 1;
  string game_id = 2;
  string message = 3;
  string bot_id = 4;
  string bot_name = 5;
}
```

**Queue Management**:
```protobuf
message CancelQueueRequest {
  string player_id = 1;
  string queue_id = 2;
}

message GetQueueStatusRequest {
  string player_id = 1;
}
```

### Bot Service (port 50057)

**Get Bot Move**:
```protobuf
service Bot {
  rpc GetMove(GetMoveRequest) returns (GetMoveResponse);
  rpc ListBots(ListBotsRequest) returns (ListBotsResponse);
  rpc CreateBot(CreateBotRequest) returns (CreateBotResponse);
}

message GetMoveRequest {
  string bot_id = 1;
  GameState game_state = 2;
  BotDifficulty difficulty = 3;
}

message GetMoveResponse {
  oneof result {
    MoveResult move = 1;
    Error error = 2;
  }
}

message MoveResult {
  uint32 pit_index = 1;
  string reasoning = 2;
  int32 evaluation_score = 3;
}
```

**Bot Management**:
```protobuf
message CreateBotRequest {
  BotDifficulty difficulty = 1;
  string name_suffix = 2;
}

message BotProfile {
  string id = 1;
  string name = 2;
  BotDifficulty difficulty = 3;
  string description = 4;
  int32 wins = 5;
  int32 losses = 6;
}

enum BotDifficulty {
  BOT_DIFFICULTY_UNSPECIFIED = 0;
  BOT_DIFFICULTY_EASY = 1;        // Random valid moves
  BOT_DIFFICULTY_MEDIUM = 2;      // Basic strategy (captures, extra turns)
  BOT_DIFFICULTY_HARD = 3;        // Advanced AI (minimax with pruning)
}
```

### API Gateway (port 8080)

**Bot Match HTTP Endpoint**:
```http
POST /api/v1/matchmaking/bot
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "player_id": "user123",
  "player_name": "Alice",
  "bot_difficulty": "medium"
}
```

**Response**:
```json
{
  "success": true,
  "game_id": "game-uuid-123",
  "message": "Bot match created! Playing against Strategic Bot",
  "bot_id": "bot-abc123",
  "bot_name": "Strategic Bot"
}
```

## Development

### Project Structure

```
.
├── cmd/                    # Service entry points
│   ├── engine/            # Engine service main
│   ├── games/             # Games service main
│   ├── matchmaking/       # Matchmaking service main
│   ├── bot/               # Bot service main
│   ├── auth/              # Auth service main
│   ├── notifications/     # Notifications service main
│   ├── gateway/           # API Gateway main
│   └── mancala/           # CLI client main
├── internal/              # Internal packages
│   ├── engine/           # Engine business logic
│   ├── games/            # Games business logic, models, storage
│   ├── matchmaking/      # Matchmaking queue and server logic
│   ├── bot/              # Bot AI engine and server logic
│   ├── auth/             # Authentication and JWT handling
│   ├── notifications/    # Real-time event notifications
│   ├── gateway/          # HTTP gateway handlers and middleware
│   ├── mancala/          # CLI client implementation
│   └── events/           # Redis Streams event publishing
├── proto/                # Protocol buffer definitions
│   ├── engine/          # Engine service protos
│   ├── games/           # Games service protos
│   ├── matchmaking/     # Matchmaking service protos
│   ├── bot/             # Bot service protos
│   ├── auth/            # Auth service protos
│   └── notifications/   # Notifications service protos
├── k8s/                  # Kubernetes manifests
├── iac/                  # Infrastructure as Code
│   ├── terraform/       # OpenStack VM provisioning
│   └── ansible/         # Kubernetes setup and deployment
└── README.md
```

### Testing

**Unit Tests**:
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/games/...
go test ./internal/matchmaking/...
go test ./internal/bot/...
```

**Integration Tests** (require Docker for testcontainers):
```bash
# Run Redis integration tests
go test ./internal/games/ -run TestRedisStorage

# Run complete integration tests
go test ./internal/games/ -run TestIntegration
```

### Building Docker Images

```bash
# Engine service
docker build -f cmd/engine/Dockerfile -t mancala/engine:latest .

# Games service
docker build -f cmd/games/Dockerfile -t mancala/games:latest .

# Matchmaking service
docker build -f cmd/matchmaking/Dockerfile -t mancala/matchmaking:latest .

# Bot service
docker build -f cmd/bot/Dockerfile -t mancala/bot:latest .
```

## Kubernetes Deployment

### Manual Deployment

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Deploy services
kubectl apply -k k8s/

# Check status
kubectl get all -n mancala

# Port forward for local access
kubectl port-forward -n mancala svc/gateway 8080:8080  # API Gateway (recommended)
kubectl port-forward -n mancala svc/games 50052:50052
kubectl port-forward -n mancala svc/matchmaking 50054:50054
kubectl port-forward -n mancala svc/bot 50057:50057
```

### Configuration

Services are configured via environment variables:

**Games Service**:
- `REDIS_ADDR`: Redis connection string (default: "localhost:6379")
- `ENGINE_ADDR`: Engine service address (default: "localhost:50051")

**Matchmaking Service**:
- `REDIS_ADDR`: Redis connection string (default: "localhost:6379")
- `GAMES_ADDR`: Games service address (default: "localhost:50052")
- `BOT_ADDR`: Bot service address (default: "localhost:50057")

**Bot Service**:
- `GRPC_PORT`: Service port (default: "50057")
- `AUTH_ADDR`: Auth service address (default: "localhost:50055")
- `JWT_SECRET`: JWT secret for authentication

## Game Rules

Mancala is a traditional board game with the following rules implemented:

1. **Board**: 14 pits (6 per player + 2 stores), each starting with 4 seeds
2. **Turns**: Players alternate, starting with Player 1
3. **Moves**: Pick up seeds from own pit, distribute counterclockwise
4. **Capture**: Landing in empty own pit captures opposite pit
5. **Extra Turn**: Landing in own store grants another turn
6. **Winning**: Game ends when one side is empty, most seeds wins

## Troubleshooting

### Common Issues

**gRPC Connection Issues**:
```bash
# Check if services are running
kubectl get pods -n mancala

# Check service logs
kubectl logs -n mancala deployment/gateway
kubectl logs -n mancala deployment/games
kubectl logs -n mancala deployment/engine
kubectl logs -n mancala deployment/matchmaking
kubectl logs -n mancala deployment/bot
```

**Redis Connection Issues**:
```bash
# Test Redis connectivity
kubectl exec -n mancala deployment/games -- nc -z redis 6379
```

**IaC Deployment Issues**:
```bash
# Check OpenStack credentials
openstack server list

# Verify Terraform state
cd iac/terraform && terraform show

# Test Ansible connectivity
cd iac/ansible && ansible all -m ping
```
