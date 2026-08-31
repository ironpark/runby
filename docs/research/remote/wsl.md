---
title: WSL
slug: wsl
research_date: 2026-08-31
open_source: true
repository: https://github.com/microsoft/WSL
product_type: remote_environment
executes_agents: []
runtime_test_required: true
runtime_test_reason: 실제 Windows + WSL 환경에서 (1) 일반 대화형 셸, (2) `wsl.exe -e`로 스크립트가 시작한 프로세스, (3) 배포판 내부 systemd 서비스(특히 root 권한), (4) 배포판으로 들어오는 SSH 세션 각각에서 WSL_DISTRO_NAME·WSL_INTEROP의 실제 존재 여부를 직접 관측해야 하며, Windows Terminal에서 WSLENV를 통해 WT_SESSION이 실제로 Linux 프로세스 환경에 나타나는지도 함께 검증해야 함
---

# WSL

WSL(Windows Subsystem for Linux)은 `runby`가 다루는 다른 어떤 대상과도 성격이 다릅니다. 에이전트·CI·터미널은 모두 "이 프로세스를 무엇이 실행했는가"를 하나의 운영체제 안에서 묻지만, WSL은 **Linux 프로세스가 자신이 Windows 위의 가상화된 배포판 안에서 실행되고 있다는 사실 자체**를 환경변수로 알 수 있는지를 묻습니다. 게다가 [`docs/research/terminals/windows-terminal.md`](../terminals/windows-terminal.md)가 이미 지적했듯, WSL은 `WSLENV`라는 별도 메커니즘을 통해 **Windows 쪽 변수를 Linux 프로세스 환경에 그대로 복사**하므로, 이 라이브러리가 대상으로 하는 Linux용 Go 바이너리는 "터미널" 변수를 관측하더라도 그것이 같은 OS에서 온 값이라고 더 이상 가정할 수 없습니다. 이 문서는 그 경계 규칙을 정리합니다.

