# 멀티플렉서와 원격 실행 계층 문서

이 계층은 사용자와 이 프로세스 **사이에** 끼어 있는 것들입니다. 양식과 front matter 필드의 의미는 [상위 문서](../README.md)와 같으며, `product_type`은 `terminal_multiplexer` 또는 `remote_environment`입니다.

## 이 범주가 다른 이유

에이전트·CI·터미널 문서는 각 제품이 **어떤 변수를 설정하는가**를 다룹니다. 이 범주는 다릅니다. 여기 있는 제품들은 자기 변수를 추가할 뿐 아니라 **다른 축의 변수가 살아남을지, 어떻게 변형될지를 결정합니다.**

| 계층 | 통과 규칙을 정하는 것 |
|---|---|
| [tmux](tmux.md) | `update-environment` — attach 시점에 목록에 있는 변수만 클라이언트에서 복사 |
| [GNU Screen](gnu-screen.md) | **없음** — 서버 시작 시점 환경에 영구 고정 |
| [Zellij](zellij.md) | **없음** — 서버 시작 시점 환경에 영구 고정 |
| [OpenSSH](openssh.md) | `SendEnv`·`AcceptEnv`·`SetEnv` — 어떤 변수가 원격으로 건너갈지 |
| [WSL](wsl.md) | `WSLENV` — 어떤 변수가 Windows/Linux 경계를 넘을지 |
| [Dev Containers](devcontainers.md) | `containerEnv`(컨테이너 전체)·`remoteEnv`(연결 세션 한정) |

따라서 이 문서들은 다른 축의 신뢰도를 해석하는 **전제**가 됩니다.

## 멀티플렉서

세 제품 모두 **이미 실행 중인 pane의 환경은 어떤 방법으로도 갱신되지 않습니다.** Unix 프로세스 의미론상 불가능합니다. 차이는 새 pane을 만들 때 무엇을 덮어쓰느냐입니다.

| | tmux | GNU Screen | Zellij |
|---|---|---|---|
| 마커 | `TMUX` | `STY` | `ZELLIJ` (값이 **리터럴 `"0"`**) |
| `TERM_PROGRAM` | `tmux`로 **덮어씀** | **통과** | 통과 (소스에 기록 경로 없음) |
| `TERM` | 덮어씀 | 덮어씀 (`screen`) | 덮어쓰지 않음 |
| 세션 식별자 | `TMUX_PANE` | `WINDOW` | `ZELLIJ_SESSION_NAME`, `ZELLIJ_PANE_ID` |
| 갱신 기제 | `update-environment` | 없음 | 없음 |
| 런타임 환경 조작 | `set-environment`/`show-environment` | `setenv`/`unsetenv` (새 window만) | 없음 (정적 KDL `env {}`만) |

### tmux와 Screen은 정반대로 실패합니다

- **tmux 안** — `TERM_PROGRAM`이 `tmux`로 덮여 iTerm2·Apple Terminal·WezTerm·Ghostty·Warp·Zed 6종의 정체성이 **소실**됩니다(거짓 음성). 통과하는 것은 `TERM_PROGRAM`을 쓰지 않는 kitty·Windows Terminal·Alacritty·Konsole·GNOME Terminal 5종이며, 이쪽이 **낡았지만 그럴듯한 거짓 양성**을 만듭니다.
- **Screen 안** — `TERM_PROGRAM`도 통과하므로 **12종 전부가 잔존 가능**하고, `update-environment`에 해당하는 기제가 없어 설정으로 완화할 수도 없습니다. `screen -d -r`로 다른 터미널에서 재접속해도 기존 window는 전혀 갱신되지 않습니다.

tmux 3.7 기준 `update-environment` 기본 목록입니다. 터미널 식별 변수는 하나도 없습니다.

```
DISPLAY KRB5CCNAME MSYSTEM SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION
WAYLAND_DISPLAY WINDOWID XAUTHORITY XDG_CURRENT_DESKTOP XDG_SESSION_DESKTOP XDG_SESSION_TYPE
```

### 잔존은 터미널 축만의 문제가 아닙니다

오래 살아 있는 멀티플렉서 서버는 처음 시작될 때의 **CI·에이전트 마커**도 나중에 만들어진 pane에 그대로 물려줍니다. CI 잡 안에서 시작한 tmux 서버에 사람이 몇 시간 뒤 붙어 새 pane을 열면 그 CI 마커가 여전히 보입니다.

### `ZELLIJ`는 불리언이 아닙니다

Zellij의 마커 값은 리터럴 문자열 **`"0"`**입니다. Go의 `strconv.ParseBool`은 이를 `false`로 읽으므로, 불리언으로 파싱하면 실제 Zellij 세션 안에서 조용히 감지에 실패합니다. **반드시 존재 여부로만 판정해야 합니다.**

