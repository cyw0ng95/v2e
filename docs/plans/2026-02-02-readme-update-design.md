# README Update Design

## Overview

Update the README.md to reflect the current implementation, emphasizing build-time configuration via vconfig and clarifying the transport layer and build workflows.

## Changes Summary

### 1. Configuration Section (Lines ~119-158)

Replace the current configuration section with a table-based overview of vconfig options.

**New structure:**
- vconfig TUI introduction
- Four category tables: Logging, Access Service, Transport, Optimizer
- Each table shows: Option, Type, Default, Description

### 2. Transport & Communication Section (Lines ~140-160)

Update to clarify the dual-mode transport with build-time configuration.

**New structure:**
- Transport Modes table (UDS default, FD fallback)
- Transport Architecture diagram
- Message Flow steps (numbered list)
- Message Types table
- Key Features list

### 3. Build & Quickstart Section (Lines ~181-305)

Emphasize using `build.sh` wrapper instead of direct `go build`.

**New structure:**
- Prerequisites with IMPORTANT note about build.sh
- Build Script Options table
- Common Workflows code block
- Development Mode subsection
- Containerized Development subsection

**Key additions:**
- Warning about not using direct `go build` or `go test`
- Emphasize `./build.sh -r` for full system testing

### 4. Performance Characteristics Section (Lines ~456-478)

Add optimizer configuration table and restructure benefits.

**New structure:**
- Optimizer Configuration table (parameters from vconfig)
- Performance Benefits table (feature → benefit format)
- Performance Monitoring list

## Implementation Notes

- Preserve all existing diagrams (Mermaid charts)
- Keep all existing content not explicitly modified
- Maintain markdown formatting consistency
- No changes to: Architecture diagrams, Component Breakdown, Job Session Management, Project Layout, Broker Interfaces sections

## Files Modified

- `README.md` - Update sections as outlined above
