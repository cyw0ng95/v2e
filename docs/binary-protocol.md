# Binary Message Protocol Documentation

## Overview

The v2e broker now supports an optimized binary message protocol with a 128-byte fixed header, providing significant performance improvements over JSON encoding for message routing.

## Performance Improvements

Benchmark results comparing binary protocol (JSON encoding) vs traditional JSON:
- **Unmarshal: ~10x faster** (232 ns vs 2322 ns)
- **Round-trip: ~2.3x faster** (1949 ns vs 4413 ns)  
- **Marshal: ~3.8x faster** (180 ns vs 683 ns)

## Binary Header Layout

The binary header is exactly 128 bytes with the following fields:

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0-1 | 2 bytes | Magic | Protocol identifier (0x56 0x32 = 'V2') |
| 2 | 1 byte | Version | Protocol version (0x01) |
| 3 | 1 byte | Encoding | Payload encoding (0=JSON, 1=GOB, 2=PLAIN) |
| 4 | 1 byte | MsgType | Message type (0=Request, 1=Response, 2=Event, 3=Error) |
| 5-7 | 3 bytes | Reserved | Reserved for future use |
| 8-11 | 4 bytes | PayloadLen | Payload length (uint32, big-endian) |
| 12-43 | 32 bytes | MessageID | Message ID (null-terminated string) |
| 44-75 | 32 bytes | SourceID | Source process ID (null-terminated string) |
| 76-107 | 32 bytes | TargetID | Target process ID (null-terminated string) |
| 108-127 | 20 bytes | CorrelationID | Correlation ID for request-response matching |

## Metrics & Telemetry

The broker now tracks comprehensive message statistics including wire size, encoding distribution, and per-process metrics.

See full documentation in the repository.
