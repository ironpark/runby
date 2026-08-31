---
title: pre-commit
slug: pre-commit
research_date: 2026-08-31
open_source: true
repository: https://github.com/pre-commit/pre-commit
product_type: task_runner
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 문서에서 PRE_COMMIT=1 주입은 확인했지만, `pre-commit run` 수동 실행과 git 훅 경유 실행이 같은지, 그리고 언어별 격리 환경(python/node/docker)에서도 값이 전달되는지는 실제 실행으로 확인하지 않음
---

# pre-commit

[pre-commit](https://pre-commit.com/)은 git 훅을 관리하는 프레임워크입니다. `runby`가 이 축에서 **훅 실행을 감지할 수 있는 유일한 경로**입니다.

## git 훅 일반은 감지할 수 없습니다

[`README.md`](README.md)의 실측에서 확인했듯, git 훅과 git이 실행한 다른 자식 프로세스는 환경변수로 구분되지 않습니다. `post-checkout` 훅과 git 별칭은 주입되는 `GIT_*` 집합이 동일합니다.

pre-commit은 이 문제를 자기 마커로 우회합니다 — 프레임워크가 직접 변수를 심으므로, git이 알려 주지 않는 사실을 프레임워크가 알려 주는 구조입니다. 따라서 `runby`가 보고하는 것은 **"git 훅 안"이 아니라 "pre-commit이 실행한 훅 안"**이며, 이 구분을 흐리면 안 됩니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 주체 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `PRE_COMMIT` | `1` | 실행 식별 | 훅 실행 중임을 알림 | **적합** — 프레임워크가 훅 실행 동안 주입 | [pre-commit 공식 문서](https://pre-commit.com/) |
| `SKIP` | 쉼표 구분 훅 id 목록 | 설정 | 특정 훅 건너뛰기 | 부적합 — 사용자가 넣는 **입력 설정**이며 이름이 지나치게 일반적 | 같은 문서 |
| `PRE_COMMIT_USE_MAMBA` 등 | `1` | 설정 | conda 언어 설치 방식 선택 | 부적합 — 사용자 설정 | 같은 문서 |

공식 문서의 문장입니다.

> **new in 2.5.0**: pre-commit sets the `PRE_COMMIT=1` environment variable during hook execution.

버전 2.5.0에서 추가되었으므로 **그 이전 버전에서는 부재가 정상**입니다. 부재를 "pre-commit이 아님"으로 읽어서는 안 됩니다.

`SKIP`은 pre-commit이 **읽는** 변수이지 설정하는 변수가 아니며, 이름이 너무 일반적이라 무관한 도구와 충돌합니다. 이 패키지의 "설정 변수는 실행 근거가 아니다" 원칙에 따라 감지에도 컨텍스트에도 쓰지 않습니다.

## 실행 파일 이름

`Executables`는 **비워 둡니다.** pre-commit은 Python 진입점이라 조상 체인에서 `python`·`python3`으로 보일 수 있고, 이 이름들은 무관한 프로세스를 잘못 라벨링할 만큼 일반적입니다.

## 결론

`PRE_COMMIT=1`의 존재가 판정 기준입니다. 이는 "git 훅 안에서 실행 중"이 아니라 "pre-commit 프레임워크가 실행한 훅 안에서 실행 중"을 뜻하며, 2.5.0 미만에서는 나타나지 않습니다.
