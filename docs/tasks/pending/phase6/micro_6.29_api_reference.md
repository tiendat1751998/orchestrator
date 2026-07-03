# Micro-Task 6.29: Create docs/api.md

## Info
- **File**: `docs/api.md`
- **Depends on**: 6.09-6.14 (gateway)
- **Time**: 15 min
- **Verify**: Visual review

## Purpose
REST API reference with endpoint documentation, request/response schemas, curl examples, and SSE streaming protocol.

## Key sections to include

### Base URL
```
http://localhost:8080/api/v1
```

### Response Envelope
All responses use consistent JSON envelope:
```json
{
  "data": { ... },
  "error": null,
  "meta": {
    "timestamp": "2026-01-01T00:00:00Z"
  }
}
```

### Endpoints

#### POST /missions — Create Mission
```bash
curl -X POST http://localhost:8080/api/v1/missions \
  -H "Content-Type: application/json" \
  -d '{"title":"Build REST API","description":"Create a user management API with Go"}'
```
Response: `202 Accepted`
```json
{"data":{"id":"mission_abc123","title":"Build REST API","status":"accepted"}}
```

#### GET /missions — List Missions
```bash
curl http://localhost:8080/api/v1/missions
```

#### GET /missions/:id — Get Mission
```bash
curl http://localhost:8080/api/v1/missions/mission_abc123
```

#### DELETE /missions/:id — Cancel Mission
```bash
curl -X DELETE http://localhost:8080/api/v1/missions/mission_abc123
```

#### GET /missions/:id/stream — SSE Stream
```bash
curl -N http://localhost:8080/api/v1/missions/mission_abc123/stream
```
SSE event format:
```
event: task_started
data: {"task_id":"t1","name":"Design schema","agent":"architect"}

event: task_completed
data: {"task_id":"t1","duration":"12.3s"}

event: heartbeat
data: {"t":"2026-01-01T00:00:15Z"}
```

#### GET /agents — List Agents
#### GET /providers — List Providers
#### GET /health — Health Check
```bash
curl http://localhost:8080/api/v1/health
```
Response: `200 OK`
```json
{"data":{"status":"healthy","timestamp":"2026-01-01T00:00:00Z"}}
```

### Error Responses
```json
{"error":{"code":400,"message":"title is required"}}
```

## Checklist
- [ ] All endpoints documented with curl examples
- [ ] Request/response schemas
- [ ] SSE event format documented
- [ ] Error response format
