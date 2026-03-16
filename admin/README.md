# Vehivle Admin

Admin backoffice scaffold based on:

- React + TypeScript + Vite
- Ant Design 5
- Axios + TanStack Query
- In-memory mock API adapter

## Features

- Login and token-based auth guard
- Dashboard summary
- Vehicle management
  - list/filter/pagination
  - create/edit
  - publish/unpublish
  - duplicate
  - batch status update
  - delete
- Category management (CRUD)
- Parameter template management (template + item list)
- System settings edit page
- Audit log list
- Unified API response format:

```json
{
  "code": "000000",
  "message": "success",
  "data": {},
  "request_id": "req_xxx",
  "timestamp": "2026-03-15T10:20:30+08:00"
}
```

## Run

```bash
cd admin
npm install
npm run dev
```

Default mock users:

- `admin / admin123`
- `operator / operator123`

## Notes

- All backend data is mocked in `src/mock/adapter.ts` and `src/mock/db.ts`.
- Mock data is in-memory only and resets after refresh/restart.
