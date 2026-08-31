---
title: systemd
slug: systemd
research_date: 2026-08-31
open_source: true
repository: https://github.com/systemd/systemd
product_type: task_runner
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 man 페이지에서 INVOCATION_ID가 유닛의 모든 프로세스에 전달된다는 계약은 확인했지만, (1) 시스템 인스턴스와 사용자 인스턴스(`systemd --user`)가 동일하게 동작하는지, (2) 유닛이 실행한 손자 프로세스까지 상속되는지, (3) `systemd-run`으로 띄운 임시 유닛에도 붙는지는 실제 Linux에서 관측하지 않음 (이 조사는 macOS에서 수행)
---

# systemd

systemd 유닛으로 실행되는 프로세스에 주입되는 변수입니다. 이 축에서 유일하게 "스크립트"가 아니라 **서비스 관리자**가 실행 주체인 경우입니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 주체 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `INVOCATION_ID` | 32자 16진 문자열 (128비트) | 실행 식별 | 유닛의 실행 사이클 식별 | **적합** — 유닛의 모든 프로세스에 같은 값이 전달됨 | [systemd.exec(5)](https://www.freedesktop.org/software/systemd/man/latest/systemd.exec.html) |
| `JOURNAL_STREAM` | `<device>:<inode>` | 상태·컨텍스트 | 표준 출력/오류가 journal에 연결됐는지 판별 | 보조 신호 — `StandardError=journal` 등 설정에 따라서만 존재 | 같은 문서 |

공식 문장입니다.

> `$INVOCATION_ID` — Contains a randomized, unique 128-bit ID identifying each runtime cycle of the unit, formatted as 32 character hexadecimal string. A new ID is assigned each time the unit changes from an inactive state into an activating or active state… **The same ID is passed to all processes run as part of the unit.**

> `$JOURNAL_STREAM` — If the standard output or standard error output of the executed processes are connected to the journal … `$JOURNAL_STREAM` contains the device and inode numbers of the connection file descriptor… **Note that it is generally not sufficient to only check whether `$JOURNAL_STREAM` is set at all** as services might invoke external processes replacing their standard output or standard error output.

두 번째 경고가 중요합니다. systemd 자신이 `JOURNAL_STREAM`의 **존재만으로 판단하지 말라**고 명시합니다 — 올바른 사용법은 값에 든 device/inode를 실제 파일 디스크립터의 것과 비교하는 것이고, 그것은 환경변수가 아니라 시스템 콜의 영역입니다. `runby`는 이 변수를 마커로 쓰지 않고 컨텍스트로만 노출하며, 이 경고를 이유로 기록합니다.

## 왜 이 축에서 가치가 큰가

"이 프로세스는 데몬으로 돌고 있는가"를 상속된 환경변수만으로 답할 수 있는 거의 유일한 경로입니다. `TTY`는 표준 스트림이 터미널이 아니라는 것까지만 말해 주는데, 그것은 파이프로 연결된 대화형 명령과 구분되지 않습니다.

WSL 문서([`../remote/wsl.md`](../remote/wsl.md))가 이미 지적한 것과도 맞물립니다 — 배포판 내부의 systemd 서비스에서는 `WSL_DISTRO_NAME`이 관측되지 않는 사례가 있습니다. 그런 프로세스에서 `INVOCATION_ID`가 보인다면 "로그인 셸 경로를 거치지 않았다"는 사실을 적극적으로 설명해 줍니다.

## 실행 파일 이름

`systemd`는 이름이 충분히 특정적이므로 `Executables`에 채웁니다. 다만 유닛의 프로세스는 대개 `systemd`의 직계 자식이므로 조상 체인에서 실제로 발견될 가능성이 높습니다.

## 결론

`INVOCATION_ID`의 존재가 판정 기준입니다. `JOURNAL_STREAM`은 공식 문서가 존재 여부만으로 판단하지 말라고 명시하므로 마커가 아니라 컨텍스트입니다.
