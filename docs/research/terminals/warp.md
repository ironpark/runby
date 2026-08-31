---
title: Warp
slug: warp
research_date: 2026-08-31
open_source: false
repository: null
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: Warp 클라이언트는 이슈 트래커만 공개된 비공개 소스 제품이라 공식 문서에 실행 시 주입되는 전체 환경변수 계약이 없습니다. TERM_PROGRAM_VERSION, WARP_IS_LOCAL_SHELL_SESSION, WARP_HONOR_PS1, WARP_USE_SSH_WRAPPER의 실존·정확한 값 형식과, TERM_PROGRAM=WarpTerminal이 SSH·tmux 하위 세션에서 실제로 어떻게 잔존·소실되는지는 Warp를 설치해 로컬/원격/tmux 각 경로에서 프로세스 환경을 직접 덤프해야 확인할 수 있습니다.
---

# Warp

Warp는 자사 환경변수에 대한 안정적 공개 계약을 문서화하지 않습니다. 공식 문서(`docs.warp.dev`)에는 `TERM_PROGRAM` 값을 이용해 Warp를 감지하는 셸 스크립트 예시가 실려 있어 이 값만은 제작사가 직접 보증하는 신호지만, 그 밖의 `WARP_*` 변수는 공식 문서 어디에도 이름·의미·값 형식이 명시되어 있지 않습니다. Warp 클라이언트 저장소(`warpdotdev/warp`)는 이슈·버그 리포트만 받는 "issues-only" 저장소이며, 제작사는 "Rust UI 프레임워크를 먼저, 이후 클라이언트 코드베이스의 일부 또는 전체를 오픈소스화할 계획"이라고 밝힐 뿐 현재는 클라이언트도 서버도 비공개입니다. 따라서 `TERM_PROGRAM` 이외의 변수는 GitHub 이슈에서 사용자가 관찰·언급한 이름을 근거로만 다룰 수 있으며, 이 문서는 그 구분을 표에 명시합니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `TERM_PROGRAM` | 문자열 (`WarpTerminal`) | 실행 식별 | Warp가 생성한 셸을 표준 터미널 식별 변수로 표시 | 적합 — 공식 문서가 `[[ $TERM_PROGRAM != "WarpTerminal" ]]` 조건문을 Warp 감지 방법으로 직접 예시함. 다만 `TERM_PROGRAM`은 다른 터미널도 쓰는 범용 변수이며 상속·재수출로 값이 복제될 수 있음 | [Terminal prompt — Disabling unsupported prompts for Warp](https://docs.warp.dev/terminal/appearance/prompt/) |
| `WARP_IS_LOCAL_SHELL_SESSION` | 불리언 형태로 추정 (공식 미확정) | 실행 식별 | 이름상 로컬 셸 세션과 원격(SSH 등) 세션을 구분하려는 목적으로 추정 | 보조 신호 — **공식 문서에는 전혀 등장하지 않습니다.** Warp 공식 이슈 트래커의 기능 요청 글에서 사용자가 "기존 Warp 환경변수(`WARP_IS_LOCAL_SHELL_SESSION`, `WARP_HONOR_PS1` 등)"라고 언급했을 뿐, 제작사가 의미·값·설정 시점을 확인해 준 적이 없습니다. 이름이 시사하는 "로컬 GUI와 같은 머신에서 실행 중"이라는 판단을 이 변수로 내리는 것은 검증되지 않은 추정이며, 아래 상속과 잔존 절의 이유로 신뢰하면 안 됩니다 | [GitHub Issue #8611 — 비공식 사용자 언급, 제작사 확인 없음](https://github.com/warpdotdev/warp/issues/8611) |
| `TERM_PROGRAM_VERSION` | 버전 문자열로 추정 (공식 미확정) | 상태·컨텍스트 | Warp 버전 컨텍스트 제공으로 추정 (다른 터미널의 관례를 따르는 것으로 보임) | 보조 신호 — 공식 문서에서 이 변수명이나 값 형식을 확인하지 못했습니다. 존재 자체가 비공식 언급에 의존합니다 | 공식 문서 없음 — 확인 불가 |
| `TERM` | 문자열 (예: `xterm-256color`류) | 설정 | 터미널 기능 유형을 셸에 알림 | 부적합 — Warp 고유 값이 아니며, 거의 모든 터미널이 설정하는 범용 변수 | 공식 문서에서 Warp 고유 값 명시 확인 안 됨 |

**가장 신뢰도 높은 단일 마커는 `TERM_PROGRAM=WarpTerminal`이며, 대소문자를 포함해 정확히 이 값입니다.** 이는 제작사가 자사 문서에서 직접 사용하는 유일한 감지 패턴입니다.

`WARP_IS_LOCAL_SHELL_SESSION`은 이름만 보면 "이 프로세스가 Warp GUI와 같은 머신에서 실행 중인가"와 "SSH로 도달한 원격 호스트에서 실행 중인가"를 구분해 줄 것처럼 보여 흥미롭지만, 조사 결과 이를 뒷받침하는 공식 근거는 없습니다. 이 변수가 실제로 존재하고 그런 의미를 가지더라도, 환경변수인 이상 상속·복제·위조에서 자유롭지 않으므로 "같은 머신에서 실행"이라는 구조적으로 어려운 판정을 안전하게 내려주지는 못합니다 (아래 상속과 잔존 절 참고).

또한 Warp는 세션·창·페인을 구분하는 공식 식별자 환경변수를 문서화하지 않습니다. 이런 식별자(`WARP_SESSION_ID`)는 외부 도구가 특정 탭/창에 딥링크할 수 있게 해달라는 기능 요청(GitHub Issue #8611)으로만 존재하며, 이 문서 작성 시점에는 구현·공개되지 않은 제안 단계입니다.

## 상태·컨텍스트 변수

Warp Drive의 "Environment Variables" 기능은 사용자가 직접 정의한 값을 팀과 공유하고 명령에 주입하는 기능이며, 이는 사용자 데이터이지 Warp가 실행 주체를 알리기 위해 주입하는 상태 신호가 아니므로 이 문서의 표에서 제외합니다. Warp 공식 문서는 `KUBECONFIG`처럼 표준 변수를 존중한다고 언급하지만 이 역시 Warp 고유 컨텍스트 신호가 아닙니다. 위 표에 실은 `TERM_PROGRAM_VERSION` 외에 제작사가 공식적으로 문서화한 상태·컨텍스트 변수는 확인되지 않았습니다 — Warp가 비공개 소스 제품이라 Zed의 Task 변수처럼 소스 코드에서 확인할 수 있는 보조 경로도 없습니다.

## 설정 변수 — 이름은 확인되나 환경변수 여부는 미확정

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `WARP_HONOR_PS1` | 불리언 형태로 추정 (공식 미확정) | 설정 | PS1 프롬프트 사용 여부를 제어하는 것으로 추정 | 부적합 — 공식 문서는 동일 기능을 **환경변수가 아니라** `honor_ps1` TOML 설정 키(`Settings > Appearance > Input > Shell (PS1)`)로만 문서화합니다. `WARP_HONOR_PS1`이라는 이름의 환경변수가 실제로 존재해 같은 역할을 하는지는 제작사가 확인해 준 적이 없습니다 | [All settings reference](https://docs.warp.dev/terminal/settings/all-settings/) · [Terminal prompt](https://docs.warp.dev/terminal/appearance/prompt/) · [GitHub Issue #8611 — 비공식 언급](https://github.com/warpdotdev/warp/issues/8611) |
| `WARP_USE_SSH_WRAPPER` | 불리언 형태로 추정 (공식 미확정) | 설정 | 레거시 SSH 래퍼 사용 여부를 제어하는 것으로 추정 | 부적합 — 공식 문서는 동일 기능을 `enable_ssh_warpification`/`enable_legacy_ssh_wrapper` TOML 키로 문서화하고, 우회 방법으로 `command ssh` 직접 호출을 안내할 뿐 이 이름의 환경변수는 등장하지 않습니다 | [Legacy SSH wrapper](https://docs.warp.dev/terminal/warpify/ssh-legacy/) · [All settings reference](https://docs.warp.dev/terminal/settings/all-settings/) |

두 변수 모두 실행 주체 감지에는 애초에 부적합한 **설정값**이라는 점에서, 존재가 확인되더라도 이 문서의 결론을 바꾸지 않습니다.

## 상속과 잔존

환경변수는 모든 자식 프로세스에 상속되므로 터미널 식별은 구조적으로 불안정합니다. Warp도 예외가 아니며, 오히려 공식 이슈 트래커에 이 불안정성을 직접 보여주는 사례가 있습니다.

- **SSH** — 일반 SSH는 클라이언트의 환경변수를 기본적으로 원격 호스트에 전달하지 않습니다(서버 `AcceptEnv`/클라이언트 `SendEnv` 설정이 없는 한). Warp 공식 이슈 트래커에는 사용자가 SSH로 원격 호스트에 접속했을 때 `TERM_PROGRAM`이 아예 존재하지 않았다고 보고한 사례가 있으며, 이는 표준 SSH 동작과 일치합니다. 즉 **Warp의 SSH 세션에서 원격 호스트의 프로세스가 `TERM_PROGRAM=WarpTerminal`을 그대로 물려받는다고 가정할 수 없습니다.** Warp는 대신 SSH 확장(companion server)이나 레거시 SSH 래퍼로 원격 셸을 별도로 부트스트랩해 Blocks·Input Editor 등 기능을 재현하지만, 공식 문서는 이 부트스트랩 과정에서 어떤 환경변수를 원격 셸에 주입하는지 명시하지 않습니다. 반대 방향의 위험도 있습니다 — SSH 클라이언트 설정에 따라 `TERM_PROGRAM`이 `SendEnv`로 전달되도록 구성된 환경이라면, 원격 호스트의 프로세스가 실제로는 Warp가 만든 것이 아닌데도 이 값을 그대로 물려받아 오탐을 일으킬 수 있습니다.
- **tmux / screen** — tmux 서버는 최초로 서버를 띄운 클라이언트의 환경을 그대로 유지하며, 이후 다른 터미널에서 붙은(attach) 클라이언트가 있어도 세션 안의 프로세스가 보는 `TERM_PROGRAM`은 서버를 처음 만든 그 터미널의 값으로 고정될 수 있습니다. 즉 Warp에서 시작한 tmux 세션에 나중에 다른 터미널(또는 다른 머신)에서 접속하면, 그 프로세스는 실제로 지금 보여주는 터미널이 아닌 Warp를 가리키는 `TERM_PROGRAM=WarpTerminal`을 계속 들고 있을 수 있습니다. tmux의 `update-environment` 설정으로 일부 변수를 attach 시점에 갱신할 수 있지만 기본 목록에 `TERM_PROGRAM`이 포함되는지는 tmux 설정에 달려 있고, Warp가 이를 위해 별도로 조정한다는 공식 언급도 없습니다. Warp 공식 문서는 레거시 tmux 기반 Warpify 기능이 있었지만 **지금은 폐기(deprecated)**되어 SSH 확장으로 대체되었다고 밝히고 있어, tmux 특유의 이 문제를 Warp가 특별히 보완하고 있다고 볼 근거는 없습니다.
- **분리·데몬화된 프로세스** — `nohup`, `disown`, 데몬화 등으로 터미널 생존 여부와 무관하게 살아남는 프로세스는 시작 시점에 상속한 `TERM_PROGRAM=WarpTerminal`을 터미널 종료 후에도 계속 들고 있을 수 있습니다. 이 값은 "Warp가 지금 이 프로세스를 소유·표시하고 있다"가 아니라 "과거 어느 시점에 Warp 셸에서 시작되었다"만을 의미합니다.
- **위조 가능성** — `TERM_PROGRAM`, `WARP_IS_LOCAL_SHELL_SESSION`을 포함한 모든 환경변수는 셸에서 `export`로 누구나 임의의 값으로 설정할 수 있는 평범한 문자열입니다. 검증되지 않은 이름(`WARP_IS_LOCAL_SHELL_SESSION` 등)은 애초에 공식 계약이 없으므로 위조 여부를 논하기 이전에 신뢰할 근거 자체가 약합니다.

결론적으로, 라이브러리가 정직하게 주장할 수 있는 것은 "이 프로세스의 환경에 `TERM_PROGRAM=WarpTerminal`이 존재한다"는 관찰뿐이며, 이는 "지금 이 프로세스가 Warp 터미널 창에 실시간으로 연결되어 표시되고 있다"거나 "Warp GUI와 같은 머신에서 실행 중이다"를 보장하지 않습니다. SSH·tmux를 거치면 이 값이 사라지거나(SSH 기본 동작), 다른 터미널로 이미 넘어간 세션에 잔존하거나(tmux), 애초에 실제 발생원과 무관하게 설정될 수 있습니다(위조).

## Warp의 에이전트 기능

Warp는 내장 AI Agent Mode("Oz")와 함께 Claude Code, Codex, Gemini CLI 같은 외부 CLI 코딩 에이전트를 자사 터미널 안에서 실행하는 기능을 마케팅합니다. 그러나 이번 조사에서 **Warp의 Agent Mode가 실행한 명령과 사용자가 터미널에 직접 입력한 명령을 구분하는 환경변수는 공식 문서 어디에서도 확인되지 않았습니다.** Terminal/Agent 모드 전환, Agent Mode Context, Warp Agent CLI 관련 공식 문서를 확인했지만 실행 주체를 표시하는 변수에 대한 언급은 없었습니다.

Warp Agent CLI(별도 CLI 제품) 문서에서는 `WARP_API_KEY`(CI 환경 인증), `WARP_TUI_DISABLE_AUTOUPDATE`(단일 실행에서 업데이트 확인 끄기), `WARP_NATURAL_LANGUAGE_DETECTION` 같은 변수가 등장하지만, 이들은 Warp Agent CLI 자체의 동작을 설정하는 **입력값**이지 그 CLI가 실행한 하위 셸 명령에 "이 명령은 에이전트가 요청했다"는 표식을 남기는 변수가 아닙니다.

이 결과는 Zed Agent 문서의 결론과 같은 패턴입니다 — `TERM_PROGRAM=WarpTerminal`은 **Warp가 만든 터미널**이라는 사실만 증명할 뿐, 그 안에서 실행된 특정 명령을 **사용자가 직접 쳤는지 Agent Mode가 요청했는지**는 구분하지 못합니다. 따라서 Warp는 `runby`에서 agent가 아니라 터미널 호스트로 취급해야 하며, `executes_agents`도 공식적으로 식별 가능한 별도 하네스 신호가 없어 빈 배열로 둡니다.

## 실행 주체 감지에 관한 결론

`TERM_PROGRAM=WarpTerminal`은 Warp가 공식 문서에서 직접 사용하는, 현재 프로세스가 Warp가 만든 셸에서 시작되었다는 가장 신뢰도 높은 공개 신호입니다. 그러나 이 신호는 다음 세 가지 이유로 제한적입니다.

1. **에이전트 여부를 구분하지 못함** — Warp Agent Mode가 실행한 명령과 사용자가 직접 입력한 명령 모두 같은 `TERM_PROGRAM` 값을 가집니다.
2. **상속·잔존에 취약함** — SSH로는 기본적으로 전달되지 않고, tmux를 거치면 실제 표시 터미널과 다른 값으로 잔존할 수 있으며, 분리된 프로세스에는 과거 값이 그대로 남습니다.
3. **`WARP_IS_LOCAL_SHELL_SESSION` 등 흥미로워 보이는 나머지 변수는 공식 확인이 없음** — 로컬/원격 구분처럼 유용해 보이는 판정을 이 변수에 맡기고 싶은 유혹이 있지만, 제작사가 공식적으로 문서화하거나 확인한 적이 없으므로 `runby`가 이를 근거로 확정적 판정을 내리는 것은 안전하지 않습니다.

따라서 라이브러리는 `TERM_PROGRAM=WarpTerminal`을 "Warp 터미널"이라는 `보조 신호` 수준의 호스트 식별에만 쓰고, 에이전트 판정이나 로컬/원격 판정의 근거로는 사용하지 않는 것이 안전합니다.

## 공식 문서

- [Terminal prompt — Disabling unsupported prompts for Warp](https://docs.warp.dev/terminal/appearance/prompt/)
- [SSH with Warp features](https://docs.warp.dev/terminal/warpify/ssh)
- [Legacy SSH wrapper](https://docs.warp.dev/terminal/warpify/ssh-legacy/)
- [Feature support over SSH](https://docs.warp.dev/code/ssh-feature-support/)
- [Warpify subshells](https://docs.warp.dev/terminal/warpify/subshells/)
- [Warpify overview](https://docs.warp.dev/terminal/warpify/)
- [All settings reference](https://docs.warp.dev/terminal/settings/all-settings/)
- [Sessions overview](https://docs.warp.dev/terminal/sessions/)
- [Terminal and Agent modes](https://docs.warp.dev/agents/local-agents/interacting-with-agents/terminal-and-agent-modes/)
- [Warp Agent CLI quickstart](https://docs.warp.dev/agents/cli/quickstart/)
- [Warp 공식 GitHub 저장소 (issues-only)](https://github.com/warpdotdev/warp)
- [GitHub Issue #8611 — Session ID Environment Variable and Deep Link Support (비공식 언급, 참고용)](https://github.com/warpdotdev/warp/issues/8611)
- [GitHub Issue #2070 — TERM_PROGRAM is not set on the remote host (SSH 상속 한계 근거)](https://github.com/warpdotdev/Warp/issues/2070)
