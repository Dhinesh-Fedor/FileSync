# FileSync

A file synchronization utility written in Go that compares directories using metadata and SHA256 hashing, detects filesystem changes, and synchronizes destination directories with source directories. 

FileSync was built to explore filesystem traversal, concurrent metadata processing, change detection, synchronization workflows, and command-line application design.

## Overview

FileSync maintains synchronization pairs consisting of a source directory and a destination directory. The application can:
* Register synchronization pairs
* Scan directory structures
* Generate file metadata
* Detect added, modified, and deleted files
* Synchronize destination directories
* Generate synchronization reports

The synchronization process follows a clear pipeline:

```text
Directory Scan
      │
      ▼
Metadata Generation
      │
      ▼
SHA256 Hashing
      │
      ▼
Change Detection
      │
      ▼
Synchronization
      │
      ▼
Report Generation
```

## Features

### Pair Management
Manage reusable synchronization pairs. Pairs are stored locally under `configs/pairs.json`.

### Directory Scanner
Recursively traverses directory structures while collecting relative file paths, file counts, and directory counts. 
> **Note:** The scanner ignores symbolic links and non-regular files.

### Metadata Generation
For every discovered file, FileSync concurrently collects:
* Relative path
* Absolute path
* File size
* Last modification time
* SHA256 hash

Metadata generation is performed concurrently using worker pools, channels, goroutines, and context cancellation to allow large directory trees to be processed efficiently.

### SHA256-Based Change Detection
FileSync determines whether files have changed using file size and SHA256 hash comparisons. Files are considered modified when the size differs **OR** the SHA256 differs. This prevents false positives caused by timestamp-only comparisons.

### Comparison Engine
The comparison engine categorizes files into the following states:

| Type | Description |
| :--- | :--- |
| **Added** | Exists in source but not destination |
| **Modified** | Exists in both locations but content differs |
| **Deleted** | Exists in destination but not source |

### Synchronization Engine
The engine applies detected changes to the destination:
* **Added Files:** Copied from Source -> Destination.
* **Modified Files:** Overwritten from Source -> Destination.
* **Deleted Files:** Removed from the Destination.

### Report Generation
Each synchronization operation generates a report containing added/modified/deleted file counts, bytes copied, start/end timestamps, and lists of affected files. Reports are stored under `reports/pair-<id>.json`.

## Architecture

```text
CLI
 │
 ├── Pair Manager       (Creates, loads, deletes, and retrieves pair details)
 │
 ├── Scanner            (Directory traversal, file discovery, stats)
 │
 ├── Metadata Generator (Extraction, SHA256 hashing, concurrent processing)
 │
 ├── Comparator         (Detects added, modified, and deleted files)
 │
 ├── Synchronizer       (Executes copy, overwrite, and delete operations)
 │
 └── Report Generator   (Summaries and persistent report storage)
```

## Project Structure

```text
FileSync
│
├── cmd/
│   └── cli/
│
├── configs/
│
├── internal/
│   ├── comparator/
│   ├── metadataV1/
│   ├── metadataV2/
│   ├── models/
│   ├── pairmanager/
│   ├── report/
│   ├── scanFiles/
│   └── synchronizer/
│
├── reports/
│
├── utils/
│   └── hash/
│
├── main.go
├── go.mod
└── go.sum
```

## Commands

**Add Pair**
```bash
fs add <source> <destination>
# Example: fs add /tmp/src /tmp/dst
```

**List Pairs**
```bash
fs list
```

**Pair Details**
```bash
fs details <pair-id>
```

**Scan Directories**
```bash
fs scan <source> <destination>
```

**Compare Directories**
```bash
fs compare <source> <destination>
```

**View Changes**
```bash
fs changes <pair-id>
```

**Synchronize**
```bash
fs sync <pair-id>
```

**View Report**
```bash
fs report <pair-id>
```

**Delete Pair**
```bash
fs delete <pair-id>
```

## Example Workflow

```bash
# Set up directories and sample file
mkdir -p /tmp/src
mkdir -p /tmp/dst
echo hello > /tmp/src/a.txt

# Add pair and compare
fs add /tmp/src /tmp/dst
fs compare /tmp/src /tmp/dst

# Sync and view report
fs sync 1
fs report 1
```

**Expected Synchronization Result:**
```text
Source
└── a.txt

Destination
└── a.txt
```

## Testing

Run all tests:
```bash
go test ./...
```
*Current test coverage includes: Scanner, Metadata generation, SHA256 hashing, Comparator, Pair management, Synchronization engine, and Report generation.*

## License

This project is licensed under the MIT License.
