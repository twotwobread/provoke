# Provoke — Design Spec
**Date:** 2026-03-27

## Overview

`provoke` is a CLI tool that lets developers manage cloud infrastructure using plain natural language. It generates and executes Terraform under the hood, maintains a semantic state of deployed resources, and saves `.tf` files for reproducibility and git-based version control.

**Tagline:** provision + invoke — 명령으로 인프라를 호출한다.

---

## Target User

Developers who occasionally need cloud infrastructure but find Terraform syntax tedious to look up. Primarily those coming from on-premises environments who manage cloud resources via UI and want a faster, reproducible alternative.

---

## Core Values

1. **자연어 명령 한 줄로 인프라 생성/수정/삭제**
2. **`.tf` 파일 저장 — git으로 재현성 확보**
3. **LLM provider 교체 가능 (OpenAI / Claude / Ollama)**

---

## Architecture

### Components

| Component | Responsibility |
|---|---|
| CLI Layer | Natural language parsing, user confirmation prompts (cobra) |
| State Manager | Read/write `state.json`, derive from `tfstate` |
| Context Builder | Assemble LLM prompt from state + .tf files + current date |
| LLM Client | Pluggable provider interface |
| Terraform Runner | Execute plan/apply/destroy, self-healing on failure (max 2 retries) |

### Data Flow

```
provoke "노드 2개로 줄여줘"
        ↓
[State Manager] Load state.json
        ↓
[Context Builder] current state + main.tf + today's date → LLM prompt
        ↓
[LLM Client] Generate updated .tf file
        ↓
[CLI Layer] Show summarized terraform plan + "적용할까요? [y/N]"
        ↓
[Terraform Runner] apply
        ↓
[State Manager] Update state.json + save .tf files
```

---

## State Management

### Two-layer state

**1. `terraform.tfstate`** — Terraform's internal ground truth. Never modified directly.

**2. `state.json`** — Semantic layer for LLM context. Derived from tfstate after every successful apply.

```json
{
  "project": "my-app",
  "provider": "gcp",
  "resources": [
    {
      "type": "google_container_cluster",
      "name": "main-cluster",
      "params": {
        "node_count": 3,
        "machine_type": "e2-micro",
        "region": "us-central1"
      },
      "created_at": "2026-03-27T10:00:00Z"
    }
  ]
}
```

- `state.json` is regeneratable from `terraform show -json` at any time
- Includes `created_at` per resource to support temporal queries ("이틀 전에 만든 거 지워줘")
- **Assumption:** All infrastructure is managed through `provoke`. Existing resources created outside the CLI are not visible (v2: `provoke sync` for import)

---

## LLM Client

### Pluggable Interface

```
LLMClient (interface)
  ├── OpenAIClient
  ├── ClaudeClient
  └── OllamaClient
```

Configured in `~/.provoke/config.yaml`:

```yaml
# Claude example
llm:
  provider: claude
  model: claude-sonnet-4-6
  api_key: sk-...

# Ollama example
llm:
  provider: ollama
  model: llama3.2
  base_url: http://localhost:11434
```

### Prompt Structure

```
[System]
You are a Terraform expert.
Current cloud provider: {provider}
Current date: {ISO timestamp}
Current deployed state: {state.json summary}
Current .tf file: {main.tf contents}

[User]
{natural language command}
```

LLM returns the **full updated `.tf` file** (not a delta) to avoid context loss. The CLI replaces the existing file with the returned content.

---

## Terraform Runner

### Execution Flow

```
LLM returns .tf file
        ↓
terraform plan → summarized diff output
        ↓
"이렇게 변경됩니다. 적용할까요? [y/N]"
        ↓
y → terraform apply
N → cancel, rollback .tf file
        ↓
success → update state.json + save .tf
failure → self-healing: send error + .tf back to LLM → retry (max 2x)
          → if still failing: show error, rollback
```

### Plan Output

terraform plan output is summarized via LLM before displaying to the user:

```
변경 사항:
  ~ google_container_cluster.main
      node_count: 3 → 2

적용할까요? [y/N]
```

---

## Project Structure

Directory-based project isolation (like git):

```
my-app/
  .provoke/
    <project-name>/
      state.json
      main.tf
      variables.tf
```

Running `provoke` in a directory uses that directory's `.provoke/` folder.

---

## CLI Commands

```bash
provoke init                        # Initialize .provoke/<project-name>/ in current directory
provoke "<natural language>"        # Main command — create/modify/destroy infrastructure
provoke status                      # Show current deployed state from state.json
```

---

## Tech Stack

- **Language:** Go
- **CLI framework:** cobra
- **Terraform execution:** `os/exec` (terraform binary must be installed)

---

## Roadmap

| Version | Features |
|---|---|
| v1 | Core: natural language → terraform apply, state.json, pluggable LLM |
| v2 | `provoke sync` — import existing cloud resources into state |
| v3 (optional) | Recipe system — community-contributed templates for common patterns |