## 원격 실행 계층

### OpenSSH — 다른 문서들이 인용하는 기준

`SendEnv`와 `AcceptEnv`의 **문서화된 기본값은 둘 다 "아무 변수도 보내지·받지 않음"**입니다. 실무에서 `LANG`/`LC_*`가 넘어가는 것은 Debian·Ubuntu 등이 배포하는 `/etc/ssh/ssh_config`와 `sshd_config`가 그렇게 설정해 두었기 때문이며, OpenSSH 소프트웨어의 기본 동작이 아닙니다. 이 구분은 널리 잘못 알려져 있습니다.

`TERM`만은 예외입니다. `SendEnv`와 무관하게 SSH 프로토콜의 pty 요청에 실려 항상 전달되므로, 아무리 잠근 설정에서도 살아남습니다.

| 변수 | 판정 |
|---|---|
| `SSH_CONNECTION` | 세션 존재의 가장 신뢰도 높은 마커. `클라이언트 IP·포트·서버 IP·포트` 4개 값 |
| `SSH_CLIENT` | 폐기 대상 — 소스에 `/* SSH_CLIENT deprecated */` 주석 |
| `SSH_TTY` | pty가 실제 할당됐을 때만 존재. `ssh host command`는 `SSH_CONNECTION`만 있고 이건 없음 |
| `SSH_AUTH_SOCK` | **SSH 세션 마커가 아님** — 로컬 데스크톱에서도 `ssh-agent`가 설정하는 흔한 오탐원 |

### 감지할 수 없는 것들

- **[Mosh](mosh.md)** — 환경변수로 감지 **불가능**합니다. `MOSH_KEY`는 제거되는 게 아니라 애초에 원격 셸 환경에 들어간 적이 없고(`mosh-server`가 메모리에서 생성해 부트스트랩 채널로 출력할 뿐), 나머지 `MOSH_*`는 전부 클라이언트 전용입니다. 정상적인 Mosh 세션 안에는 `MOSH_*`가 하나도 없는 것이 정상입니다. `MOSH_KEY`는 **자격증명**이므로 어떤 경우에도 읽거나 기록해서는 안 됩니다.
- **컨테이너 일반** — Docker·Podman은 식별 환경변수를 **전혀 설정하지 않습니다.** 관례적 감지는 `/.dockerenv`, `/run/.containerenv`, cgroup 같은 **파일시스템 경로**를 씁니다. 환경변수만 읽는 `runby`는 "컨테이너 안"이라는 사실을 원리적으로 알 수 없으며, VS Code나 Codespaces처럼 도구가 스스로 광고한 경우만 알 수 있습니다. `HOSTNAME`이 짧은 16진 문자열인 것은 근거가 아니라 추측이므로 쓰면 안 됩니다.

### Mosh 안의 `SSH_CONNECTION`은 함정입니다

`SSH_CONNECTION`은 Mosh 세션에 **살아남습니다**. `mosh-server`가 읽기만 하고 unset하지 않기 때문입니다. 그런데 Mosh의 실제 연결은 UDP이고 로밍으로 IP가 바뀌므로, 이 값은 세션 시작 시점의 수명 짧은 **부트스트랩 연결**만 설명합니다. 일반 SSH의 잔존보다 심합니다.

### 경계를 넘는 규칙

**WSL `WSLENV`** — 콜론 구분 변수 이름 목록이며 각 이름에 플래그를 붙일 수 있습니다.

| 플래그 | 의미 |
|---|---|
| `/p` | 경로를 WSL↔Win32 양방향 변환 |
| `/l` | 경로 목록 (WSL은 `:`, Win32는 `;` 구분) |
| `/u` | Win32에서 WSL 호출 시에만 포함 |
| `/w` | WSL에서 Win32 호출 시에만 포함 |
| 없음 | 양방향, 경로 변환 없음 |

Windows Terminal이 `WT_SESSION`·`WT_PROFILE_ID`를 **플래그 없이** 추가하므로, Linux용으로 빌드한 Go 바이너리가 Windows 터미널의 세션 GUID를 그대로 봅니다. 버그가 아니라 정상 동작이지만, **"터미널" 변수가 더 이상 "같은 OS"를 뜻하지 않는다**는 의미입니다. 반대로 **에이전트 마커는 기본적으로 경계를 넘지 않습니다** — 벤더가 `WSLENV`에 명시해야 하는데, 이 저장소에 문서화된 에이전트 중 그렇게 하는 곳은 확인되지 않았습니다.

`WSL_DISTRO_NAME`은 root·systemd 서비스에서 부재가 확인됐고 `WSL_INTEROP`도 없을 수 있으므로, **부재가 "WSL 아님"을 뜻하지 않습니다.**

