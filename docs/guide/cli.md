# CLI

Go 코드에 연결하려면 [시작하기](getting-started.md)를 참고하세요. CLI는 셸 스크립트에서 같은 질문을 하거나, 버그 리포트에 실행 환경 요약을 첨부할 때 사용합니다.

```
go install github.com/ironpark/runby/cmd/runby@latest
```

## 사용법

```
runby [-json] [-v]     사람이 읽는 요약, 또는 Result 전체 JSON
runby is <축> [제품]   종료 코드로만 답
runby chain            "paseo>codex" 한 줄. 감지 실패 시 "unknown"
```

| 축 | 제품을 덧붙일 수 있나 |
|---|---|
| `agent` `ci` `terminal` `remote` `runner` | 예 |
| `tty` | 아니오 — 제품 차원이 없습니다 |

| 플래그 | 설명 |
|---|---|
| `-json` | `Result` 전체를 JSON으로 출력. 필드는 라이브러리와 동일합니다 |
| `-v` | 각 축이 근거로 삼은 환경변수 **이름**과 조상 프로세스 목록을 함께 출력 |

## 기본 보고

```
$ runby
agent     paseo>claude-code
            paseo          orchestrator  delegated     definite  살아 있는 조상 pid=2540
            claude-code    harness       first-party   definite  살아 있는 조상 pid=4344
ci        -
terminal  ghostty (definite)
remote    tmux (multiplexer), openssh (environment)
runner    npm (script) test, gnu-make (script)
tty       대화형 (stdin과 출력이 터미널)
process   조상 7개

주의: tmux 안에서 실행 중입니다. 멀티플렉서는 이미 열린 pane의 환경을
갱신하지 않으므로 터미널 축의 값이 낡았을 수 있습니다.
```

`-v`를 붙이면 각 줄 아래에 `←`로 근거가 된 환경변수 이름과, 조상 프로세스 체인이 붙습니다.

## 셸에서 쓰기

`is`는 아무것도 출력하지 않고 **종료 코드로만** 답하므로 조건문에 그대로 넣을 수 있습니다.

```sh
if runby is agent; then
	export NO_COLOR=1        # 에이전트가 실행했으면 색을 끕니다
fi

if runby is ci; then
	go test ./... -race
fi
```

제품 이름을 덧붙이면 **어떤** 제품인지까지 좁힙니다.

```sh
if runby is agent codex; then
	echo "Codex가 실행했습니다"
fi

if runby is remote tmux; then
	echo "tmux 안입니다 — 터미널 축의 값이 낡았을 수 있습니다"
fi

runby is ci github-actions
runby is runner npm
runby is terminal ghostty
```

제품 이름은 `-json`에 나오는 슬러그와 **같습니다** — `claude-code`, `github-actions`, `gnu-make`처럼요. 사용 가능한 목록은 [지원 범위](../../README.md#지원-범위)에 있고, 오타를 냈을 때 stderr에도 전부 출력됩니다.

에이전트·원격·실행 도구 축은 계층이 여럿일 수 있으므로 **어느 계층에든 있으면 참**입니다. Paseo가 Codex를 구동했다면 `runby is agent paseo`와 `runby is agent codex`가 둘 다 0입니다.

| 종료 코드 | 의미 |
|---|---|
| `0` | 정상. `is`에서는 참 |
| `1` | `is`에서 거짓 |
| `2` | 사용법 오류 (알 수 없는 명령·축·**제품**·플래그) |

**오타는 거짓이 아니라 오류입니다.** `runby is agent codexx`는 1이 아니라 2로 답합니다. 조용히 1을 돌려주면 스크립트가 영원히 잘못된 분기를 타고 아무도 이유를 알 수 없기 때문입니다. 조건문에서 이 구분이 필요하면 `2`를 따로 처리하십시오.

```sh
runby is agent codex
case $? in
	0) echo "codex" ;;
	1) echo "codex 아님" ;;
	*) echo "runby 호출이 잘못됨" >&2; exit 2 ;;
esac
```

JSON은 `jq`로 바로 다룰 수 있습니다.

```sh
runby -json | jq -r '.agents[] | select(.ancestor_pid != null) | .name'
```

## 환경변수 값은 출력하지 않습니다

어떤 모드에서도 환경변수의 **값**은 출력하지 않습니다. `-v`가 보여주는 `←` 줄은 변수 **이름**뿐이고, 이는 라이브러리의 `Evidence` 규칙과 같습니다 — 값에는 토큰이 들어 있을 수 있기 때문입니다.

**단, `-json`은 다릅니다.** 환경변수를 그대로 덤프하지는 않지만, 제품이 스스로 광고한 식별자와 경로가 값으로 들어갑니다.

| 필드 | 내용 |
|---|---|
| `agents[].agent_id`, `agents[].session_id` | 에이전트/세션 UUID |
| `agents[].paths.*` | 작업 디렉터리, 데이터 디렉터리 |
| `agents[].extra`, `ci.extra`, `terminal.extra`, `remotes[].extra` | 제품 전용 값 (worktree 경로, 호스트 ID 등) |
| `process.ancestors[].path` | 조상 프로세스의 실행 파일 **전체 경로** |

여기에는 사용자 이름이 들어간 홈 디렉터리 경로, 저장소 위치, 세션 식별자가 포함될 수 있습니다. 텍스트 모드는 이 값들을 찍지 않고 이름과 PID만 보여주므로, **버그 리포트에 붙일 때는 `-json`보다 `runby -v`가 안전합니다.**

`-json`을 로그나 텔레메트리로 보낸다면 필요한 필드만 골라 쓰십시오.

```sh
runby -json | jq '{chain: [.agents[].name] | join(">"), ci: .ci.provider, tty: .tty.interactive}'
```

## 라이브러리와의 관계

`is`의 각 축은 라이브러리의 같은 판정 결과를 사용합니다.

| CLI | 라이브러리 (`result := runby.Current()`) |
|---|---|
| `runby is agent` | `result.IsAgent()` |
| `runby is agent codex` | `_, ok := result.Agent(runby.AgentCodex)` |
| `runby is runner` | `result.HasRunner()` |
| `runby is runner npm` | `_, ok := result.Runner(runby.RunnerNPM)` |
| `runby is ci` | `result.IsCI()` |
| `runby is ci github-actions` | `result.CI.Provider == runby.CIProviderGitHubActions` |
| `runby is terminal` | `result.HasTerminal()` |
| `runby is terminal ghostty` | `result.Terminal.Program == runby.TerminalGhostty` |
| `runby is remote` | `result.IsRemote()` |
| `runby is remote tmux` | `_, ok := result.Remote(runby.RemoteTmux)` |
| `runby is tty` | `result.TTY.Interactive` |

CLI가 별도의 판단을 갖지 않는다는 사실은 테스트로 고정되어 있습니다.
