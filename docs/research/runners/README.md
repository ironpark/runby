# 실행 주체(runner) 문서

이 프로세스를 **직접 실행한 도구**를 식별하는 축입니다. 양식과 front matter 필드의 의미는 [상위 문서](../README.md)와 같으며, `product_type`은 모두 `task_runner`입니다.

## 이 축이 존재하는 이유

`runby`가 답한다고 표방하는 질문은 "무엇이 이 프로세스를 실행했는가"입니다. 그런데 기존 네 축은 그 질문의 흔한 답 하나를 통째로 비워 두고 있었습니다 — **스크립트가 실행했다.**

`npm test`로 시작된 프로세스를 생각해 보십시오. `TTY.Interactive`는 참입니다(터미널이 붙어 있으므로). `IsCI()`는 거짓입니다(CI가 아니므로). `IsAgent()`도 거짓입니다. 세 축 모두 "사람이 대화형으로 실행했다"고 답하지만 실제로는 `package.json`의 스크립트가 실행했습니다. 출력 형식이나 색상, 프롬프트 여부를 결정하려는 도구에게 이 차이는 실질적입니다.

systemd 서비스도 마찬가지입니다. 데몬으로 도는 프로세스인지를 환경변수만으로 확실히 아는 방법은 `$INVOCATION_ID` 말고 사실상 없습니다.

## 감지 대상

| 도구 | 마커 | `Kind` | 작업 이름 | 검증 |
|---|---|---|---|---|
| [npm](npm.md) | `npm_config_user_agent`가 `npm/`으로 시작 | `script` | `npm_lifecycle_event` | 실측 (npm 11.13.0) |
| [pnpm](pnpm.md) | `npm_config_user_agent`가 `pnpm/`으로 시작 | `script` | `npm_lifecycle_event` | 실측 (pnpm 11.24.0) |
| [Bun](bun.md) | `npm_config_user_agent`가 `bun/`으로 시작 | `script` | `npm_lifecycle_event` | 실측 (bun 1.3.5) |
| [GNU Make](gnu-make.md) | `MAKELEVEL` | `script` | 없음 | 실측 (GNU Make 3.81) |
| [systemd](systemd.md) | `INVOCATION_ID` | `service` | 없음 | 공식 문서 |
| [pre-commit](pre-commit.md) | `PRE_COMMIT=1` | `hook` | 없음 | 공식 문서 |

## npm 계열은 하나의 사실상 프로토콜입니다

npm, pnpm, Bun(그리고 Yarn)은 서로 다른 제품이지만 **같은 환경변수 계약**을 구현합니다. npm이 정의한 것을 나머지가 호환 목적으로 따라간 결과입니다.

npm 공식 문서는 두 조각을 각각 규정합니다.

