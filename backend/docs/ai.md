# AI Service — `internal/ai`

> **Source:** [`backend/internal/ai/ai_service.go`](../internal/ai/ai_service.go)

---

## Overview

The AI service integrates with a **locally hosted Ollama** instance to provide three clinical NLP capabilities:

1. **Department Triage** — Suggests the appropriate CHU department based on patient symptoms
2. **WhatsApp Message Generation** — Creates culturally appropriate notification messages in Moroccan Darija + French
3. **Symptom Summarization** — Condenses referral notes into a 3-line TL;DR

All inference runs **locally** (default model: `llama3`), ensuring no patient data leaves the network.

---

## Architecture

```
┌──────────────┐     HTTP      ┌──────────────┐
│  Handlers    │──────────────▶│  AIService   │
│  (Gin)       │               │              │
└──────────────┘               └──────┬───────┘
                                      │
                                      │ POST /api/generate
                                      │ stream: false
                                      ▼
                               ┌─────────────────┐
                               │      Ollama     │
                               │     (llama3)    │
                               │ localhost:11434 │
                               └─────────────────┘
```

---

## Struct

### `AIService`

```go
type AIService struct {
    client *http.Client  // HTTP client with 10-minute timeout
    url    string        // Ollama API base URL (default: http://localhost:11434)
    model  string        // Model name (default: llama3)
}
```

### `DepartmentSuggestion`

```go
type DepartmentSuggestion struct {
    Department string  `json:"department"`
    Urgency    string  `json:"urgency"`    // LOW, MEDIUM, HIGH, CRITICAL
    Confidence float64 `json:"confidence"` // 0.0 – 1.0
    Reasoning  string  `json:"reasoning"`  // French clinical reasoning
}
```

---

## Functions

### `NewAIService(url, model) *AIService`

Creates a new AI service instance.

```go
ai := ai.NewAIService(
    os.Getenv("OLLAMA_URL"),   // default: http://localhost:11434
    os.Getenv("OLLAMA_MODEL"), // default: llama3
)
```

### `SuggestDepartment(symptoms string, patientAge int, departments []string) (*DepartmentSuggestion, error)`

Asks the LLM to act as a clinical triage assistant. Returns a structured department suggestion.

**Prompt strategy:** System prompt constrains output to JSON, with explicit department list and urgency scale.

**Error handling:** If JSON extraction fails, returns a fallback suggestion with the raw response in `Reasoning`.

```go
suggestion, err := ai.SuggestDepartment(
    "Douleur thoracique irradiant bras gauche, sueurs froides",
    65,
    []string{"Cardiology", "Neurology", "General Surgery"},
)
// → {Department: "Cardiology", Urgency: "HIGH", Confidence: 0.92, Reasoning: "..."}
```

### `GenerateWhatsAppMessage(patientName, symptoms, deptName, appointmentDate string) (string, error)`

Generates a culturally appropriate notification message combining Moroccan Darija (Arabic script) and French.

**Used by:** `WhatsAppService` when scheduling/referring appointments.

### `SummarizeSymptoms(symptoms string) (string, error)`

Condenses clinical notes into a 3-line French summary:

- **Line 1:** Primary complaint
- **Line 2:** Key clinical findings
- **Line 3:** Urgency assessment

**Stored in:** `Referral.AISummary` field (run asynchronously on referral creation).

---

## Internal

### `generate(prompt string) (string, error)`

Core LLM call — sends a POST to Ollama's `/api/generate` endpoint with `stream: false`.

```go
body := map[string]interface{}{
    "model":  a.model,
    "prompt": prompt,
    "stream": false,
}
```

---

## Configuration

| Environment Variable | Default                  | Description           |
| -------------------- | ------------------------ | --------------------- |
| `OLLAMA_URL`         | `http://localhost:11434` | Ollama API base URL   |
| `OLLAMA_MODEL`       | `llama3`                 | Model to use for inference |

---

## Integration Points

| Consumer                  | Method Called              | Async? |
| ------------------------- | -------------------------- | ------ |
| `CreateReferral` handler  | `SummarizeSymptoms`        | Yes (goroutine) |
| `SuggestDepartment` handler | `SuggestDepartment`      | No (synchronous) |
| `ScheduleReferral` handler | `GenerateWhatsAppMessage` | No (called by WhatsApp service) |
