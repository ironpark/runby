---
title: Gitea Actions
slug: gitea-actions
research_date: 2026-09-02
open_source: true
repository: https://github.com/go-gitea/gitea
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Gitea 공식 Actions variables 문서가 GITEA_ACTIONS 전용 마커와 workflow 실행 시 주입되는 GitHub 호환 변수·이벤트명을 명시하므로 별도 실행 검증 없이 감지 규칙을 확정할 수 있음
---

# Gitea Actions

Gitea Actions는 workflow run의 모든 step 환경에 `GITEA_ACTIONS=true`와 `CI=true`를 자동으로 설정합니다. 동시에 GitHub Actions 호환 변수도 제공하므로 `GITEA_ACTIONS`를 `GITHUB_ACTIONS`보다 먼저 검사해야 Gitea run을 GitHub Actions로 오인하지 않습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `GITEA_ACTIONS` | 불리언 문자열 (`true`) | 실행 식별 | Gitea Actions workflow에서 실행 중임을 표시 | 적합 — Gitea가 다른 CI와 구분하기 위해 자동 설정하는 전용 마커 | [Actions variables](https://docs.gitea.com/usage/actions/actions-variables/) |
| `GITHUB_RUN_ID` | 문자열 | 실행 식별 | 현재 workflow run ID | 적합 — Gitea가 제공하는 호환 run 식별자 | [Actions variables](https://docs.gitea.com/usage/actions/actions-variables/) |
| `GITHUB_JOB` | 문자열 | 실행 식별 | 현재 job ID | 적합 — Gitea가 제공하는 호환 job 식별자 | [Actions variables](https://docs.gitea.com/usage/actions/actions-variables/) |
| `GITHUB_EVENT_NAME` | 이벤트명 (`push`, `pull_request` 등) | 상태·컨텍스트 | run을 시작한 이벤트 | 보조 신호 — `pull_request`일 때 PR 실행 판정에 사용 | [Actions variables](https://docs.gitea.com/usage/actions/actions-variables/) |

`GITEA_ACTIONS`의 값만으로 실행 주체를 판정하고, `GITHUB_*` 값은 실행 위치와 이벤트를 채우는 컨텍스트로만 읽습니다. `GITHUB_EVENT_NAME=pull_request`이면 `PullRequest=true`로 보고하지만 Gitea가 별도 PR 번호 환경변수를 보장하지 않으므로 `PullRequestID`는 비워 둡니다. 판정 변수 이름만 `Evidence`에 남깁니다.

## 공식 문서

- [Actions variables](https://docs.gitea.com/usage/actions/actions-variables/)
- [Gitea Actions quickstart](https://docs.gitea.com/usage/actions/quickstart/)
- [Gitea 공식 저장소](https://github.com/go-gitea/gitea)
