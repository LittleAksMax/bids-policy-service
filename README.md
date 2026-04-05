# Policy Service

## Module Summary

- `cmd/policy-service`: service entrypoint and startup wiring.
- `internal/api`: HTTP routing, controllers, request DTOs, and middleware wiring.
- `internal/service`: application logic for policy operations and format conversion.
- `internal/repository`: MongoDB persistence and policy models.
- `internal/cache`: Redis-backed request caching.
- `internal/config`: environment-based configuration loading.
- `internal/db`: MongoDB connection setup and health checks.
- `internal/health`: shared health check interface.
- `internal/convert`: tree/script conversion structures and validation helpers.
- `internal/validation`: request validation for policies and conversion payloads.
