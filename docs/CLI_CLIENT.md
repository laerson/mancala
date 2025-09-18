# Mancala CLI Client

A command-line client for playing Mancala games against other players online.

## Overview

The Mancala CLI client provides a terminal-based interface to connect to a Mancala game server, create accounts, join matchmaking queues, and play games in real-time with other players.

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/laerson/mancala.git
cd mancala

# Build the client
go build -o mancala cmd/mancala/main.go

# Or install directly
go install ./cmd/mancala
```

### Using Docker

```bash
# Build the Docker image
docker build -f cmd/mancala/Dockerfile -t mancala-client .

# Run the client
docker run -it mancala-client
```

## Quick Start

1. **Start the server** (see server documentation)
2. **Connect to server**: `mancala connect <server-ip>`
3. **Create account**: `mancala register`
4. **Join a game**: `mancala play`
5. **Make moves**: `mancala move <pit-number>`

## Commands Reference

### Connection Management

#### `mancala connect <server-ip>`
Connect to a Mancala game server.

```bash
# Connect to local server
mancala connect localhost

# Connect to remote server
mancala connect 192.168.1.100

# Connect with custom port
mancala connect 192.168.1.100:8080
```

**Notes:**
- Default port is 8080 if not specified
- Server connection is saved and persists between sessions
- Use `mancala status` to check connection

#### `mancala status`
Display current connection and login status.

```bash
mancala status
```

**Output:**
```
=== CONNECTION STATUS ===
Server: http://192.168.1.100:8080
Connected: ✓
Username: player1
Logged in: ✓
```

### Account Management

#### `mancala register`
Create a new user account.

```bash
mancala register
```

**Interactive prompts:**
- Username (3-30 characters)
- Password (minimum 8 characters)
- Confirm password

**Example:**
```
=== REGISTER NEW ACCOUNT ===
Username: player1
Password: ********
Confirm Password: ********
Creating account...
✅ Account created successfully!
Welcome, player1!
```

#### `mancala login`
Login to an existing account.

```bash
mancala login
```

**Interactive prompts:**
- Username
- Password (hidden input)

**Example:**
```
=== LOGIN ===
Username: player1
Password: ********
Logging in...
✅ Login successful!
Welcome back, player1!
```

#### `mancala logout`
Logout from the current account.

```bash
mancala logout
```

### Gameplay

#### `mancala play`
Join the matchmaking queue and wait for an opponent.

```bash
mancala play
```

**What happens:**
1. Joins matchmaking queue
2. Waits for another player
3. Receives match notification when paired
4. Displays real-time game updates
5. Shows ASCII game board after each move

**Example output:**
```
🎮 Joining matchmaking queue as player1...
✅ Player player1 successfully enqueued
⏳ Waiting for an opponent...
Press Ctrl+C to cancel and leave the queue.

🎯 MATCH FOUND!
==================
Player 1: player1 (user123)
Player 2: player2 (user456)

📝 Game ID: game789
Use 'mancala move <pit>' to make moves (in a new terminal)
```

**Controls:**
- **Ctrl+C**: Cancel queue and exit
- Keep this terminal open during the game to receive updates

#### `mancala move <pit-number>`
Make a move in the current game.

```bash
# Move stones from pit 3
mancala move 3
```

**Pit numbering (for Player 1):**
```
  0  1  2  3  4  5
```

**Example:**
```
🎲 Making move: pit 3...
✅ Move successful!
⏳ Waiting for opponent's move...
```

**Rules:**
- Pit numbers range from 0-5
- You can only move from your own pits
- Can only move when it's your turn
- Invalid moves will show an error message

## Game Interface

### Board Display

The game board is displayed in ASCII art format:

```
    MANCALA BOARD
  Player 2's side

  [  4 ][  4 ][  4 ][  4 ][  4 ][  4 ]
[  0 ]                              [  0 ]
  [  4 ][  4 ][  4 ][  4 ][  4 ][  4 ]

  Player 1's side

  Pit numbers (Player 1):
    0    1    2    3    4    5

  >>> Player 1's turn <<<
