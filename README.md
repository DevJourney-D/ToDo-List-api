# ToDo List Backend

Go backend API สำหรับ ToDo List application ที่เชื่อมต่อกับ Supabase

## Features

- 🔐 Authentication (Register/Login) with JWT
- 👤 User management and profile information
- 📊 User statistics (tasks, habits)
- 🗄️ PostgreSQL database integration with Supabase
- 🛡️ Secure password hashing with bcrypt
- 🚀 RESTful API with Gin framework

## API Endpoints

### Public Endpoints

- `POST /api/v1/register` - สมัครสมาชิก
- `POST /api/v1/login` - เข้าสู่ระบบ
- `GET /health` - Health check

### Protected Endpoints (ต้องมี JWT token)

- `GET /api/v1/user/info` - ดึงข้อมูลผู้ใช้และสถิติ

## Quick Start

1. Install dependencies:
```bash
cd backend
go mod tidy
```

2. Set up environment variables in `.env`:
```
DATABASE_URL=your_supabase_database_url
SUPABASE_URL=your_supabase_project_url
SUPABASE_ANON_KEY=your_supabase_anon_key
JWT_SECRET=your_jwt_secret
PORT=8080
```

3. Run the server:
```bash
go run main.go
```

## Authentication

Use Bearer token in Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

## Project Structure

```
backend/
├── config/          # Database configuration
├── handlers/        # HTTP handlers
├── middleware/      # Authentication middleware
├── models/          # Data models
├── utils/           # Utility functions
├── main.go          # Main application
├── go.mod           # Go modules
└── .env             # Environment variables
```

## Database Schema

ใช้ Supabase PostgreSQL database ตาม schema ที่กำหนด:
- users: ข้อมูลผู้ใช้
- tasks: งานที่ต้องทำ
- habits: นิสัยที่ต้องติดตาม
- logs: บันทึกการใช้งาน
