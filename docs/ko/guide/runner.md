# 실행 주체 축 (`Result.Runners`)

**무엇이 이 프로세스를 직접 실행했는가** — 패키지 매니저 스크립트인가, 빌드 레시피인가, 서비스 관리자인가.

단순히 도구 실행 여부만 필요하면 `HasRunner()`를 사용하세요. 프롬프트나 출력 형식을 정하려면 도구 이름보다 `Runner.Kind`의 `script`, `hook`, `service` 구분이 더 유용합니다.

```go
result := runby.Detect()
result.HasRunner()                          // 도구가 실행했는가
result.Runner(runby.RunnerNPM)              // (Runner, bool)
result.RunnerOfKind(runby.RunnerKindService) // (Runner, bool) — 데몬인가
```

## 이 축이 왜 필요한가

`npm test`로 시작된 프로세스를 생각해 보십시오.

| | 값 | 뜻 |
|---|---|---|
| `TTY.Interactive` | `true` | 터미널이 붙어 있다 |
| `IsCI()` | `false` | CI가 아니다 |
| `IsAgent()` | `false` | 에이전트가 아니다 |

세 축 모두 **"사람이 대화형으로 친 명령"**이라고 답합니다. 그리고 셋 다 틀리지 않았습니다 — 다만 아무도 `package.json`의 스크립트가 실행했다는 사실을 말해 주지 않습니다. 출력 형식, 색상, 진행률 표시, 프롬프트 여부를 결정하려는 도구에게 이 차이는 실질적입니다.

`systemd` 서비스도 마찬가지입니다. "이 프로세스가 데몬으로 돌고 있는가"를 상속된 환경변수만으로 답할 수 있는 사실상 유일한 경로가 `INVOCATION_ID`입니다.

## 감지 대상

| 도구 | `RunnerTool` | `Kind` | 마커 | `Task` |
|---|---|---|---|---|
| npm | `RunnerNPM` | `script` | `npm_config_user_agent`가 `npm/`으로 시작 | 스크립트 이름 |
| pnpm | `RunnerPNPM` | `script` | `pnpm/`으로 시작 | 스크립트 이름 |
| Bun | `RunnerBun` | `script` | `bun/`으로 시작 | 스크립트 이름 |
| GNU Make | `RunnerMake` | `script` | `MAKELEVEL` | 없음 |
| systemd | `RunnerSystemd` | `service` | `INVOCATION_ID` | 없음 |
| pre-commit | `RunnerPreCommit` | `hook` | `PRE_COMMIT=1` | 없음 |

`Kind`는 "누가 실행했나"가 아니라 **무엇을 뜻하나**를 답합니다.

- `script` — 명령이 사람 손이 아니라 프로젝트 파일에 적혀 있었다
- `hook` — 요청이 아니라 저장소 이벤트에 반응해 실행됐다
- `service` — 서비스 관리자가 유닛의 일부로 실행했다. **아무도 출력을 보고 있지 않다**

## 계층은 여러 개가 정상입니다

remote 축처럼 슬라이스입니다. pre-commit 훅이 npm 스크립트를 부르고 그 스크립트가 make를 부르는 것은 흔한 구성이고, 세 계층이 모두 참입니다.

```go
for _, r := range result.Runners {
	log.Printf("%s (%s) %s", r.Tool, r.Kind, r.Task)
}
// npm (script) lint
// gnu-make (script)
// pre-commit (hook)
```

순서는 **감지 순서이지 중첩 순서가 아닙니다.** 환경변수는 누가 누구를 불렀는지 증명하지 못합니다.

## CI 축과 독립입니다

CI 잡 안에서 `npm test`를 돌리면 두 축이 함께 잡힙니다. `CI`는 "어디서 도는가", `Runners`는 "무엇이 직접 실행했는가"로 서로 다른 질문입니다.

## 감지할 수 없는 것

### git 훅

**환경변수로는 불가능합니다.** git 2.55.0 실측에서 `post-checkout` 훅과 평범한 git 별칭이 **완전히 같은 `GIT_*` 집합**을 받는 것을 확인했습니다.

| 실행 맥락 | 주입되는 `GIT_*` |
|---|---|
| `post-checkout` 훅 | `GIT_EDITOR`, `GIT_EXEC_PATH`, `GIT_PREFIX` |
| git 별칭 (훅 아님) | `GIT_CONFIG_PARAMETERS`, `GIT_EDITOR`, `GIT_EXEC_PATH`, `GIT_PREFIX` |

훅에만 있는 변수가 없습니다. `GIT_INDEX_FILE`은 커밋 계열 훅에만 있어 훅 전체를 덮지 못하고, `GIT_EXEC_PATH`는 git이 실행한 모든 자식(별칭·페이저·필터·자격증명 헬퍼)에 붙는 데다 **git이 문서화한 입력 설정 변수**라 사용자가 직접 설정할 수도 있습니다.

그래서 `runby`는 git 훅을 감지하지 않습니다. **pre-commit은 프레임워크가 자기 마커를 심으므로** 예외이며, 그 경우에도 보고하는 것은 "git 훅 안"이 아니라 "pre-commit이 실행한 훅 안"입니다.

### cron

실행 식별 변수를 설정하지 않습니다. 환경이 빈약하다는 것은 근거가 아니라 추측입니다.

### Yarn

같은 계약을 구현하는 것으로 알려져 있으나 조사 시점에 실측하지 못했고 공식 문서에서 확인하지 못해 v1에서 제외했습니다. 근거 없는 마커를 넣지 않는다는 이 저장소의 기준을 따른 것입니다.

## 알아둘 함정

- **`npm_config_user_agent`는 접두사로 비교해야 합니다.** pnpm과 Bun이 값 뒤쪽에 `npm/?`을 붙이므로, 부분 문자열로 `npm`을 찾으면 셋 다 npm으로 잡힙니다.
- **`MAKEFLAGS`는 마커가 아닙니다.** GNU Make 매뉴얼은 "항상 export된다"고 하지만 플래그 없이 실행하면 **빈 문자열**이고, `runby`는 빈 값을 미설정으로 취급합니다. 그래서 `MAKELEVEL`을 씁니다.
- **`INIT_CWD`는 계열 공통이 아닙니다.** Bun은 설정하지 않습니다.
- **`JOURNAL_STREAM`은 존재만으로 판단하면 안 됩니다.** systemd 문서가 직접 경고합니다 — 서비스가 자식의 표준 스트림을 교체할 수 있으므로, 올바른 확인은 값의 device/inode를 실제 파일 디스크립터와 비교하는 것이고 그건 환경변수가 아니라 시스템 콜의 영역입니다. `runby`는 컨텍스트로만 노출합니다.
- **`npm_lifecycle_script`은 읽지 않습니다.** 스크립트 본문 전체가 들어 있어 인라인 자격증명을 포함할 수 있습니다. `Evidence`에도 `Extra`에도 넣지 않습니다.
- **부재는 부정이 아닙니다.** `PRE_COMMIT`은 pre-commit 2.5.0부터이고, 스크립트가 `env -i`로 환경을 비우면 어떤 마커도 남지 않습니다.

조사 근거는 [`docs/research/runners/`](../../research/runners/)에 있습니다.
