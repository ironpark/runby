---
title: kitty
slug: kitty
research_date: 2026-08-31
open_source: true
repository: https://github.com/kovidgoyal/kitty
product_type: terminal_emulator
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 Glossary가 각 KITTY_* 변수의 존재 조건과 값 형식을 명시하고, 공개 소스(child.py, tabs.py, shell_integration.py)에서 주입 지점을 직접 확인했으며 저자 본인이 GitHub Issue에서 KITTY_WINDOW_ID가 항상 정의됨을 확인해 별도 실행 검증 없이 표를 확정할 수 있음
---

# kitty

아래 내용은 조사 시점의 공식 소스 [`kovidgoyal/kitty`](https://github.com/kovidgoyal/kitty) 커밋 [`d5e912d`](https://github.com/kovidgoyal/kitty/commit/d5e912d999a3c1dce75f8da163785a9baaa40bf5)와 공식 문서 [Glossary](https://sw.kovidgoyal.net/kitty/glossary/)를 기준으로 확인했습니다. kitty는 새 창(window)을 만들 때마다 `KITTY_WINDOW_ID`, `KITTY_PID`, `KITTY_PUBLIC_KEY`, `KITTY_INSTALLATION_DIR`, `TERM` 등을 자식 프로세스의 환경에 주입한다고 소스에서 확인되며, 저자 kovidgoyal은 공식 이슈에서 `KITTY_WINDOW_ID`가 "항상 정의됨이 보장된다(guaranteed to be defined)"고 직접 밝혔습니다. 다만 이 보장은 kitty가 "만든" 프로세스 트리 안에서만 유효하며, 환경변수 자체의 상속 특성 때문에 그 트리를 벗어나면 신뢰도가 급격히 떨어집니다.

## 터미널 식별 신호

kitty의 두드러진 특징은 iTerm2·Zed·VS Code 등 다수의 터미널이 쓰는 `TERM_PROGRAM` 변수를 **설정하지 않는다**는 점입니다. 공식 Glossary 어디에도 `TERM_PROGRAM` 항목이 없고, 소스 코드에서도 해당 변수를 쓰는 지점을 찾을 수 없습니다. 저자 kovidgoyal은 "kitty를 어떻게 감지해야 하는가"를 묻는 [GitHub Issue #957](https://github.com/kovidgoyal/kitty/issues/957)에서 `TERM_PROGRAM` 대신 `KITTY_WINDOW_ID`가 "항상 정의됨이 보장된다"고 답했고, 더 강한 확인이 필요하면 이스케이프 시퀀스로 터미널에 직접 TN(terminfo name) capability를 질의해 SSH를 통해서도 `xterm-kitty` 응답을 받을 수 있다고 안내했습니다. 즉 kitty는 `TERM_PROGRAM` 관행에 편입하지 않기로 한 명시적 입장을 취하고 있으며, 공개 환경변수만으로는 `KITTY_WINDOW_ID`의 **존재 여부**가 가장 신뢰할 수 있는 정적 마커입니다.

`TERM=xterm-kitty`는 보조 신호로만 쓸 수 있습니다. `TERM`은 사용자가 셸 설정이나 `TERM=` 접두사로 얼마든지 덮어쓸 수 있고, 본래 터미널의 신원이 아니라 터미널이 흉내 내는 terminfo 능력 집합을 가리키는 값이기 때문입니다. SSH로 kitty 터미널이 없는 원격 호스트에 접속하면 `xterm-kitty` terminfo가 없어 오류가 나는 경우가 흔하다는 점([FAQ](https://sw.kovidgoyal.net/kitty/faq/))도 `TERM`이 신원이 아니라 기능 협상용 값임을 보여줍니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `KITTY_WINDOW_ID` | 정수 문자열 | 실행 식별 | 프로그램이 실행 중인 kitty 창(OS 창이 아니라 kitty 내부 tab/window)의 id. 원격 제어(remote control)에 사용 가능 | 적합 — 저자가 "항상 정의됨이 보장된다"고 명시한 가장 구체적인 존재 마커. 단, 값 자체는 다른 창을 가리키도록 상속·잔존될 수 있음(아래 상속과 잔존 참고) | [Glossary — `KITTY_WINDOW_ID`](https://sw.kovidgoyal.net/kitty/glossary/#envvar-KITTY_WINDOW_ID), [소스: `next_window_id()` 주입](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/kitty/tabs.py#L777), [저자 답변](https://github.com/kovidgoyal/kitty/issues/957) |
| `KITTY_PID` | 정수 문자열 (kitty 프로세스의 OS PID) | 실행 식별 | 현재 프로그램을 실행 중인 kitty 프로세스의 PID. 이 PID로 SIGUSR1을 보내 kitty 설정을 다시 읽게 할 수 있음 | 적합 — kitty 프로세스 자체의 PID이므로, 이 PID가 실제로 살아 있고 `kitty` 실행 파일인지 확인하는 **생존 검사(liveness check)**가 가능한 몇 안 되는 변수. 단, ssh kitten 문서는 tmux를 경유하면 값이 틀릴 수 있다고 명시 | [Glossary — `KITTY_PID`](https://sw.kovidgoyal.net/kitty/glossary/#envvar-KITTY_PID), [소스: `env['KITTY_PID'] = getpid()`](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/kitty/child.py#L384), [ssh kitten — tmux 경고](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/docs/kittens/ssh.rst#L158-L164) |
| `TERM` | 문자열, 기본값 `xterm-kitty` | 설정 | 터미널이 지원을 주장하는 terminfo capability 집합의 이름 | 부적합 단독 사용 시 — 사용자가 쉽게 덮어쓰고, 신원이 아니라 기능 협상용 값이며, 다른 프로그램(예: `tmux`)도 자신의 값으로 재작성함. `KITTY_WINDOW_ID`·`KITTY_PID`와 결합할 때만 보조 신호로 유효 | [Glossary — `TERM`](https://sw.kovidgoyal.net/kitty/glossary/#envvar-TERM), [소스: `env['TERM'] = opts.term`](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/kitty/child.py#L381) |

## 상태·컨텍스트 변수

`runby`가 실행 메타데이터로 안전하게 노출할 만한 값만 선별했습니다. `KITTY_PIPE_DATA`(스크롤백 내용 파이프)와 `KITTY_CHILD_CMDLINE`(벨 알림 콜백용 자식 커맨드라인)은 화면 내용·명령행 등 민감할 수 있는 페이로드를 담을 수 있어 제외했습니다. `KITTY_COMMON_OPTS`, `KITTY_SI_RUN_COMMAND_AT_STARTUP`은 kitten/셸 통합 내부 배관용이라 실행 주체 식별에 쓰이지 않아 제외했습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `KITTY_INSTALLATION_DIR` | 절대 디렉터리 경로 | 상태·컨텍스트 | kitty 설치 디렉터리 경로 | 보조 신호 — kitty가 만든 창임을 시사하지만 경로 문자열은 사용자가 동일하게 흉내 낼 수 있음 | [Glossary — `KITTY_INSTALLATION_DIR`](https://sw.kovidgoyal.net/kitty/glossary/#envvar-KITTY_INSTALLATION_DIR), [소스: 주입 지점](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/kitty/child.py#L404) |
| `KITTY_PUBLIC_KEY` | `protocol:key-data` 형식 문자열 | 상태·컨텍스트 | 원격 제어 프로토콜과 안전하게 통신하기 위한 공개키 | 보조 신호 — 원격 제어가 활성화된 kitty 창임을 시사하나, 원격 제어를 쓰지 않는 일반 창에는 값이 있어도 감지 목적으로는 존재 여부만 참고할 값 | [Glossary — `KITTY_PUBLIC_KEY`](https://sw.kovidgoyal.net/kitty/glossary/#envvar-KITTY_PUBLIC_KEY), [소스: `env['KITTY_PUBLIC_KEY']`](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/kitty/child.py#L385) |
| `KITTY_LISTEN_ON` | 소켓 경로 문자열 또는 `fd:<num>` | 상태·컨텍스트 | 원격 제어(remote control) 소켓 위치. 설정된 경우에만 존재 | 보조 신호 — 원격 제어가 명시적으로 켜진 창에서만 조건부로 존재하며, 일반 창 감지에는 쓸 수 없음 | [Glossary — `KITTY_LISTEN_ON`](https://sw.kovidgoyal.net/kitty/glossary/#envvar-KITTY_LISTEN_ON), [소스: 조건부 주입](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/kitty/child.py#L386-L391) |
| `KITTY_SHELL_INTEGRATION` | 공백으로 구분된 키워드 문자열 | 상태·컨텍스트 | 활성화된 셸 통합 기능 목록. 셸 통합 스크립트가 읽은 뒤 스스로 제거함 | 부적합 — 문서가 "셸 통합 스크립트가 자동으로 제거한다"고 명시해, `runby`가 관측할 시점에는 이미 사라진 경우가 대부분인 순간적 신호 | [Glossary — `KITTY_SHELL_INTEGRATION`](https://sw.kovidgoyal.net/kitty/glossary/#envvar-KITTY_SHELL_INTEGRATION), [소스: `env['KITTY_SHELL_INTEGRATION'] = ksi`](https://github.com/kovidgoyal/kitty/blob/d5e912d999a3c1dce75f8da163785a9baaa40bf5/kitty/shell_integration.py#L245) |

## 상속과 잔존

환경변수는 프로세스가 자식을 만들 때마다 그대로 복사되어 내려가므로, 어떤 후손 프로세스든 이 값들을 그대로 들고 있을 수 있습니다. 이는 kitty 식별을 **구조적으로 불안정**하게 만드는 근본 원인이며, `runby`가 이 문서의 어떤 신호도 "지금 이 순간 kitty가 이 프로세스를 소유한다"는 증명으로 격상시켜서는 안 되는 이유입니다.

- **SSH.** 일반 `ssh`는 `SendEnv`/`AcceptEnv`로 명시 허용된 변수(대개 `LANG`, `LC_*` 계열)만 전달하므로, `KITTY_*` 변수는 기본 설정에서 원격 호스트로 자동 전달되지 않습니다. 반면 kitty가 제공하는 [`kitten ssh`](https://sw.kovidgoyal.net/kitty/kittens/ssh/)는 이와 반대로, base64로 압축된 부트스트랩 스크립트를 TTY로 전송해 kitty의 terminfo와 셸 통합을 원격 호스트에 **의도적으로 복제**합니다. 즉 kitty는 "환경변수가 우연히 새지 않는다"는 SSH의 기본 원칙을 뒤집어, 원격 세션도 로컬 kitty 세션처럼 보이도록 상태를 능동적으로 전파하는 도구를 공식 제공합니다. ssh kitten 문서 자체가 "멀티플렉서를 경유하면 `KITTY_PID`와 `KITTY_WINDOW_ID` 값이 현재 세션에서 올바른지 여부에 따라 동작이 달라질 수 있고, tmux 세션에 다른 창에서 접속하면 값이 틀릴 수 있다"고 명시적으로 경고합니다.
- **tmux / screen.** tmux 서버는 그 서버를 최초로 띄운 클라이언트의 환경을 기준으로 세션 환경을 구성합니다. 따라서 kitty 창 A에서 시작한 tmux 세션을 나중에 kitty 창 B(또는 다른 터미널)에서 attach해도, 세션 안의 팬이 들고 있는 `KITTY_WINDOW_ID`/`KITTY_PID`는 여전히 창 A를 가리키는 **오래된(stale)** 값일 수 있습니다. tmux의 `update-environment` 설정으로 특정 변수를 attach 시점마다 갱신하게 만들 수 있지만, 이는 사용자가 명시적으로 설정해야 하며 기본으로 `KITTY_*` 변수를 포함하지도 않습니다. 이 문제는 kitty 공식 ssh kitten 문서가 직접 인정하는 한계이기도 합니다.
- **분리(detach)·데몬화된 프로세스.** `nohup`, `setsid`, 백그라운드 데몬 등으로 터미널에서 완전히 분리된 프로세스는 시작 당시 물려받은 `KITTY_WINDOW_ID`/`KITTY_PID`를 프로세스가 살아있는 한 계속 들고 있습니다. 이때 그 값이 가리키던 kitty 창이나 kitty 프로세스 자체는 이미 닫혔을 수 있습니다. 흥미롭게도 `KITTY_PID`는 이 정확한 상황을 스스로 검출할 방법을 제공하는 몇 안 되는 변수입니다 — 값을 스냅샷으로만 신뢰하는 대신, 실제로 그 PID의 프로세스가 살아 있고 `kitty` 실행 파일인지 조회하면 "당시엔 kitty가 있었지만 지금은 없다"는 상태를 구분할 수 있습니다. `KITTY_WINDOW_ID`에는 이런 생존 검사 수단이 없습니다.
- **위조 가능성.** 모든 환경변수는 사용자나 스크립트가 `env KITTY_WINDOW_ID=1 KITTY_PID=1 ...`처럼 임의로 설정할 수 있는 평범한 문자열입니다. kitty 소스 어디에도 서명이나 검증 메커니즘은 없으므로, 이 값들의 존재만으로는 위조와 정상 주입을 구분할 수 없습니다.

결론적으로 `runby`가 이 문서의 신호로 정직하게 주장할 수 있는 것은 "이 프로세스 트리의 어느 시점에 kitty가 환경을 주입했다"는 **역사적 사실**까지입니다. "지금 이 프로세스를 보고 있는 kitty 창이 실제로 존재하고 살아 있다"는 **현재 사실**은 `KITTY_PID`에 대해서만, 그것도 별도의 프로세스 존재 확인을 추가로 수행했을 때만 주장할 수 있습니다.

## 실행 주체 감지에 관한 결론

kitty는 `TERM_PROGRAM` 관행에 참여하지 않기로 한 명시적 입장을 취하므로, 공개 환경변수만으로 kitty를 식별하는 가장 신뢰도 높은 방법은 `KITTY_WINDOW_ID`의 **존재**를 확인하는 것입니다 — 저자 본인이 이 변수는 kitty가 만든 프로세스 트리 안에서 항상 정의됨을 보장한다고 밝혔습니다. `KITTY_PID`는 여기에 더해 그 값이 가리키는 프로세스가 지금도 살아 있는지 검사할 수 있다는 점에서 한 단계 더 강한 신호이며, `runby`가 "터미널이 지금도 존재하는가"를 판단해야 한다면 이 변수를 우선 활용해야 합니다. `TERM=xterm-kitty`는 `KITTY_*` 변수가 이미 존재를 확정한 뒤에만 보강용으로 쓸 가치가 있고, 단독으로는 사용자가 손쉽게 덮어쓸 수 있는 약한 신호입니다.

다만 SSH(특히 `kitten ssh`의 의도적 상태 전파), tmux/screen의 클라이언트별 서버 환경 상속, 그리고 detach된 장수 프로세스라는 세 가지 경로 모두에서 이 변수들이 실제 소유 터미널과 어긋날 수 있음이 공식 문서로 확인됩니다. 따라서 `runby`는 이 신호들을 "kitty" 식별의 **적합** 등급으로 보고하되, 그 판정이 프로세스가 태어난 시점의 환경에 대한 것이며 판정 시점에 그 kitty 창(또는 프로세스)이 여전히 살아 있음을 별도로 검증하지 않는 한 현재형 보장이 아니라는 점을 함께 알려야 합니다.

## 공식 문서

- [Glossary — 환경변수 전체 목록](https://sw.kovidgoyal.net/kitty/glossary/)
- [Truly convenient SSH (`kitten ssh`)](https://sw.kovidgoyal.net/kitty/kittens/ssh/)
- [Shell integration](https://sw.kovidgoyal.net/kitty/shell-integration/)
- [Frequently Asked Questions](https://sw.kovidgoyal.net/kitty/faq/)
- [GitHub Issue #957 — kitty 감지 방법에 대한 저자 답변](https://github.com/kovidgoyal/kitty/issues/957)
- [공식 소스 저장소 (`kovidgoyal/kitty`)](https://github.com/kovidgoyal/kitty)
