# CLAUDE.md

## Development Rules

### Branching & Worktree

모든 기능 구현 시작 전에 반드시 아래 순서를 따른다:

1. `main` 브랜치에서 기능 브랜치 생성
2. 해당 브랜치로 git worktree 생성
3. worktree 내에서 구현 진행

```bash
# 예시
git worktree add ../provoke-feature-xxx feature/xxx
cd ../provoke-feature-xxx
```

구현 완료 후 PR을 통해 `main`에 머지한다.
