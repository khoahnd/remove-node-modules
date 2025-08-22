# Node Modules Cleaner

A cross-platform Go tool to find and delete all `node_modules` directories in a given path with concurrent processing and platform-specific notifications.

## Features

- 🚀 **Multi-threaded scanning and deletion**
- 🧪 **Dry-run mode** to preview what would be deleted
- 📊 **Detailed statistics** and logging
- 🎯 **Smart directory skipping** (system dirs, version control, etc.)
- 💬 **Cross-platform notifications**:
  - **Windows**: Native message box
  - **macOS**: System notification + dialog box
  - **Linux**: Console output

## Usage

```bash
# Show help
go run main.go --help

# Scan and delete (with confirmation)
go run main.go -path /path/to/your/projects

# Dry run (preview only)
go run main.go -path /path/to/your/projects -dry-run

# Custom number of workers
go run main.go -path /path/to/your/projects -workers 8
```

## Platform-specific Notifications

### Windows
- Displays a native Windows message box with OK button
- Shows completion status and directs user to console for details

### macOS
- Shows a system notification with sound
- Displays a dialog box for user interaction
- Requires macOS's built-in `osascript` command

### Linux
- Prints formatted completion message to console
- Works on all Linux distributions without additional dependencies

## Command Line Options

- `-path`: Root directory path to scan (default: current directory)
- `-workers`: Number of worker threads (default: number of CPU cores)
- `-dry-run`: Only show what would be deleted, don't actually delete
- `-help`: Show help message

## Safety Features

- **Confirmation required**: Asks for 'yes' confirmation before actual deletion
- **System directory protection**: Skips system, version control, and hidden directories
- **Error handling**: Continues operation even if some directories can't be accessed
- **Detailed logging**: Shows progress and any errors encountered

## Examples

### Dry Run
```bash
go run main.go -path ~/Documents/projects -dry-run
```
Output includes what would be deleted without actual deletion.

### Production Run
```bash
go run main.go -path ~/Documents/projects
```
Requires typing 'yes' to confirm deletion.

### High Performance
```bash
go run main.go -path ~/Documents/projects -workers 16
```
Uses 16 worker threads for faster processing.

## Building

```bash
# Build executable
go build -o node-cleaner main.go

# Run executable
./node-cleaner -path /your/path
```

## Statistics Reported

- Total directories scanned
- Total node_modules directories found
- Total node_modules directories deleted
- Total errors encountered
- Execution time

## Requirements

- Go 1.16 or later
- **macOS**: Built-in `osascript` (available on all macOS systems)
- **Windows**: No additional requirements
- **Linux**: No additional requirements
