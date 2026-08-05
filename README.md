# Computer Vision · Exam Review

An interactive exam practice app for computer vision topics. Students can take randomly generated exams with auto-saved progress, submit for grading, and keep personal notes on each question group.

Built with **Go + Gin + GORM** (backend) and **React + TypeScript + Tailwind CSS** (frontend), orchestrated via **Docker Compose** with PostgreSQL.

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Or: Go 1.25+, Node.js 20+, PostgreSQL 16+

### Run with Docker

```bash
# Copy and configure environment
cp .env.example .env

# Start all services (PostgreSQL, backend, frontend)
docker compose up --build
```

The app will be available at **http://localhost:3000**.

- Frontend: `http://localhost:3000` (Nginx serving React build)
- Backend API: `http://localhost:8080/api/v1`
- PostgreSQL: `localhost:5432`

### Run locally (development)

```bash
# ── Backend ──────────────────────────────────────
cd backend
cp ../.env.example ../.env   # adjust DB_HOST=localhost if needed
go run ./cmd/server/main.go

# ── Frontend (separate terminal) ──────────────────
cd frontend
npm install
npm run dev                    # Vite dev server on :5173
```

## Usage

### 1. Start an exam

Open the app, select the number of questions (10 / 20 / 30 / 40 / 50), and click **Start Exam**. The backend randomly selects questions from the database, shuffles the option order per question (anti-cheat), and returns a response with **no correct-answer data**.

### 2. Take the exam

- Questions are paginated **10 per page**.
- Select an option for each question — your answers are **auto-saved**:
  - On every page flip (Previous / Next)
  - Every 30 seconds (interval timer)
  - On page close / refresh (best-effort via `fetch` keepalive)
- A countdown timer runs based on question count (2 min per question).
- Keyboard shortcuts: `←` `→` to navigate pages, `A` `B` `C` `D` to pick options.

### 3. Submit and review

When you reach the last page, click **Submit Exam**. A confirmation dialog warns about any unanswered questions. After submission:

- The exam is graded — you see your **score** (percentage).
- Each question shows whether you got it right/wrong, with the correct answer highlighted.
- Expand the **analysis** section for detailed explanations.
- Your answers are permanently recorded.

### 4. Take notes

Visit **/notes** (or click "View my notes" on the home page) to manage personal notes. Notes are keyed by **question group** (not individual question versions), so they survive content updates.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/health` | Health check |
| `POST` | `/api/v1/exams` | Create a new exam session |
| `PATCH` | `/api/v1/exams/:id/answers` | Save answers for an active exam |
| `POST` | `/api/v1/exams/:id/submit` | Submit exam for grading |
| `GET` | `/api/v1/notes` | List all user notes |
| `GET` | `/api/v1/notes/:group_id` | Get a note by question group |
| `PUT` | `/api/v1/notes/:group_id` | Create or update a note |
| `DELETE` | `/api/v1/notes/:group_id` | Delete a note |

### Anti-cheat design

During an active exam, the API **never** exposes:

- `Option.is_correct` — the frontend can't peek at answers
- `Question.analysis` — explanations are only revealed after grading

Option order is shuffled per question and persisted in `exam_questions.option_order`, so the same shuffled order is preserved across page saves and grading.

## Project Architecture

The project was built incrementally across 7 feature layers. Here's how each layer builds on the previous one:

### 1. Scaffold monorepo with Docker Compose

```
.
├── docker-compose.yml       # PostgreSQL + backend + frontend (nginx)
├── .env.example             # Database credentials, ports
├── backend/
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   └── cmd/server/main.go   # Entry point: env config, DB connect, router
└── frontend/
    ├── Dockerfile
    ├── nginx.conf            # Reverse proxy /api → backend
    ├── package.json
    └── index.html
