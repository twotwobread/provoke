# provoke

> **⚠️ This project is currently in draft. No features are functional yet.**

![Version](https://img.shields.io/badge/version-0.1.0--draft-orange)
![Status](https://img.shields.io/badge/status-WIP-red)
![License](https://img.shields.io/badge/license-MIT-blue)

[한국어](#한국어) | [English](#english)

---

## 한국어

> 자연어로 인프라를 호출하다.

> **⚠️ 현재 초안(draft) 단계입니다. 아직 동작하는 기능이 없습니다.**
> 설계 및 구현이 진행 중입니다. 진행 상황은 [로드맵](#로드맵)을 참고하세요.

**버전:** `0.1.0-draft`

`provoke`는 자연어 명령으로 클라우드 인프라를 관리하는 CLI 도구입니다. 내부적으로 Terraform을 생성하고 실행하며, 배포된 리소스를 추적하고 `.tf` 파일을 저장해 재현성과 git 관리를 지원합니다.

```bash
provoke "gcp에 gke 하나 구축해줘. 노드는 3개, 최대한 무료로."
```

```
생성할 리소스:
  + google_container_cluster.main  (us-central1, e2-micro x3)
  + google_container_node_pool.main

예상 비용: ~$0/월 (무료 티어)
적용할까요? [y/N] y

✓ 완료. .provoke/my-app/main.tf 저장됨.
```

### 주요 기능

- **자연어로 인프라 생성/수정/삭제** — Terraform 문법 몰라도 됨
- **상태 추적** — 이전에 만든 리소스를 기억하고 follow-up 명령 처리
- **`.tf` 파일 저장** — 생성된 Terraform 파일을 git으로 관리 가능
- **LLM provider 교체 가능** — OpenAI, Claude, 로컬 Ollama 모두 지원
- **프로젝트별 격리** — 디렉토리 기반으로 여러 프로젝트 독립 관리

### 요구사항

- [Terraform](https://developer.hashicorp.com/terraform/install) (>= 1.0)
- Cloud provider CLI (e.g. `gcloud`, `aws`)
- LLM API key (OpenAI / Anthropic) 또는 로컬 Ollama

### 설치

```bash
go install github.com/yourusername/provoke@latest
```

또는 [Releases](https://github.com/yourusername/provoke/releases)에서 바이너리 다운로드.

### 빠른 시작

```bash
# 1. 프로젝트 초기화
cd my-app
provoke init

# 2. 인프라 생성
provoke "gcp에 gke 클러스터 만들어줘. 노드 3개."

# 3. 수정
provoke "노드 2개로 줄여줘"

# 4. 현재 상태 확인
provoke status

# 5. 삭제
provoke "gke 클러스터 지워줘"
```

### 설정

`~/.provoke/config.yaml`:

```yaml
# Claude (Anthropic)
llm:
  provider: claude
  model: claude-sonnet-4-6
  api_key: YOUR_API_KEY

# OpenAI
llm:
  provider: openai
  model: gpt-4o
  api_key: YOUR_API_KEY

# Ollama (로컬)
llm:
  provider: ollama
  model: llama3.2
  base_url: http://localhost:11434
```

### 동작 방식

```
자연어 명령
    ↓
현재 state.json 로드 (배포된 리소스 요약)
    ↓
LLM에 컨텍스트 + 명령 전송 → .tf 파일 생성
    ↓
terraform plan 결과 요약 출력
    ↓
확인 후 terraform apply
    ↓
state.json 업데이트 + .tf 파일 저장
```

`.provoke/<project>/state.json`에 배포된 리소스 정보(타입, 이름, 파라미터, 생성 시각)를 저장해 follow-up 명령 시 컨텍스트로 활용합니다.

### 프로젝트 구조

```
my-app/
  .provoke/
    my-app/
      state.json      # 시맨틱 state (LLM 컨텍스트용)
      main.tf         # 생성된 Terraform 파일
      variables.tf
```

### 제한사항

- `provoke` 외부에서 생성된 기존 클라우드 리소스는 인식하지 못합니다 (v2에서 `provoke sync`로 지원 예정)
- Terraform이 로컬에 설치되어 있어야 합니다

### 로드맵

| 버전 | 상태 | 내용 |
|---|---|---|
| `0.1.0` | 🚧 진행 중 | 핵심 기능: 자연어 → terraform apply, state 추적, pluggable LLM |
| `0.2.0` | 📋 예정 | `provoke sync` — 기존 클라우드 리소스 임포트 |
| `0.3.0` | 💡 검토 중 | Recipe 시스템 — 커뮤니티 검증 템플릿 |

### 라이선스

MIT

---

## English

> Provision infrastructure by invoking it in plain language.

> **⚠️ This project is currently in draft. No features are functional yet.**
> Design and implementation are in progress. See the [Roadmap](#roadmap) for status.

**Version:** `0.1.0-draft`

`provoke` is a CLI tool that lets you manage cloud infrastructure using natural language. It generates and executes Terraform under the hood, tracks your deployed resources, and saves `.tf` files so everything stays reproducible and git-friendly.

```bash
provoke "set up a GKE cluster on GCP with 3 nodes, keep it as free as possible"
```

```
Resources to create:
  + google_container_cluster.main  (us-central1, e2-micro x3)
  + google_container_node_pool.main

Estimated cost: ~$0/month (free tier)
Apply? [y/N] y

✓ Done. Saved to .provoke/my-app/main.tf
```

### Features

- **Create, modify, and destroy infrastructure in plain language** — no Terraform syntax required
- **State tracking** — remembers what you've deployed and handles follow-up commands
- **`.tf` file persistence** — generated Terraform files are saved and git-trackable
- **Pluggable LLM provider** — works with OpenAI, Claude, or local Ollama
- **Per-project isolation** — directory-based project management, just like git

### Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) (>= 1.0)
- Cloud provider CLI (e.g. `gcloud`, `aws`)
- LLM API key (OpenAI / Anthropic) or local Ollama

### Installation

```bash
go install github.com/yourusername/provoke@latest
```

Or download a binary from [Releases](https://github.com/yourusername/provoke/releases).

### Quick Start

```bash
# 1. Initialize a project
cd my-app
provoke init

# 2. Create infrastructure
provoke "create a GKE cluster on GCP with 3 nodes"

# 3. Modify
provoke "scale down to 2 nodes"

# 4. Check current state
provoke status

# 5. Destroy
provoke "tear down the GKE cluster"
```

### Configuration

`~/.provoke/config.yaml`:

```yaml
# Claude (Anthropic)
llm:
  provider: claude
  model: claude-sonnet-4-6
  api_key: YOUR_API_KEY

# OpenAI
llm:
  provider: openai
  model: gpt-4o
  api_key: YOUR_API_KEY

# Ollama (local)
llm:
  provider: ollama
  model: llama3.2
  base_url: http://localhost:11434
```

### How It Works

```
Natural language command
    ↓
Load current state.json (summary of deployed resources)
    ↓
Send context + command to LLM → generate .tf file
    ↓
Show summarized terraform plan output
    ↓
Confirm → terraform apply
    ↓
Update state.json + save .tf files
```

`.provoke/<project>/state.json` stores deployed resource info (type, name, parameters, created_at) and is passed as context to the LLM on every command, enabling follow-up commands like "scale down the cluster I made yesterday."

### Project Structure

```
my-app/
  .provoke/
    my-app/
      state.json      # Semantic state (LLM context layer)
      main.tf         # Generated Terraform file
      variables.tf
```

### Limitations

- Resources created outside of `provoke` (e.g. via the cloud console) are not visible to the tool. (`provoke sync` planned for v0.2.0)
- Terraform must be installed locally.

### Roadmap

| Version | Status | Description |
|---|---|---|
| `0.1.0` | 🚧 In progress | Core: natural language → terraform apply, state tracking, pluggable LLM |
| `0.2.0` | 📋 Planned | `provoke sync` — import existing cloud resources into state |
| `0.3.0` | 💡 Considering | Recipe system — community-contributed templates for common patterns |

### License

MIT
