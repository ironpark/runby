# Remote 축

사용자와 이 프로세스 사이에 낀 계층입니다. 이 축이 따로 있는 이유는 여기 있는 것들이 **자기 변수를 추가하는 데 그치지 않고 다른 축의 변수가 살아남을지를 결정**하기 때문입니다 — tmux는 `update-environment`로, OpenSSH는 `SendEnv`/`AcceptEnv`로, WSL은 `WSLENV`로, Dev Containers는 `containerEnv`/`remoteEnv`로 거릅니다. 따라서 이 축의 감지 결과는 독립된 사실이 아니라 **다른 축을 얼마나 믿을 수 있는지에 대한 단서**입니다.

원격 계층 존재 여부만 필요하면 `IsRemote()`, 오래된 환경 가능성을 확인하려면 `Multiplexer()`를 사용하세요.

**여러 계층이 동시에 존재할 수 있으므로 슬라이스입니다.** Codespace에 SSH로 붙어 tmux를 쓰면 세 계층이 함께 잡힙니다.

```go
result := runby.Detect()
result.IsRemote()                      // 낀 계층이 있는가
result.Remote(runby.RemoteTmux)         // (Remote, bool) — 특정 계층
result.Multiplexer()                    // (Remote, bool) — 잔존 위험의 주 원인
```

| 계층 | `RemotePlatform` | `Kind` | 마커 |
|---|---|---|---|
| tmux | `RemoteTmux` | `multiplexer` | `TMUX` |
| GNU Screen | `RemoteScreen` | `multiplexer` | `STY` |
| Zellij | `RemoteZellij` | `multiplexer` | `ZELLIJ` (값은 리터럴 `"0"`) |
| OpenSSH | `RemoteSSH` | `environment` | `SSH_CONNECTION` |
| WSL | `RemoteWSL` | `environment` | `WSL_DISTRO_NAME` 또는 `WSL_INTEROP` |
| GitHub Codespaces | `RemoteCodespaces` | `environment` | `CODESPACES=true` |
| Gitpod | `RemoteGitpod` | `environment` | `GITPOD_WORKSPACE_ID` |
| Dev Containers | `RemoteDevContainer` | `environment` | `REMOTE_CONTAINERS` 또는 `DEVCONTAINER` |

`Remotes`의 순서는 **감지 순서일 뿐 중첩 순서가 아닙니다.** 환경변수로는 어느 계층이 바깥인지 증명할 수 없습니다.

## 멀티플렉서만 신뢰도를 낮춥니다

멀티플렉서 서버는 처음 붙은 클라이언트의 환경을 유지하고 **이미 실행 중인 pane의 환경은 갱신할 수 없습니다.** 그래서 `Multiplexer()`가 잡히면 `Terminal.Confidence`를 `probable`로 낮춥니다. SSH는 다릅니다 — 터미널이 다른 머신에 있을 수 있다는 뜻이지만 값이 낡은 것은 아니므로, 신뢰도를 낮추는 대신 `RemoteSSH` 계층의 존재로 그 사실을 표현합니다.

tmux와 Screen은 실패 방향이 정반대입니다. tmux는 `TERM_PROGRAM`을 `tmux`로 덮어써 그 계열 터미널 6종의 정체성을 **지우고**(거짓 음성), 건드리지 않는 마커만 통과시켜 **잔존**시킵니다(거짓 양성). Screen과 Zellij는 `TERM_PROGRAM`을 덮어쓰지 않고 갱신 기제도 없어 **모든 마커가 잔존 가능**합니다.

잔존 위험은 터미널 축에만 걸리지 않습니다. 오래 사는 서버는 처음 시작될 때의 CI·에이전트 마커도 나중 pane에 물려줍니다. `runby`는 이를 신뢰도로 자동 반영하지 않고 `Multiplexer()`라는 사실만 노출하며, 세 축에 대한 해석은 소비자에게 맡깁니다.

## 감지할 수 없는 것

- **Mosh** — 원리적으로 불가능합니다. `MOSH_KEY`는 제거되는 게 아니라 애초에 원격 셸 환경에 들어간 적이 없고, 나머지 `MOSH_*`는 전부 클라이언트 전용입니다. 정상 세션에는 `MOSH_*`가 하나도 없습니다. `MOSH_KEY`는 **자격증명**이므로 어떤 경우에도 읽거나 기록하면 안 됩니다.
- **컨테이너 일반** — Docker·Podman은 식별 환경변수를 설정하지 않습니다. 관례적 감지는 `/.dockerenv`, `/run/.containerenv`, cgroup 같은 **파일시스템 경로**를 쓰므로 환경변수만 읽는 이 라이브러리의 범위 밖입니다. Dev Containers나 Codespaces처럼 도구가 스스로 광고한 경우만 보입니다. `HOSTNAME`이 짧은 16진 문자열인 것은 근거가 아니라 추측이라 쓰지 않습니다.

## 오탐을 피하려고 일부러 쓰지 않는 변수

- **`SSH_AUTH_SOCK`** — `ssh-agent`가 로컬 데스크톱에서도 설정하므로 SSH 세션 마커가 아닙니다.
- **`WINDOW`** (Screen) — 무조건 설정되지만 다른 소프트웨어와 이름이 겹치기 쉬워 `STY`와 함께일 때만 컨텍스트로 씁니다.
- **`LC_TERMINAL`** — 배포판이 배포하는 `SendEnv LC_*` 설정에 걸려 SSH를 건너가므로 다른 머신의 터미널을 가리킬 수 있습니다.

조사 근거는 [`docs/research/remote/`](../research/remote/)에 있습니다.
