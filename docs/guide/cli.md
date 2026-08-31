# CLI

Go 코드에 연결하려면 [시작하기](getting-started.md)를 참고하세요. CLI는 셸 스크립트에서 같은 질문을 하거나, 버그 리포트에 실행 환경 요약을 첨부할 때 사용합니다.

```
go install github.com/ironpark/runby/cmd/runby@latest
```

## 사용법

```
runby [-json] [-v]     사람이 읽는 요약, 또는 Result 전체 JSON
runby is <축>          종료 코드로만 답. 축: agent ci terminal remote runner tty
runby chain            "paseo>codex" 한 줄. 감지 실패 시 "unknown"
```

| 플래그 | 설명 |
|---|---|
| `-json` | `Result` 전체를 JSON으로 출력. 필드는 라이브러리와 동일합니다 |
| `-v` | 각 축이 근거로 삼은 환경변수 **이름**과 조상 프로세스 목록을 함께 출력 |

## 기본 보고

```
$ runby
agent     paseo>claude-code
            paseo          l3  orchestrator  delegated     definite  살아 있는 조상 pid=2540
            claude-code    l1  harness       first-party   definite  살아 있는 조상 pid=4344
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

| 종료 코드 | 의미 |
|---|---|
| `0` | 정상. `is`에서는 해당 축이 참 |
| `1` | `is`에서 해당 축이 거짓 |
| `2` | 사용법 오류 (알 수 없는 명령·축·플래그) |

JSON은 `jq`로 바로 다룰 수 있습니다.

```sh
runby -json | jq -r '.layers[] | select(.ancestor_pid != null) | .agent'
```

## 환경변수 값은 출력하지 않습니다

어떤 모드에서도 환경변수의 **값**은 출력하지 않습니다. `-v`가 보여주는 `←` 줄은 변수 **이름**뿐이고, 이는 라이브러리의 `Evidence` 규칙과 같습니다 — 값에는 토큰이 들어 있을 수 있기 때문입니다.

**단, `-json`은 다릅니다.** 환경변수를 그대로 덤프하지는 않지만, 제품이 스스로 광고한 식별자와 경로가 값으로 들어갑니다.

| 필드 | 내용 |
|---|---|
| `layers[].agent_id`, `layers[].session_id` | 에이전트/세션 UUID |
| `layers[].paths.*` | 작업 디렉터리, 데이터 디렉터리 |
| `layers[].extra`, `ci.extra`, `terminal.extra`, `remote[].extra` | 제품 전용 값 (worktree 경로, 호스트 ID 등) |
| `process.ancestors[].path` | 조상 프로세스의 실행 파일 **전체 경로** |

여기에는 사용자 이름이 들어간 홈 디렉터리 경로, 저장소 위치, 세션 식별자가 포함될 수 있습니다. 텍스트 모드는 이 값들을 찍지 않고 이름과 PID만 보여주므로, **버그 리포트에 붙일 때는 `-json`보다 `runby -v`가 안전합니다.**

`-json`을 로그나 텔레메트리로 보낸다면 필요한 필드만 골라 쓰십시오.

```sh
runby -json | jq '{chain: [.layers[].agent] | join(">"), ci: .ci.provider, tty: .tty.interactive}'
```

## 라이브러리와의 관계

`is`의 각 축은 라이브러리의 같은 판정 결과를 사용합니다.

| CLI | 라이브러리 |
|---|---|
| `runby is agent` | `runby.IsAgent()` |
| `runby is runner` | `runby.HasRunner()` |
| `runby is ci` | `runby.IsCI()` |
| `runby is terminal` | `runby.HasTerminal()` |
| `runby is remote` | `runby.IsRemote()` |
| `runby is tty` | `runby.Current().TTY.Interactive` |

CLI가 별도의 판단을 갖지 않는다는 사실은 테스트로 고정되어 있습니다.