**Dev Containers** — 사양 자체는 식별 환경변수를 **전혀 요구하지 않습니다.** `REMOTE_CONTAINERS=true`는 VS Code 확장만의 구현 세부사항이고 `devcontainer` CLI·JetBrains Gateway·DevPod는 설정하지 않습니다. `DEVCONTAINER=true`는 관례로 제안됐다가 보류된 상태입니다. `containerEnv`는 컨테이너 전체에, `remoteEnv`는 연결 세션에만 적용되므로 백그라운드 프로세스나 외부 `docker exec` 셸은 `remoteEnv`를 보지 못합니다.

### 클라우드 개발 환경은 CI가 아닙니다

`runby`는 `CI=true`만 있으면 `CIProviderGeneric`으로 보고하므로, 대화형 클라우드 개발 환경이 CI 잡으로 오보고되면 실제 결함입니다.

| | 마커 | `CI` | `GITHUB_ACTIONS` |
|---|---|---|---|
| [GitHub Codespaces](github-codespaces.md) | `CODESPACES=true` | 공식 목록에 없음 | **설정하지 않음** |
| [Gitpod](gitpod.md) | `GITPOD_WORKSPACE_ID` | 근거 없음 | — |

두 `CI` 판정은 모두 **부재 기반 추론**입니다. 공식 목록에 없을 뿐 "설정하지 않는다"는 명시 문장은 없으므로, 두 문서 모두 `runtime_test_required: true`로 표시했습니다. 구현이 이 가정에 의존하기 전에 실제 환경에서 확인해야 합니다.

Codespaces는 `GITHUB_*` 네임스페이스의 세 번째 소비자이지만(GitHub Actions·Forgejo에 이어), `GITHUB_ACTIONS`를 설정하지 않고 run/job/attempt 계열도 전혀 없어 겹치는 이름은 `GITHUB_REPOSITORY`·`GITHUB_SERVER_URL`·`GITHUB_API_URL`·`GITHUB_TOKEN` 4개뿐입니다.

원격 개발 환경 안에서 보이는 터미널·에이전트 마커는 **컨테이너의 것**이지 사용자 노트북의 것이 아닙니다. 로컬 환경은 그곳까지 전달되지 않습니다.

## `runby`의 구현 대응

| 계층 | `RemotePlatform` | `Kind` | 마커 |
|---|---|---|---|
| tmux | `RemoteTmux` | `multiplexer` | `TMUX` |
| GNU Screen | `RemoteScreen` | `multiplexer` | `STY` |
| Zellij | `RemoteZellij` | `multiplexer` | `ZELLIJ` (존재 검사) |
| OpenSSH | `RemoteSSH` | `environment` | `SSH_CONNECTION` |
| WSL | `RemoteWSL` | `environment` | `WSL_DISTRO_NAME` 또는 `WSL_INTEROP` |
| GitHub Codespaces | `RemoteCodespaces` | `environment` | `CODESPACES=true` |
| Gitpod | `RemoteGitpod` | `environment` | `GITPOD_WORKSPACE_ID` |
| Dev Containers | `RemoteDevContainer` | `environment` | `REMOTE_CONTAINERS` 또는 `DEVCONTAINER` |

Mosh와 컨테이너 런타임 일반은 위에서 설명한 이유로 감지하지 않습니다. 죽은 코드를 넣는 대신 한계로 문서화했습니다.

`Result.Remote`는 슬라이스이며 **순서는 감지 순서일 뿐 중첩 순서가 아닙니다.** 환경변수로는 어느 계층이 바깥인지 증명할 수 없습니다.

멀티플렉서가 잡히면 `Terminal.Confidence`를 `probable`로 낮춥니다. SSH는 낮추지 않습니다 — 값이 낡은 것이 아니라 다른 머신을 가리킬 수 있다는 뜻이므로, `RemoteSSH` 계층의 존재로 그 사실을 표현합니다.

오탐 때문에 일부러 마커로 쓰지 않는 변수는 `SSH_AUTH_SOCK`(로컬 `ssh-agent`도 설정), `WINDOW`(이름이 너무 일반적), `LC_TERMINAL`(SSH를 건너감)입니다.

## 실행 검증이 필요한 항목

`runtime_test_required: true`로 표시된 문서는 공식 문서 조사만으로 확정하지 못한 부분이 있습니다.

| 문서 | 확인이 필요한 것 |
|---|---|
| tmux | 버전별 `update-environment` 기본 목록 변화 |
| Mosh | `MOSH_KEY` 무유출 주장(보안 관련) 및 man page와 소스의 불일치 |
| WSL | `WSL_DISTRO_NAME` 부재 조건 |
| Dev Containers | 도구별 `REMOTE_CONTAINERS` 설정 여부 |
| GitHub Codespaces | `CI` 설정 여부 |
| Gitpod | `CI` 설정 여부, 신세대(Ona) 워크스페이스의 `GITPOD_*` 유지 여부 |
