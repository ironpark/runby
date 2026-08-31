---
title: Apple Terminal
slug: apple-terminal
research_date: 2026-08-31
open_source: false
repository: null
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: TERM_PROGRAM/TERM_SESSION_ID/TERM_PROGRAM_VERSION은 Apple의 공식 사용자 가이드에 이름조차 등장하지 않고 /etc/bashrc_Apple_Terminal 주석과 관찰 사례로만 확인되므로, 현재 macOS 버전에서 실제 값과 셸(bash/zsh)별 주입 여부를 직접 실행해 확인해야 함
---

# Apple Terminal

Apple은 macOS에 기본 내장된 Terminal.app이 어떤 환경변수를 설정하는지 공식적으로 계약하지 않습니다. Apple 공식 사용자 가이드([Use environment variables in Terminal on Mac](https://support.apple.com/guide/terminal/use-environment-variables-apd382cc5fa-4f58-4449-b20a-41c53c006f8f/mac))는 환경변수의 일반 개념과 `PATH` 예시만 다루며, `TERM_PROGRAM`, `TERM_SESSION_ID`, `TERM_PROGRAM_VERSION`이라는 이름은 이 문서 어디에도 등장하지 않습니다. 이 변수들의 존재는 macOS가 기본 셸 설정으로 배포하는 Apple 작성 스크립트 `/etc/bashrc_Apple_Terminal`의 주석과, 다수의 독립적 관찰(포렌식 분석, 커뮤니티 문서)로만 확인됩니다. 반면 `TERM`(Declare terminal as)과 `LANG`(Set locale environment variables on startup)은 Apple 공식 가이드의 Advanced 프로파일 설정 페이지에 이름과 동작이 명시되어 있습니다. Terminal.app 자체는 폐쇄 소스이며 별도의 공개 저장소가 없습니다.

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `TERM_PROGRAM` | 문자열 (`Apple_Terminal`) | 실행 식별 | 현재 셸을 실행한 터미널 호스트 프로그램 식별 | 보조 신호 — 관찰상 가장 널리 쓰이는 존재 마커이지만, Apple 공식 사용자 가이드에는 이름이 등장하지 않고 값 보증도 문서화되어 있지 않음. `/etc/bashrc_Apple_Terminal`(macOS 배포 스크립트)이 이 변수를 전제로 동작을 분기하는 것으로 존재를 간접 확인함 | [공식 소스: `/etc/bashrc_Apple_Terminal`](https://gist.github.com/floam/f535842a16226e77d014d67bade2b2f3) (macOS 기본 배포 파일, Apple 공식 사용자 가이드에는 미기재) |
| `TERM_SESSION_ID` | UUID 문자열 (예: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` 형태) | 실행 식별 | Terminal이 각 세션에 부여하는 고유 식별자 | 보조 신호 — 세션(주로 창) 단위로는 강한 신호이지만, Apple 공식 사용자 가이드에 문서화되지 않았고 정확한 부여 단위(창/탭/셸 재시작)에 대한 공식 규격이 없음 | [공식 소스: `/etc/bashrc_Apple_Terminal`](https://gist.github.com/floam/f535842a16226e77d014d67bade2b2f3) — 주석: "Terminal assigns each terminal session a unique identifier and communicates it via the TERM_SESSION_ID environment variable" |
| `TERM_PROGRAM_VERSION` | 버전/빌드 문자열 | 상태·컨텍스트 | Terminal.app 버전 정보 제공 | 보조 신호 — `TERM_PROGRAM=Apple_Terminal`과 함께일 때만 의미가 있으며, 값 형식과 보증 범위가 Apple 공식 문서에 전혀 기재되지 않음 | 공식 문서 없음 — 관찰 사례로만 확인 |
| `TERM` | 문자열 (기본값은 Advanced 프로파일의 "Declare terminal as" 팝업 메뉴 선택값을 따름) | 설정 | `terminfo`가 참조할 터미널 기능 유형 지정 | 부적합 — Apple 공식 가이드가 명시하는 사용자 설정값이며, 다른 터미널·SSH 세션에서도 동일 이름의 아주 일반적인 표준 변수임. 문서는 구체적 기본값 문자열을 명시하지 않음 | [Change profiles: Advanced preferences](https://support.apple.com/guide/terminal/change-profiles-advanced-preferences-trmladvn/mac) |

## 상태·컨텍스트 변수

`runby`가 실행 메타데이터로 노출할 만한 값만 선별했습니다. `LANG`은 Apple 공식 가이드가 명시적으로 다루는 유일한 로케일 변수이므로 포함했고, `COLORTERM`은 관찰 결과 Terminal.app이 기본적으로 설정하지 않는 값(24비트 트루컬러 미지원과 일치)이라 표에서 제외했습니다.

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `LANG` | 로케일 문자열 (예: `en_US.UTF-8`), 설정을 끄면 빈 문자열 | 설정 | 프로파일의 "Set locale environment variables on startup" 옵션이 켜져 있으면 사용자 로케일로 설정 | 부적합 — Apple 공식 가이드가 문서화하는 설정값이지만 다른 모든 유닉스 계열 프로그램이 공유하는 표준 변수이며, 값이 없거나 사용자가 임의로 바꿀 수 있음 | [Display high-bit characters in Terminal on Mac](https://support.apple.com/en-in/guide/terminal/trml103/mac) |

## 상속과 잔존

`TERM_PROGRAM`, `TERM_SESSION_ID` 같은 값은 Terminal.app이 최초 셸에 주입하는 순간부터 평범한 환경변수이며, 이후에는 운영체제의 표준 상속 규칙만 적용됩니다. 즉 "Apple Terminal이 지금 이 프로세스를 그리고 있다"는 사실이 아니라 "이 환경을 처음 만든 프로세스가 Apple Terminal이었다"는 사실만 증명합니다. 이 구분이 무너지는 대표적인 경로는 다음과 같습니다.

- **SSH**: OpenSSH는 클라이언트의 `SendEnv`와 서버의 `AcceptEnv`가 모두 허용하는 변수만 전달합니다. macOS가 배포하는 기본 `/etc/ssh/ssh_config`는 관례적으로 `LANG`과 `LC_*`만 `SendEnv`에 포함하며, `TERM_PROGRAM`이나 `TERM_SESSION_ID`는 기본값에 없습니다. 따라서 별도 설정이 없는 한 원격 호스트의 프로세스는 이 두 변수를 아예 보지 못하는 것이 정상입니다. `TERM` 자체는 `SendEnv` 메커니즘이 아니라 pty 요청(`pty-req`)을 통해 클라이언트 값이 그대로 원격 셸에 전달되므로, 로컬에서 관찰한 값이 원격에도 나타날 수 있습니다. 반대로 사용자가 `~/.ssh/config`에 `SendEnv TERM_PROGRAM`을 추가하고 서버 `sshd_config`가 이를 `AcceptEnv`로 허용하도록 관리자가 설정했다면, 원격 프로세스가 로컬 Apple Terminal의 값을 그대로 물려받아 마치 원격 호스트에서 Apple Terminal이 실행 중인 것처럼 보이는 오탐이 발생합니다.
- **tmux/screen**: 멀티플렉서 서버는 그 서버를 최초로 띄운 클라이언트 프로세스의 환경을 복사해 보관합니다. 이후 다른 터미널(예: iTerm2나 SSH로 연결한 세션)에서 같은 서버에 attach하거나 새 pane을 열어도, 서버가 새로 실행하는 셸은 서버가 최초에 캡처해 둔 오래된 `TERM_PROGRAM=Apple_Terminal` 값을 그대로 물려받을 수 있습니다. 결과적으로 화면을 실제로 표시 중인 터미널과 무관한 값이 남습니다. tmux의 `update-environment` 설정은 `attach` 시점에 지정된 변수 목록(기본값에 `TERM_PROGRAM`은 포함되지 않음)을 클라이언트 환경 값으로 갱신하지만, 이는 새로 attach하는 클라이언트 세션에만 적용되고 이미 실행 중인 pane의 프로세스 환경을 소급 갱신하지는 않습니다. 즉 `update-environment`가 설정되어 있어도 기존 pane 안에서 실행 중인 프로세스가 보는 값은 여전히 서버 시작 시점의 값입니다.
- **분리·상주 프로세스**: 터미널 창을 닫아도 `nohup`, `disown`, 데몬화된 백그라운드 프로세스는 종료되지 않고 살아남으며, 그 프로세스는 이미 상속받은 `TERM_PROGRAM=Apple_Terminal` 값을 그대로 유지합니다. 이 값은 Terminal.app 창이 사라진 뒤에도 무기한 남아 있을 수 있어, 시점상 "지금도 Apple Terminal이 붙어 있다"는 근거가 되지 못합니다.
- **위조 가능성**: 어떤 사용자나 스크립트도 `export TERM_PROGRAM=Apple_Terminal`을 실행해 임의로 값을 만들어낼 수 있습니다. 여기에는 별도 권한이 필요 없습니다.

이 모든 경로를 종합하면, `runby`가 정직하게 주장할 수 있는 것은 "이 환경을 처음 만든 터미널이 Apple Terminal이었다"까지이며, "이 프로세스의 TTY에 지금 붙어 있는 터미널이 Apple Terminal이다"라는 주장은 성립하지 않습니다.

## 실행 주체 감지에 관한 결론

`TERM_PROGRAM=Apple_Terminal`은 macOS 기본 배포 스크립트가 그 존재를 전제로 동작을 분기할 만큼 사실상의 표준으로 관찰되지만, Apple 공식 사용자 가이드는 이 변수의 이름조차 문서화하지 않습니다. `TERM_SESSION_ID`도 마찬가지로 macOS 배포 스크립트 주석에서만 공식적으로 확인되며, 세션 부여 단위(창/탭/셸 재시작)에 대한 Apple의 명시적 규격은 없습니다. `TERM`과 `LANG`은 Apple 공식 가이드가 이름과 설정 경로를 명시하지만, 둘 다 유닉스 계열 프로그램 전반이 공유하는 매우 일반적인 변수라 터미널 식별에는 부적합합니다.

따라서 `runby`가 안전하게 취할 수 있는 태도는 다음과 같습니다.

1. `TERM_PROGRAM == "Apple_Terminal"`을 macOS Terminal.app이 이 환경을 최초로 만들었다는 보조 신호로 사용하되, 공식 문서가 없는 관찰 기반 신호임을 결과에 함께 표시한다.
2. `TERM_SESSION_ID`는 같은 세션 내 프로세스를 서로 구분하는 데는 쓸 수 있지만, 그 자체로 "Apple Terminal이 만들었다"는 확정 근거로 격상하지 않는다.
3. `TERM`, `LANG`, `TERM_PROGRAM_VERSION`은 존재 여부만으로 Apple Terminal을 판정하는 근거로 쓰지 않고, 이미 `TERM_PROGRAM`으로 확정된 판정을 보강하는 부가 정보로만 취급한다.
4. tmux/screen 세션, SSH로 원격 전달된 값, 분리된 백그라운드 프로세스에서는 이 신호들이 실제 표시 터미널과 어긋날 수 있음을 감안해, `runby`의 결과를 "이 프로세스 트리를 만든 터미널"로 한정해 보고하고 "현재 화면에 붙어 있는 터미널"이라고 확정하지 않는다.

## 공식 문서

- [Use environment variables in Terminal on Mac](https://support.apple.com/guide/terminal/use-environment-variables-apd382cc5fa-4f58-4449-b20a-41c53c006f8f/mac)
- [Change profiles: Advanced preferences in Terminal on Mac](https://support.apple.com/guide/terminal/change-profiles-advanced-preferences-trmladvn/mac)
- [Display high-bit characters in Terminal on Mac](https://support.apple.com/en-in/guide/terminal/trml103/mac)
- [Terminal User Guide (전체)](https://support.apple.com/guide/terminal/welcome/mac)
- `/etc/bashrc_Apple_Terminal` — macOS가 배포하는 Apple 작성 셸 스크립트. Apple의 공식 사용자 가이드에는 포함되지 않으나 macOS 시스템 파일로 배포되는 공식 소스이며, 사본은 [gist](https://gist.github.com/floam/f535842a16226e77d014d67bade2b2f3)에서 확인할 수 있음
