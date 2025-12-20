# booking-service

<!--

├── cmd/
│   ├── api/
│   │   └── main.go              # HTTP API entry
│   └── worker/
│       └── main.go              # Reminder worker (cron-vari proses)
│
├── internal/
│   ├── config/                  # env, config struct
│   ├── logger/                  # logging setup
│   ├── http/                    # interface layer (REST API)
│   │   ├── router.go            # routes
│   │   └── handlers/            # HTTP handlers (ports)
│   │       ├── booking_handler.go
│   │       ├── user_handler.go
│   │       └── webhook_handler.go
│   │
│   ├── domain/                  # domain + application layers
│   │   ├── booking/
│   │   │   ├── entity.go        # Booking, TimeSlot, Availability
│   │   │   ├── service.go       # BookingService (use-cases)
│   │   │   └── ports.go         # interfaces: BookingRepo, SlotRepo, Notifier
│   │   ├── user/
│   │   │   ├── entity.go        # User, Business
│   │   │   ├── service.go       # UserService, AuthService
│   │   │   └── ports.go         # UserRepo, PasswordHasher, TokenManager
│   │   └── notification/
│   │       ├── entity.go        # NotificationLog, Template
│   │       ├── service.go       # NotificationService (enqueue, send)
│   │       └── ports.go         # NotificationRepo, WhatsAppClient, TelegramClient
│   │
│   ├── infrastructure/          # adapters (Postgres, Redis, HTTP clients)
│   │   ├── postgres/
│   │   │   ├── db.go            # sql.DB/pgx pool init
│   │   │   ├── booking_repo.go  # implements BookingRepo
│   │   │   ├── user_repo.go     # implements UserRepo
│   │   │   └── notification_repo.go
│   │   ├── redis/
│   │   │   └── cache.go         # optional: slot cache, rate limit
│   │   ├── whatsapp/
│   │   │   └── client.go        # WhatsApp Cloud API adapter
│   │   ├── telegram/
│   │   │   └── client.go        # Telegram Bot API adapter
│   │   ├── auth/
│   │   │   ├── jwt_manager.go   # JWT token generator/validator
│   │   │   └── password_hasher.go
│   │   └── queue/
│   │       └── db_queue.go      # DB-based job queue implementation
│   │
│   ├── worker/                  # background jobs (entry from cmd/worker)
│   │   └── reminder_worker.go   # uses domain.booking + domain.notification
│   │
│   └── shared/                  # common utilities/types
│       ├── errors.go
│       ├── timeutil.go
│       └── pagination.go
│
├── migrations/                  # SQL migration files (PostgreSQL)
│   ├── 001_init.up.sql
│   ├── 001_init.down.sql
│   └── ...
├── configs/                     # optional: config files
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum

 -->
