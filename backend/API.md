# CV Review API Documentation

All endpoints are prefixed with `/api/v1`. The server listens on port `8080` by default (configurable via `API_PORT`).

---

## Health

### `GET /api/v1/health`

Returns a simple health-check response.

**Handler:** inline in `internal/router/router.go:24`

**Response 200:**
```json
{ "status": "ok" }
```

---

## Exams

Handlers defined in `internal/handler/exam.go`.

### `POST /api/v1/exams`

Generate a new exam session. Randomly selects questions from the database, shuffles the options, and returns the exam **without** correct-answer data (so the frontend can display questions without revealing answers).

**Handler:** `CreateExam` — `internal/handler/exam.go:36`

**Request body:**
```json
{
  "question_count": 20,
  "module_id": 1
}
```
- `question_count` (required) — must be one of: `10`, `20`, `30`, `40`, `50`
- `module_id` (optional) — filter questions by module/theme

**Response 201:** exam session object
**Response 400:** invalid `question_count`
**Response 503:** no questions available (database not seeded)

---

### `PATCH /api/v1/exams/:id/answers`

Save answer progress incrementally. Called by the frontend on every page flip to persist the user's current selections. **Idempotent** — repeated calls with the same data are safe.

**Handler:** `SaveAnswers` — `internal/handler/exam.go:74`

**Request body:**
```json
{
  "answers": [
    { "exam_question_id": 1, "selected_option_id": 3 },
    { "exam_question_id": 2, "selected_option_id": 7 }
  ]
}
```

**Response 200:**
```json
{ "status": "saved" }
```
**Response 404:** exam not found
**Response 409:** exam has already been submitted

---

### `POST /api/v1/exams/:id/submit`

Submit the exam for grading. Records the score, marks the exam as submitted, and returns a full snapshot with correct answers and analysis revealed.

**Handler:** `SubmitExam` — `internal/handler/exam.go:107`

**Response 200:** graded exam result (includes correct answers, analysis, and score)
**Response 404:** exam not found
**Response 409:** exam has already been submitted

---

## Notes

Handlers defined in `internal/handler/note.go`. Notes are keyed by **question group ID** (`group_id`).

### `GET /api/v1/notes`

List all user notes.

**Handler:** `ListNotes` — `internal/handler/note.go:88`

**Response 200:** array of note objects

---

### `GET /api/v1/notes/:group_id`

Get a single note for a specific question group.

**Handler:** `GetNote` — `internal/handler/note.go:21`

**Response 200:** note object
**Response 404:** note not found

---

### `PUT /api/v1/notes/:group_id`

Create or update (upsert) a note for a question group.

**Handler:** `SaveNote` — `internal/handler/note.go:43`

**Request body:**
```json
{
  "content": "My study notes for this topic..."
}
```
- `content` (required)

**Response 200:** created or updated note object

---

### `DELETE /api/v1/notes/:group_id`

Delete a note for a question group.

**Handler:** `DeleteNote` — `internal/handler/note.go:68`

**Response 200:**
```json
{ "status": "deleted" }
```
**Response 404:** note not found

---

## Modules (Themes)

Handlers defined in `internal/handler/catalog.go`.

### `GET /api/v1/modules`

List all modules/themes.

**Handler:** `ListModules` — `internal/handler/catalog.go:40`

**Response 200:** array of module objects

---

### `POST /api/v1/modules`

Create a new module/theme.

**Handler:** `CreateModule` — `internal/handler/catalog.go:50`

**Request body:**
```json
{
  "name": "Computer Vision Basics"
}
```
- `name` (required)

**Response 201:** created module object
**Response 409:** module with that name already exists

---

## Question Groups

Handlers defined in `internal/handler/catalog.go`.

### `GET /api/v1/question-groups`

List question groups, optionally filtered by module.

**Handler:** `ListQuestionGroups` — `internal/handler/catalog.go:73`

**Query parameters:**
- `module_id` (optional) — filter groups belonging to a specific module

**Response 200:** array of question group objects

---

### `POST /api/v1/question-groups`