- `npm_lifecycle_event`는 "실행 중인 사이클 단계"로 설정됩니다 — [npm scripts](https://docs.npmjs.com/cli/v11/using-npm/scripts)
- `npm_config_` 접두사가 붙은 이름은 설정 파라미터로 해석되며, `user-agent` 설정의 기본 템플릿은 `npm/{npm-version} node/{node-version} {platform} {arch} workspaces/{workspaces} {ci}`입니다 — [npm config](https://docs.npmjs.com/cli/v11/using-npm/config)

즉 `npm_config_user_agent`의 **첫 토큰이 도구 이름**입니다. 2026-08-31 실측 결과입니다.

```
npm    npm_config_user_agent=npm/11.13.0 node/v26.1.0 darwin arm64 workspaces/false
pnpm   npm_config_user_agent=pnpm/11.24.0 npm/? node/v26.1.0 darwin arm64
bun    npm_config_user_agent=bun/1.3.5 npm/? node/v24.3.0 darwin arm64
```

pnpm과 Bun이 `npm/?`을 뒤에 붙이는 것에 주의하십시오. **부분 문자열 검색으로 `npm`을 찾으면 세 도구 모두에 걸립니다.** 반드시 `<이름>/` 형태의 **접두사**로 비교해야 합니다.

`npm_lifecycle_event`(스크립트 이름)는 세 도구 모두 동일하게 설정하지만 차이도 있습니다.

| | `npm_lifecycle_event` | `npm_config_user_agent` | `INIT_CWD` | `npm_execpath` |
|---|---|---|---|---|
| npm | ✅ | ✅ | ✅ | ✅ |
| pnpm | ✅ | ✅ | ✅ | ✅ |
| Bun | ✅ | ✅ | **❌ 설정 안 함** | ✅ |

`INIT_CWD`는 Bun에 없으므로 계열 공통 신호로 쓸 수 없습니다.

### `npm_lifecycle_script`은 읽지 않습니다

세 도구 모두 `npm_lifecycle_script`에 **스크립트 본문 전체**를 넣습니다. 이 값은 임의의 셸 명령문이므로 인라인 토큰이나 자격증명을 포함할 수 있습니다. `runby`는 이 변수를 근거 목록에도 `Extra`에도 넣지 않습니다.

이는 `Evidence`가 이름만 담는다는 기존 규칙보다 한 단계 강한 조치입니다. 이름만 노출하는 것은 안전하지만, 이 변수는 그 이름이 존재한다는 사실만으로도 "여기에 스크립트 본문이 있다"를 알려 주므로 굳이 가리킬 이유가 없습니다.

### Yarn은 아직 넣지 않았습니다

Yarn도 같은 계약을 구현하는 것으로 알려져 있으나, **이 조사 시점에 로컬에 설치되어 있지 않아 실측하지 못했고** Yarn 공식 문서에서 `npm_config_user_agent`를 자신이 설정한다고 명시한 문장을 확인하지 못했습니다. 확인되지 않은 마커를 넣지 않는다는 이 저장소의 기준에 따라 v1에서 제외합니다.

추가하려면 `yarn run`으로 `npm_config_user_agent`의 첫 토큰이 `yarn/`인지 실측하고, Yarn 1.x(Classic)와 Berry가 같은지 확인해야 합니다.

## 감지할 수 없는 것들

### git 훅 — 환경변수로 **구분 불가능**합니다

가장 원했던 항목이지만 실측 결과 불가능합니다. git 2.55.0에서 훅별 환경을 직접 관측했습니다.

| 실행 맥락 | 주입되는 `GIT_*` |
|---|---|
| `pre-commit` 훅 | `GIT_AUTHOR_DATE`, `GIT_AUTHOR_EMAIL`, `GIT_AUTHOR_NAME`, `GIT_CONFIG_PARAMETERS`, `GIT_EDITOR`, `GIT_EXEC_PATH`, `GIT_INDEX_FILE`, `GIT_PREFIX` |
| `prepare-commit-msg` 훅 | 위와 동일 |
| `post-commit` 훅 | 위와 동일 |
| `post-checkout` 훅 | `GIT_EDITOR`, `GIT_EXEC_PATH`, `GIT_PREFIX` |
| **git 별칭 (훅이 아님)** | `GIT_CONFIG_PARAMETERS`, `GIT_EDITOR`, `GIT_EXEC_PATH`, `GIT_PREFIX` |

**`post-checkout` 훅과 훅이 아닌 git 자식 프로세스는 환경이 사실상 같습니다.** 훅에만 존재하는 변수가 없으므로 "나는 git 훅 안에 있다"를 환경변수로 판정할 수 없습니다.

`GIT_INDEX_FILE`은 커밋 계열 훅을 별칭과 구분해 주지만 훅 전체를 덮지 못하고(`post-checkout`에 없음), git이 훅 외의 상황에서도 설정합니다. `GIT_EXEC_PATH`는 모든 경우에 존재하지만 git이 실행한 **모든** 자식(훅·별칭·외부 명령·페이저·필터·자격증명 헬퍼)에 붙으므로 훅을 뜻하지 않습니다. 게다가 `GIT_EXEC_PATH`는 git이 문서화한 **입력 설정 변수**이기도 해서 사용자가 직접 설정할 수 있고, 이는 "설정 변수는 실행 근거가 아니다"라는 이 패키지의 원칙과 정면으로 충돌합니다.

`GIT_PREFIX`는 저장소 루트에서 실행하면 **빈 문자열**이라 마커가 될 수 없습니다(`runby`는 빈 값을 미설정으로 취급합니다).

따라서 `runby`는 git 훅을 감지하지 않습니다. 훅 안에서 도는지 알아야 한다면 그것은 환경변수가 아니라 프로세스 조상 체인이나 실행 경로(`$0`)로 답할 질문입니다. 다만 **pre-commit 프레임워크는 자기 마커를 두므로 감지 대상입니다** — 훅 일반이 아니라 그 프레임워크를 식별하는 것입니다.

### `MAKEFLAGS`는 마커가 될 수 없습니다

GNU Make 공식 문서는 `MAKEFLAGS`가 "항상 export된다"고 명시하지만, 실측 결과 플래그 없이 실행하면 **빈 문자열**입니다.

```
top recipe: MAKELEVEL=[1] MAKEFLAGS=[]
```

`runby`는 빈 값을 미설정으로 취급하므로 `MAKEFLAGS`는 평범한 `make` 실행에서 감지에 쓸 수 없습니다. `MAKELEVEL`을 마커로 쓰는 이유입니다.

### cron

cron은 실행 식별 변수를 설정하지 않습니다. 오히려 환경을 거의 비운 채로 넘기는 것이 특징이며, "환경이 빈약하다"는 것은 근거가 아니라 추측입니다. `runby`는 cron을 감지하지 않습니다.

## 상속과 잔존

다른 축과 같은 한계가 적용되지만, 이 축에는 **중첩이 정상**이라는 고유한 성질이 있습니다.

- **여러 계층이 동시에 참일 수 있습니다.** `pre-commit`이 훅에서 `npm` 스크립트를 부르고, 그 스크립트가 `make`를 부르는 것은 흔한 구성입니다. 그래서 `Result.Runner`는 remote 축처럼 **슬라이스**이고, 순서는 중첩 순서가 아니라 감지 순서입니다.
- **부모가 끝나도 값이 남습니다.** `npm run dev &`로 띄운 백그라운드 프로세스는 npm이 끝난 뒤에도 `npm_lifecycle_event`를 계속 들고 있습니다. 다른 축과 같은 스냅샷 한계입니다.
- **`MAKELEVEL`은 깊이를 셉니다.** 값이 `1`이면 최상위 make의 recipe, `2` 이상이면 하위 make입니다. 실측으로 확인했듯 최상위 make **자신**은 `0`이지만 recipe에는 `1`이 전달되므로, `runby`가 보는 값은 언제나 1 이상입니다.
- **CI 축과 독립입니다.** CI 잡 안에서 `npm test`를 돌리면 두 축이 함께 잡힙니다. CI는 "어디서 도는가", runner는 "무엇이 직접 실행했는가"로 서로 다른 질문입니다.

## 실행 파일 이름

`Executables`(살아 있는 조상으로 확증하는 목록)는 도구별로 판단이 갈립니다.

- **`make`·`systemd`** — 이름이 충분히 특정적이므로 채웁니다.
- **npm 계열** — 실제 프로세스는 `node`이거나(npm) 도구 자신입니다(pnpm, bun). `node`는 무관한 프로세스를 잘못 라벨링할 만큼 일반적이라 npm은 비워 두고, pnpm과 bun은 자기 이름으로 돕니다.
- **pre-commit** — Python 진입점이라 `python`으로 보일 수 있어 비워 둡니다.

각 문서에 근거를 남겼습니다.