WSL의 핵심 구현은 [`microsoft/WSL`](https://github.com/microsoft/WSL) 저장소에서 2025년 5월 Build 컨퍼런스를 통해 오픈소스로 공개되었습니다([Windows Developer Blog: The Windows Subsystem for Linux is now open source](https://blogs.windows.com/windowsdeveloper/2025/05/19/the-windows-subsystem-for-linux-is-now-open-source/)). 다만 전부가 공개된 것은 아닙니다 — WSL 1의 커널 쪽 드라이버인 `Lxcore.sys`, 그리고 `\\wsl.localhost\` 경로 리다이렉션을 담당하는 `P9rdr.sys`/`p9np.dll`은 여전히 비공개입니다([Open Source WSL | Microsoft Learn](https://learn.microsoft.com/en-us/windows/wsl/opensource)). 이 문서에서 다루는 환경변수 관련 동작(interop, `WSLENV`)은 공개된 구성 요소에 속합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 컨텍스트 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `WSL_DISTRO_NAME` | 문자열 (예: `Ubuntu-22.04`) | 실행 식별 | 현재 프로세스가 속한 WSL 배포판의 이름 | 보조 신호 — 대화형 로그인 셸에서는 신뢰도 높은 식별자지만, `sudo -i`로 전환한 root나 root로 실행되는 systemd 서비스에서는 관측되지 않는다는 사례가 공식 이슈 트래커에서 보고됨. Microsoft Learn 공식 문서 페이지에 안정적 공개 계약으로 등재되어 있지는 않음 | [microsoft/WSL#9719 — Environment variable WSL_DISTRO_NAME not available as root](https://github.com/microsoft/WSL/issues/9719) |
| `WSL_INTEROP` | Unix 소켓 경로 문자열 (예: `/run/WSL/9042_interop`) | 상태·컨텍스트 | `/init`이 Windows 프로세스를 실행할 때 접속할 interop 서버 소켓을 가리킴 | 부적합(단독) — 값이 없어도 `/init`은 자신의 PID를 시작으로 부모 PID 체인을 거슬러 올라가며 `/run/WSL/${pid}_interop`을 자동으로 탐색하므로, 이 변수의 부재가 "WSL이 아니다"를 의미하지 않음. 반대로 존재는 "이 세션에서 Windows interop이 최소 한 번은 활성화되어 있었다"는 상태 신호에 가까움 | [microsoft/WSL: interop.md — 기술 문서](https://github.com/microsoft/WSL/blob/master/doc/docs/technical-documentation/interop.md) |
| `WSLENV` | 콜론 구분 변수 목록 문자열 (예: `GOPATH/p:SOMEVAR/u`) | 설정 | Windows↔WSL 경계를 넘길 변수 이름과 방향/변환 규칙을 선언 | 부적합(직접 식별용 아님) — WSL 여부 자체를 알려주지 않고 "무엇이 넘어올지"를 정의하는 설정값. 다음 절 참고 | [Microsoft Learn: WSL interop — 환경변수 공유](https://learn.microsoft.com/en-us/windows/dev-environment/wsl-interop) |
| `WSL2_GUI_APPS_ENABLED` | 불리언형 문자열로 보고됨 (예: `1`) | 상태·컨텍스트 | WSLg(GUI 앱 지원)가 활성화된 세션에서 관측된다고 커뮤니티에서 보고되는 변수 | 부적합 — Microsoft Learn 공식 문서나 `microsoft/wslg` 공식 저장소에서 이 이름을 안정적 계약으로 확인하지 못함. `.wslconfig`의 공식 설정 키는 `[wsl2] guiApplications`이며 환경변수가 아님. 이 변수는 커뮤니티 보고 수준에서만 존재가 확인되므로 감지에 사용하지 않음 | 공식 문서 미확인 — [Microsoft Learn: Advanced settings configuration in WSL (`guiApplications` 설정 키)](https://learn.microsoft.com/en-us/windows/wsl/wsl-config) 참고용으로만 인용 |
| `NAME` / `HOSTNAME` | 문자열 | 상태·컨텍스트 | 일반 Linux 배포판/호스트명 변수. `[network] hostname` 설정으로 WSL이 값을 지정할 수 있음 | 부적합 — WSL 고유 변수가 아니라 모든 Linux 시스템에 존재하는 일반 변수이며, WSL은 이 값을 자동 생성하는 옵션(`generateHosts`)을 제공할 뿐 WSL임을 표시하는 전용 마커가 아님 | [Microsoft Learn: Advanced settings configuration in WSL — `[network]` 섹션](https://learn.microsoft.com/en-us/windows/wsl/wsl-config) |

가장 신뢰할 수 있는 단일 마커는 **`WSL_DISTRO_NAME`의 존재**입니다. 다만 이는 두 가지 이유로 "적합"이 아니라 "보조 신호"로 분류됩니다. 첫째, Microsoft Learn의 정식 문서 페이지(예: `basic-commands`, `wsl-config`, `wsl-interop`)에 안정적 공개 API로 등재되어 있지 않고, 오직 커뮤니티와 공식 이슈 트래커(`microsoft/WSL`)를 통해 그 존재와 동작이 확인됩니다. 둘째, 공식 이슈([#9719](https://github.com/microsoft/WSL/issues/9719))가 보여주듯 `sudo -i`로 전환한 root 셸이나 root로 실행되는 systemd 서비스에서는 이 값이 비어 있을 수 있어, "WSL_DISTRO_NAME이 없다"가 "이 프로세스는 WSL 안에 있지 않다"를 증명하지 못합니다. 즉 이 변수는 로그인 셸 초기화 시점에 (전형적으로 `/etc/profile.d` 계열 스크립트를 통해) 주입되는 값이지, WSL 커널/init이 모든 프로세스에 무조건 심어 주는 값이 아닙니다.

**WSL 1과 WSL 2 사이에 이 마커가 달라지는지**는 공식 문서에서 명시적으로 구분하지 않습니다. `WSL_DISTRO_NAME`과 `WSL_INTEROP`은 둘 다 Windows 10 버전 1809(빌드 17763)부터 도입된 interop 계층에 속하며, 이 계층은 WSL 1과 WSL 2 모두에서 동작하도록 설계되어 있습니다([Microsoft Learn: Advanced settings configuration in WSL — Interop settings](https://learn.microsoft.com/en-us/windows/wsl/wsl-config)). 따라서 환경변수만으로는 WSL 1과 WSL 2를 구분할 수 없습니다 — 이 구분은 아래 "범위 밖" 절에서 다루는 커널 신호(`/proc/version`, `/proc/sys/kernel/osrelease`)의 영역입니다.

### 시작 방식에 따른 차이

이 라이브러리는 환경변수만 읽으므로, 같은 배포판 안에서도 프로세스를 **무엇이 시작했는가**에 따라 마커의 존재 여부가 갈립니다.

- **대화형 터미널에서 로그인 셸로 시작** — `WSL_DISTRO_NAME`, `WSL_INTEROP`이 정상적으로 상속됩니다. 로그인 셸 초기화 스크립트가 이를 설정하기 때문입니다.
- **`wsl.exe -e <command>`로 스크립트가 직접 실행** — 로그인 셸을 거치지 않을 수 있어, 배포판/셸 설정에 따라 `WSL_DISTRO_NAME`이 설정되지 않을 수 있습니다. 이는 셸 초기화 경로에 의존하는 문제이지 WSL 자체가 이 값을 항상 보장한다는 뜻이 아닙니다.
- **배포판 내부의 systemd 서비스(특히 root)** — 공식 이슈 트래커에 보고된 바와 같이 `WSL_DISTRO_NAME`이 비어 있는 사례가 확인됩니다. systemd가 관리하는 서비스는 로그인 셸의 환경을 상속하지 않고 자체 환경을 구성하기 때문입니다.
- **배포판으로 들어오는 SSH 세션** — OpenSSH는 기본적으로 클라이언트가 보낸 환경변수를 서버가 받지 않으며(`AcceptEnv` 미설정 시), `WSL_DISTRO_NAME`/`WSL_INTEROP`은 SSH로 원격 접속한 이 세션 자체의 로컬 값이므로 서버 프로세스(sshd가 새로 띄우는 로그인 셸) 쪽에서는 배포판 자신의 값이 정상적으로 설정됩니다. 다만 이 SSH 세션이 Windows 쪽에서 시작된 것인지, 순수 Linux 네트워크 안에서 시작된 것인지는 이 변수들이 구분해 주지 않습니다 — SSH로 들어온 것이 WSL 배포판이라면 그 배포판의 로그인 셸이 정상적으로 `WSL_DISTRO_NAME`을 설정할 뿐입니다.

## WSLENV: 경계를 넘는 규칙

`WSLENV`는 이 문서의 핵심입니다. WSL 자신을 식별하는 변수가 아니라, **어떤 변수가 Windows와 Linux 사이의 프로세스 생성 경계를 넘어 상대편 프로세스 환경에 나타날지를 선언하는 설정**입니다([Microsoft Learn: WSL interop — Windows and Linux integration](https://learn.microsoft.com/en-us/windows/dev-environment/wsl-interop)).

### 형식

`WSLENV`는 콜론(`:`)으로 구분된 변수 이름의 목록이며, 각 이름 뒤에 슬래시(`/`)와 하나 이상의 플래그를 선택적으로 붙일 수 있습니다.

```
WSLENV=GOPATH/l:USERPROFILE/w:SOMEVAR/wp
```

원본은 다음과 같이 설명합니다: *"WSLENV is a colon-delimited list of environment variables that should be included when launching WSL processes from Win32 or Win32 processes from WSL"*([Windows Command Line Blog: Share Environment Vars between WSL and Windows](https://devblogs.microsoft.com/commandline/share-environment-vars-between-wsl-and-windows/)).

### 플래그 (원문 그대로)

- **`/p`** — *"This flag indicates that a path should be translated between WSL paths and Win32 paths."* 양방향 경로 변환.
- **`/l`** — *"This flag indicates the value is a list of paths. In WSL, it is a colon-delimited list. In Win32, it is a semicolon-delimited list."* 경로 목록(구분자가 OS별로 다름을 자동 변환).
- **`/u`** — *"This flag indicates the value should only be included when invoking WSL from Win32."* Windows → WSL 단방향.
- **`/w`** — *"This flag indicates the value should only be included when invoking Win32 from WSL."* WSL → Windows 단방향.

**플래그가 없는 경우**, 위 원문 어디에도 "플래그 없음 = 양방향, 변환 없음"이라고 명시적으로 정의한 문장은 없습니다. 다만 `/u`·`/w`가 각각 "이 방향에서만" 포함되도록 *제한하는* 플래그로 정의되어 있고, `/p`·`/l`은 값의 *변환 방식*만 지정할 뿐 방향을 제한하지 않는다는 점에서, 플래그가 전혀 없는 이름은 **양방향으로, 값 변환 없이 그대로** 복사되는 것으로 동작합니다. 실제로 [`docs/research/terminals/windows-terminal.md`](../terminals/windows-terminal.md)가 확인한 `microsoft/terminal` 소스에서 Windows Terminal은 `WT_SESSION`/`WT_PROFILE_ID`를 `WSLENV`에 **플래그 없이** 추가하며, 그 결과 이 GUID 문자열이 변환 없이 Linux 프로세스 환경에 그대로 나타납니다.

### 방향 규칙

기본값(플래그 없음)은 양방향입니다 — Win32에서 WSL을 호출할 때도, WSL에서 Win32를 호출할 때도 포함됩니다. `/u`를 붙이면 Windows→WSL 방향으로만 좁아지고, `/w`를 붙이면 WSL→Windows 방향으로만 좁아집니다. `/p`와 `/l`은 방향이 아니라 **값의 형태**(단일 경로인지 경로 목록인지, 경로 변환이 필요한지)를 지정하는 플래그이므로 `/u`나 `/w`와 조합해서 쓸 수 있습니다(`/wp`처럼).

### 이 라이브러리에 대한 결론

Windows 쪽에서 `WSLENV`에 이름이 올라간 변수는, 그 값이 경로가 아니라면 **변환 없이 그대로** Linux 프로세스 환경 블록에 나타납니다. Windows Terminal이 `WT_SESSION`/`WT_PROFILE_ID`를 플래그 없이 등록하는 것이 정확히 이 경우이며, 그래서 Linux용으로 컴파일된 Go 바이너리가 Windows 터미널의 세션 GUID를 자신의 환경변수로 관측하게 됩니다. **이것은 버그가 아니라 WSL이 설계한 정상적인 경계 넘기**입니다. 그러나 `runby`의 관점에서는 중요한 함의가 있습니다 — "터미널을 가리키는 변수가 있다"는 사실이 더 이상 "이 프로세스가 그 터미널과 같은 OS에서 실행 중이다"를 의미하지 않습니다. `WT_SESSION`이 있다고 해서 이 프로세스가 Windows 프로세스라는 뜻이 아니라, 오히려 WSL 안에서 실행되는 Linux 프로세스일 가능성까지 포함하게 됩니다.

이 경계 넘기는 **`wsl.exe`가 시작한 프로세스 트리 안에서만** 일어납니다. `WSLENV`는 Win32 프로세스가 `wsl.exe`(또는 그에 준하는 interop 경로)를 통해 WSL 프로세스를 생성하거나, WSL 프로세스가 interop을 통해 Win32 실행 파일을 호출하는 그 생성 순간에 적용되는 규칙입니다. WSL을 다른 방식으로 시작한 경우 — 예약된 작업, 서비스, 원격 SSH로 배포판에 직접 접속한 세션 등 — 이 생성 경로를 타지 않으므로 `WSLENV`에 나열된 값이 실려 오지 않고, 해당 변수들은 나타나지 않습니다. [`docs/research/terminals/windows-terminal.md`](../terminals/windows-terminal.md)가 확인한 것과 동일한 구조입니다.

## 다른 축에 미치는 영향

`runby`는 에이전트·CI·터미널 세 축을 보고합니다. WSL은 이 셋에 서로 다른 방식으로 영향을 미칩니다.

- **터미널 축** — 가장 직접적인 영향입니다. `WSLENV`를 통해 Windows 터미널 에뮬레이터의 마커(`WT_SESSION` 등)가 Linux 프로세스 환경에 그대로 나타날 수 있으므로, WSL 안에서 실행되는 Go 바이너리도 [`docs/research/terminals/windows-terminal.md`](../terminals/windows-terminal.md)의 감지 로직을 그대로 적용할 수 있습니다. 다만 이는 `WSLENV`에 그 변수가 명시적으로 등록되어 있고 `wsl.exe` 프로세스 트리 안에서 실행되었을 때만 성립합니다.
- **CI 축** — WSL 자체는 CI 플랫폼이 아니며 CI 마커를 주입하지 않습니다. WSL 안에서 CI 러너(예: 자체 호스팅 러너)가 동작한다면 그 CI 플랫폼의 마커가 정상적으로 설정되고, WSL은 그 과정에 개입하지 않습니다. 다만 그 CI 러너가 Windows 쪽 프로세스에 의해 `wsl.exe`로 감싸져 시작되었고 CI 플랫폼 특유의 변수가 `WSLENV`에 실려 있다면(일반적이지는 않지만 사용자가 직접 설정할 수 있음) 이론적으로는 여기서도 같은 경계 넘기가 일어날 수 있습니다.
- **에이전트 축** — 질문의 핵심 사례입니다. **Windows 쪽에서 실행 중인 에이전트가 `wsl.exe`를 호출해 WSL 프로세스를 자식으로 띄우는 경우, 그 에이전트의 마커(예: 어떤 도구의 `AGENT_SESSION_ID`류 변수)가 WSL 쪽 프로세스에 나타나는가?** 답은 전적으로 `WSLENV`에 달려 있습니다 — **규칙**: 그 변수 이름이 Windows 쪽 프로세스의 `WSLENV`에 (플래그 없이 또는 `/u`를 붙여) 등록되어 있어야만 넘어옵니다. **실무적 답**: 이 저장소가 조사한 에이전트 하네스들(`docs/research/agents/` 아래 문서 참고, 예: Claude Code의 `CLAUDECODE=1`, Codex의 `CODEX_SESSION_ID` 등) 중 공식 문서·소스에서 자신의 마커를 `WSLENV`에 등록한다고 확인된 사례는 없습니다. 즉 이런 에이전트가 Windows에서 `wsl.exe`로 WSL 프로세스를 실행하더라도, 별도로 `WSLENV`를 조정하지 않는 한 그 에이전트 마커는 기본적으로 WSL 프로세스 환경에 나타나지 않습니다. Windows Terminal이 `WT_SESSION`/`WT_PROFILE_ID`에 대해 한 것처럼 에이전트 벤더가 스스로 `WSLENV`를 조정해야만 이 경계를 넘을 수 있습니다.

## 범위 밖: `/proc` 기반 신호

WSL 여부를 가장 안정적으로 판별하는 방법은 사실 환경변수가 아니라 커널이 노출하는 정보입니다.

- **`/proc/sys/kernel/osrelease`** — WSL 커널 버전 문자열에는 관례적으로 `microsoft` 또는 `WSL`이 포함됩니다(예: WSL 2에서는 `*-microsoft-standard-WSL2`, 과거 WSL 1 계열에서는 `*-Microsoft`). 여러 커뮤니티 도구(예: `wsl-rs`, Conan의 `detect_windows_subsystem`)가 이 문자열 검사를 기본 감지 방법으로 사용합니다.
- **`/proc/version`** — 같은 커널 버전 문자열이 노출되며 같은 방식으로 검사할 수 있습니다.

이 두 경로는 **환경변수가 아니므로 이 라이브러리(`runby`)의 조사 범위 밖**입니다 — `runby`는 상속된 환경변수만 읽습니다. 그럼에도 언급하는 이유는, 구조적으로 이 신호들이 환경변수보다 **더 신뢰할 수 있기 때문**입니다. 환경변수는 프로세스 생성 시점에 심어지고 이후 상속·잔존·위조될 수 있는 반면, `/proc/sys/kernel/osrelease`는 커널 자신이 매 조회마다 답하는 값이라 프로세스별로 설정되거나 사라지지 않고, `export`로 흉내 낼 수도 없습니다. 다만 이 방법에도 한계는 있습니다 — 사용자가 정의한 커스텀 커널을 WSL 2에 올리면(`.wslconfig`의 `kernel` 옵션) 문자열에서 "microsoft" 표시가 빠질 수 있어, 완벽하게 보장된 계약은 아닙니다. WSL 1과 WSL 2를 구분하는 것도 이 경로의 몫입니다 — WSL 1은 Windows 커널을 흉내 내는 호환 계층이라 `osrelease`에 실제 리눅스 커널 빌드 정보가 없는 반면, WSL 2는 Microsoft가 빌드한 실제 리눅스 커널을 부팅하므로 `microsoft-standard-WSL2`처럼 구체적인 커널 식별자가 나타납니다.

## 실행 주체 감지에 관한 결론

환경변수만으로 "이 Linux 프로세스는 WSL 안에서 실행되고 있다"를 판정하려 한다면, 가장 근거 있는 단일 신호는 `WSL_DISTRO_NAME`의 존재입니다. 그러나 이는 Microsoft Learn의 정식 공개 계약이 아니라 공식 이슈 트래커에서 관찰된 런타임 동작이며, 로그인 셸 경로를 거치지 않는 프로세스(특히 root로 실행되는 systemd 서비스)에서는 관측되지 않을 수 있습니다. `WSL_INTEROP`은 존재 자체가 "Windows interop이 이 세션에서 최소 한 번 활성화됐다"는 상태 신호일 뿐이며, `/init`이 값 없이도 소켓을 자동 탐색하도록 폴백을 두고 있어 부재가 곧 "WSL이 아니다"를 의미하지 않습니다. WSL 1과 WSL 2 사이에 이 두 변수의 동작 차이는 공식 문서에서 확인되지 않으며, 두 버전을 구분해야 한다면 환경변수가 아니라 `/proc/sys/kernel/osrelease`·`/proc/version` 같은 커널 신호(이 라이브러리의 조사 범위 밖)로 넘어가야 합니다. 마지막으로, `WSLENV`는 WSL 자신을 식별하는 신호가 아니라 **Windows-Linux 경계를 넘는 모든 변수의 통행증 목록**입니다 — 여기 이름이 오른 변수는 (Windows 쪽에서 `wsl.exe` 프로세스 트리를 타고 시작된 경우에 한해) 플래그가 없으면 양방향·무변환으로 상대편 환경에 그대로 나타나므로, `runby`가 다른 문서에서 "터미널 마커"나 "에이전트 마커"로 분류한 값이라도 WSL 경계 안에서는 그 마커의 발신처가 반드시 같은 OS라고 더 이상 보장하지 못합니다.

## 공식 문서

- [Microsoft Learn: WSL interop — Windows and Linux integration (WSLENV, 환경변수 공유)](https://learn.microsoft.com/en-us/windows/dev-environment/wsl-interop)
- [Windows Command Line Blog: Share Environment Vars between WSL and Windows (WSLENV 플래그 원문)](https://devblogs.microsoft.com/commandline/share-environment-vars-between-wsl-and-windows/)
- [Microsoft Learn: Advanced settings configuration in WSL (`wsl.conf`의 `[interop]`·`[network]` 섹션, `.wslconfig`의 `guiApplications`)](https://learn.microsoft.com/en-us/windows/wsl/wsl-config)
- [Microsoft Learn: Basic commands for WSL](https://learn.microsoft.com/en-us/windows/wsl/basic-commands)
- [Microsoft Learn: Open Source WSL (오픈소스 전환 범위, 비공개로 남은 컴포넌트)](https://learn.microsoft.com/en-us/windows/wsl/opensource)
- [Windows Developer Blog: The Windows Subsystem for Linux is now open source](https://blogs.windows.com/windowsdeveloper/2025/05/19/the-windows-subsystem-for-linux-is-now-open-source/)
- [공식 저장소: `microsoft/WSL`](https://github.com/microsoft/WSL)
- [`microsoft/WSL`: `interop.md` 기술 문서 (`WSL_INTEROP`, `/init`의 소켓 탐색 폴백)](https://github.com/microsoft/WSL/blob/master/doc/docs/technical-documentation/interop.md)
- [`microsoft/WSL#9719` — `WSL_DISTRO_NAME`이 root/systemd 컨텍스트에서 관측되지 않는 사례](https://github.com/microsoft/WSL/issues/9719)
- [`microsoft/WSL#11920` — `interop.enabled = false` 설정에도 `WSLInterop`이 남아있는 사례 보고](https://github.com/microsoft/WSL/issues/11920)
- [`docs/research/terminals/windows-terminal.md` — `WT_SESSION`/`WT_PROFILE_ID`가 `WSLENV`를 통해 WSL로 넘어가는 근거 소스](../terminals/windows-terminal.md)
