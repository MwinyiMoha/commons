# commons

Reusable packages for Go applications

## Installation

```bash
go get github.com/mwinyimoha/commons
```

## Packages

### Errors
Custom error handling with support for error wrapping and categorization. Uses a custom type that implements Go's `Error` interface. The package supports:

- Reporting field violations as extra information in validation errors
- Error wrapping with custom context and error codes
- Standard error types: `InvalidArgument`, `NotFound`, `Internal`, etc.
- Structured error responses for APIs. Includes Google-style error enrichment for gRPC APIs

### Logging
Structured logging utilities for consistent log output across applications. Uses [Zap](https://github.com/uber-go/zap) under the hood and supports:

- Contextual logging with key-value pairs
- Multiple log levels: Debug, Info, Warning, Error
- Integration with standard Go logging patterns

### Serializer
Type conversion and serialization utilities using reflection hooks. Uses [Mapstructure](https://github.com/mitchellh/mapstructure) under the hood and supports:

- Automatic conversion between struct types
- Built in hooks for common conversions e.g `uuid <-> string`
- Extensible hook system for custom field type conversions
- Batch serialization for slices