```

**Key decisions:**
- Three services: `db` (PostgreSQL 16), `backend` (Go API), `frontend` (Nginx + React SPA)
- Nginx proxies `/api/*` requests to the backend container
- Database credentials flow through environment variables

### 2. Database models and auto-migration

```
backend/internal/
├── database/
│   ├── db.go                # GORM connection, AutoMigrate, connection pool
│   └── seed.go              # Idempotent seed from seed_data.json
├── model/
│   ├── model.go             # Entity relationship documentation
│   ├── question_group.go    # Logical question grouping (topic, difficulty)
│   ├── question.go          # Versioned questions (content, analysis)
│   ├── option.go            # Answer choices (content, is_correct, sort_order)
│   ├── exam.go              # Exam sessions (status, score, timestamps)
│   ├── exam_question.go     # Snapshot: pinned question version + shuffled order
│   ├── user_answer.go       # User's selected option per exam question
│   └── user_note.go         # Personal notes keyed by question group
└── data/
    └── seed_data.json       # 50+ English CV questions (groups → questions → options)
```

**Data model:**
```
question_groups  1──N  questions  1──N  options
       │                    │
       │                    │ (pinned version)
       │                    ▼
       │              exam_questions  N──1  exams
       │                    │
       │                    │ (1:1)
       │                    ▼
       │              user_answers
       │
       └──────────────── user_notes (1:1 by group_id)
```

**Key decisions:**
- **Versioned questions**: Updating a question inserts a new row with `version+1`. Exam snapshots reference the old version, so historical exam results stay accurate.
- **Auto-migration**: GORM AutoMigrate creates/updates tables on startup — no manual SQL.
- **Idempotent seeding**: `SeedIfEmpty()` only inserts if the `questions` table is empty, safe across restarts.

### 3. Exam generation with anti-cheat shuffle

```
backend/internal/
├── handler/exam.go           # POST /exams (CreateExam)
├── service/exam.go           # GenerateExam — random selection, shuffle, DTO stripping
├── repository/exam.go        # GetRandomQuestions, CreateExamWithQuestions (transaction)
└── pkg/shuffle/shuffle.go    # Fisher-Yates shuffle for option order
```

**Flow:**
1. Client sends `{ "question_count": 20 }`
2. Repository fetches N random questions (PostgreSQL `ORDER BY RANDOM()`)
3. Service shuffles option IDs per question, serializes the order to JSONB
4. Transaction creates `Exam` + `ExamQuestion` rows
5. Response DTO strips `is_correct` and `analysis` before serialization

**DTO layers:**
- `ExamResponse` / `QuestionResponse` / `OptionResponse` — exam mode (no answers)
- `ExamResultResponse` / `QuestionResultResponse` / `OptionResultResponse` — result mode (answers revealed)

### 4. Progress auto-save and exam submission

```
handler/exam.go  →  SaveAnswers (PATCH), SubmitExam (POST)
service/exam.go  →  SaveProgress (upsert), SubmitExam (grade + score)
repository/exam.go → UpsertUserAnswers, GetExamQuestionsForGrading, SubmitExam
```

**Auto-save:**
- `PATCH /api/v1/exams/:id/answers` — upserts `user_answers` rows (idempotent)
- GORM `OnConflict` clause handles insert-or-update in one query
- The frontend calls this on every page flip

**Submission & grading:**
1. Load all `exam_questions` with preloaded `Question.Options` and `UserAnswer`
2. Compare `selected_option_id` against `Option.is_correct` for each question
3. Calculate score: `correct_count / total_questions * 100` (rounded to 1 decimal)
4. Persist score and mark exam as `submitted`
5. Return full snapshot with `is_correct`, `analysis`, and `selected_option_id` revealed

**State machine:**
```
in_progress ──submit──▶ submitted
     │                      │
     └── save ──▶ (stays in_progress)
```

### 5. User notes CRUD API

```
handler/note.go   →  ListNotes, GetNote, SaveNote (PUT upsert), DeleteNote
service/note.go   →  DTO mapping, sentinel error translation
repository/note.go → UpsertNote (OnConflict), DeleteNoteByGroupID
```

**Key decisions:**
- Notes are keyed by `question_group_id`, not `question_id` — they survive content updates
- `PUT` is an upsert — the frontend doesn't care whether the note exists yet
- `GET /api/v1/notes/:group_id` returns `404` if no note exists (not `200` with null)

### 6. React frontend scaffold

```
frontend/src/
├── main.tsx                  # React entry point
├── App.tsx                   # React Router routes (/ → /exam/:id → /result/:id, /notes)
├── api/client.ts             # Typed fetch wrapper (ApiError, all API functions)
├── types/index.ts            # TypeScript interfaces mirroring backend DTOs
├── pages/
│   ├── HomePage.tsx          # Question count selector, exam creation
│   ├── ExamPage.tsx          # (stub) paginated exam with auto-save
│   ├── ResultPage.tsx        # Score display, per-question review
│   └── NotesPage.tsx         # Full CRUD with inline editing
├── components/
│   ├── QuestionCard.tsx      # Radio-button options with A/B/C/D labels
│   └── Pagination.tsx        # Previous / Next with page counter
└── index.css                 # Tailwind directives + base styles
```

**Key decisions:**
- **API client** (`client.ts`): A thin `fetch` wrapper with typed functions for every backend endpoint. Returns typed promises — no Redux or React Query (keeps dependencies minimal for a learning project).
- **React Router**: `react-router-dom` v7 with navigation state for passing exam data between pages.
- **Tailwind**: `brand` color palette (blue tones), academic/clean feel.
- **Type sharing**: TypeScript interfaces in `types/index.ts` mirror the backend's JSON DTOs — `Exam`, `ExamResult`, `Note`, `AnswerInput`, etc.

### 7. Exam page with pagination and auto-save

```
frontend/src/pages/ExamPage.tsx   # Full implementation
frontend/src/components/
├── QuestionCard.tsx              # (enhanced) touch-friendly option rows
└── Pagination.tsx                # (enhanced) 44px touch targets
```

**Features layered on the scaffold:**

| Feature | Implementation |
|---------|---------------|
| Pagination | 10 questions per page, `slice(startIdx, startIdx + PAGE_SIZE)` |
| Auto-save on page flip | `flushPage(currentPage)` before navigating |
| Interval auto-save | `setInterval(() => flushPage(currentPage), 30_000)` |
| Countdown timer | `2 min × question_count`, auto-submit at zero |
| Progress tracking | Progress bar = `answeredCount / totalQuestions × 100%` |
| Save status indicator | "Saving..." / "Unsaved changes" / "Saved at HH:MM" |
| Submit confirmation | Modal warning about unanswered questions |
| Keyboard shortcuts | `←` `→` for pages, `A` `B` `C` `D` for options |
| Before-unload save | `fetch` with `keepalive: true` on `beforeunload` |
| Mobile responsive | Compact header, 48px option rows, stacked modal buttons |

**Stale closure handling:** Timer and interval callbacks use `useRef` for `answers`, `currentPage`, `flushPage`, and `handleSubmit` — so the callbacks always read the latest state without re-registering.

## Project Structure

```
Computer-Vision-Go/
├── docker-compose.yml
├── .env.example
├── README.md
├── backend/
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   ├── cmd/server/main.go
│   ├── data/seed_data.json
│   └── internal/
│       ├── database/          # GORM connection, migration, seeding
│       ├── model/             # GORM entity structs
│       ├── repository/        # Data access (SQL via GORM)
│       ├── service/           # Business logic + DTOs
│       ├── handler/           # HTTP handlers (thin)
│       └── router/            # Gin route registration
└── frontend/
    ├── Dockerfile
    ├── nginx.conf
    ├── package.json
    ├── tailwind.config.js
    ├── index.html
    └── src/
        ├── App.tsx            # Routes
        ├── api/client.ts      # API client
        ├── types/index.ts     # TypeScript types
        ├── pages/             # Page components
        └── components/        # Reusable UI components
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend framework | Go 1.25 + Gin 1.12 |
| ORM | GORM 1.25 + PostgreSQL driver |
| Database | PostgreSQL 16 |
| Frontend framework | React 19 + TypeScript 5.7 |
| Build tool | Vite 6 |
| Styling | Tailwind CSS 3.4 |
| Routing | react-router-dom 7 |
| Containerization | Docker + Docker Compose |
| Reverse proxy | Nginx (production build) |