```

### Real-time Notifications

While `mancala play` is running, you'll receive:

1. **Match Found**: When paired with an opponent
2. **Move Made**: When opponent makes a move (shows updated board)
3. **Game Over**: When the game ends (shows final results)

### Multi-terminal Workflow

**Recommended setup:**
- **Terminal 1**: Run `mancala play` (keeps connection open for notifications)
- **Terminal 2**: Use `mancala move <pit>` to make moves

This allows you to see real-time updates while being able to make moves.

## Configuration

### Config File Location

The client stores configuration in:
- **Linux/macOS**: `~/.mancala/config.json`
- **Windows**: `%USERPROFILE%\.mancala\config.json`

### Config File Format

```json
{
  "server_url": "http://192.168.1.100:8080",
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "username": "player1",
  "user_id": "user123"
}
```

**What's stored:**
- Server connection details
- Authentication tokens
- User information

**Security:**
- Config file has restricted permissions (600)
- Contains sensitive authentication data
- Should not be shared or committed to version control

## Troubleshooting

### Connection Issues

**Problem**: Cannot connect to server
```bash
❌ Failed to connect to server: connection refused
```

**Solutions:**
1. Check if server is running
2. Verify server IP address and port
3. Check network connectivity
4. Ensure firewall allows connections

### Authentication Issues

**Problem**: Login failed
```bash
❌ Login failed: Invalid credentials
```

**Solutions:**
1. Double-check username and password
2. Ensure account exists (use `mancala register` if needed)
3. Check if server is accessible

### Game Issues

**Problem**: Cannot make moves
```bash
❌ No active game. Use 'mancala play' to join a game first.
```

**Solutions:**
1. Ensure you're in a game (run `mancala play` first)
2. Wait for match to be found
3. Check if it's your turn

**Problem**: Invalid pit number
```bash
❌ Invalid pit number: 6. Must be between 0-5.
```

**Solutions:**
1. Use pit numbers 0-5 only
2. Ensure the pit has stones to move
3. Check if it's your turn

### Configuration Issues

**Problem**: Settings not persisting
- Check file permissions on config directory
- Ensure adequate disk space
- Verify home directory is accessible

## Advanced Usage

### Scripting and Automation

The client can be used in scripts:

```bash
#!/bin/bash
# Auto-connect and register
echo "Connecting to server..."
mancala connect $SERVER_IP

echo "Creating account..."
# Note: Interactive commands need proper input handling
mancala register
```

### Multiple Accounts

To use multiple accounts, you can:
1. Use different user accounts on the system
2. Backup/restore config files
3. Use `mancala logout` to switch accounts

### Docker Usage

```bash
# Run client in container
docker run -it -v ~/.mancala:/root/.mancala mancala-client

# Connect to server
docker run -it mancala-client connect 192.168.1.100
```

## Game Rules (Quick Reference)

### Mancala Rules
1. **Objective**: Capture more stones than your opponent
2. **Setup**: 6 pits per player, 4 stones per pit initially
3. **Gameplay**:
   - Pick up all stones from one of your pits
   - Distribute one stone per pit, going counter-clockwise
   - Include your own mancala (store) but skip opponent's
4. **Scoring**: Stones in your mancala count as points
5. **Winning**: Player with most stones when game ends wins

### Turn Rules
- If last stone lands in your mancala, take another turn
- If last stone lands in empty pit on your side, capture opponent's stones
- Game ends when one side has no stones left

## Support

For issues, questions, or contributions:

1. Check this documentation first
2. Review server logs for connectivity issues
3. Verify server is running and accessible
4. Test with `mancala status` command

## Examples

### Complete Game Session

```bash
# 1. Connect to server
$ mancala connect 192.168.1.100
✅ Successfully connected to http://192.168.1.100:8080

# 2. Create account
$ mancala register
Username: alice
Password: ********
✅ Account created successfully!

# 3. Join game (Terminal 1)
$ mancala play
🎮 Joining matchmaking queue as alice...
⏳ Waiting for an opponent...

🎯 MATCH FOUND!
Player 1: alice (user123)
Player 2: bob (user456)

# 4. Make moves (Terminal 2)
$ mancala move 2
✅ Move successful!

# 5. Game continues with real-time updates...
📱 MOVE MADE
Player: bob
Pit: 1

# Updated board displays automatically
    MANCALA BOARD
  [  4 ][  4 ][  0 ][  5 ][  5 ][  5 ]
[  0 ]                              [  1 ]
  [  4 ][  4 ][  0 ][  5 ][  4 ][  4 ]

  >>> alice's turn <<<

# 6. Continue until game ends
🏁 GAME OVER!
Winner: alice
```

This documentation provides complete guidance for using the Mancala CLI client effectively.