Create a new question group.

**Handler:** `CreateQuestionGroup` — `internal/handler/catalog.go:89`

**Request body:**
```json
{
  "module_id": 1,
  "title": "Image Filtering",
  "topic": "Convolution",
  "difficulty": "medium"
}
```
- `module_id` (required)
- `title` (required)
- `topic` (required)
- `difficulty` (required) — one of: `easy`, `medium`, `hard`

**Response 201:** created question group object
**Response 404:** module not found

---

## Questions

Handlers defined in `internal/handler/catalog.go`.

### `GET /api/v1/questions`

List questions, optionally filtered by group or module.

**Handler:** `ListQuestions` — `internal/handler/catalog.go:112`

**Query parameters:**
- `group_id` (optional) — filter by question group
- `module_id` (optional) — filter by module

**Response 200:** array of question objects

---

### `POST /api/v1/questions`

Create a new question.

**Handler:** `CreateQuestion` — `internal/handler/catalog.go:133`

**Request body:**
```json
{
  "group_id": 1,
  "module_id": null,
  "content": "What is a convolution kernel?",
  "analysis": "A convolution kernel is a small matrix...",
  "options": [
    { "content": "A small matrix used for filtering", "is_correct": true },
    { "content": "A type of neural network layer", "is_correct": false }
  ]
}
```
- `content` (required)
- `analysis` (required)
- `options` (required, minimum 2)
- `group_id` or `module_id` — at least one is required

**Response 201:** created question object
**Response 400:** invalid options
**Response 404:** question group or module not found

---

### `DELETE /api/v1/questions/:id`

Delete a question. Cascades to remove related `exam_questions` and `user_answers` records.

**Handler:** `DeleteQuestion` — `internal/handler/catalog.go:167`

**Response 200:**
```json
{ "status": "deleted" }
```

---

## Summary

| # | Method | Path | Handler | File |
|---|--------|------|---------|------|
| 1 | `GET` | `/api/v1/health` | inline | `router.go` |
| 2 | `POST` | `/api/v1/exams` | `CreateExam` | `exam.go` |
| 3 | `PATCH` | `/api/v1/exams/:id/answers` | `SaveAnswers` | `exam.go` |
| 4 | `POST` | `/api/v1/exams/:id/submit` | `SubmitExam` | `exam.go` |
| 5 | `GET` | `/api/v1/notes` | `ListNotes` | `note.go` |
| 6 | `GET` | `/api/v1/notes/:group_id` | `GetNote` | `note.go` |
| 7 | `PUT` | `/api/v1/notes/:group_id` | `SaveNote` | `note.go` |
| 8 | `DELETE` | `/api/v1/notes/:group_id` | `DeleteNote` | `note.go` |
| 9 | `GET` | `/api/v1/modules` | `ListModules` | `catalog.go` |
| 10 | `POST` | `/api/v1/modules` | `CreateModule` | `catalog.go` |
| 11 | `GET` | `/api/v1/question-groups` | `ListQuestionGroups` | `catalog.go` |
| 12 | `POST` | `/api/v1/question-groups` | `CreateQuestionGroup` | `catalog.go` |
| 13 | `GET` | `/api/v1/questions` | `ListQuestions` | `catalog.go` |
| 14 | `POST` | `/api/v1/questions` | `CreateQuestion` | `catalog.go` |
| 15 | `DELETE` | `/api/v1/questions/:id` | `DeleteQuestion` | `catalog.go` |

### Architecture

```
cmd/server/main.go          — entry point, wires DB + router
internal/router/router.go   — Gin engine setup, CORS, all route registration
internal/handler/           — HTTP handlers (thin: bind → call service → respond)
  ├── handler.go            — package doc
  ├── exam.go               — CreateExam, SaveAnswers, SubmitExam
  ├── catalog.go            — Modules, Question Groups, Questions CRUD
  └── note.go               — Notes CRUD
internal/service/           — business logic layer
internal/repository/        — database access layer
internal/model/             — data models
internal/database/          — DB connection setup
```
