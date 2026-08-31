# 드라이버 만들기와 배포하기

`runby`가 모르는 제품을 감지하려면 **드라이버**를 만듭니다. 내장 제품과 사내 제품이 완전히 같은 타입을 쓰므로, 특별 취급이나 포크가 필요 없습니다.

드라이버를 쓰는 방법은 두 가지입니다.

| | `Register` | `With*Drivers` | `WithOnlyDrivers` |
|---|---|---|---|
| 범위 | 프로세스 전체 | 그 `Detect` 호출 하나 | 〃 |
| 내장 드라이버 | 같이 실행 (같은 식별자는 교체) | 같이 실행 | **무시** |
| 등록된 드라이버 | — | 같이 실행 | **무시** |
| `IsAgent()`·`Current()`·CLI | ✅ | ❌ | ❌ |
| 쓰는 곳 | 드라이버 모듈의 `init` | 호출부 | 테스트 |

**대표 API(`IsAgent()`, `Current()`)는 옵션을 받지 않습니다.** 그래서 라이브러리로 배포할 드라이버는 `Register`가 유일한 길입니다.

## 1. 드라이버 모듈 만들기

```go
// example.com/runby-acme
package acme

import "github.com/ironpark/runby"

func init() {
	runby.Register(
		runby.AgentDriver{
			Agent:  "acme",
			Kind:   runby.KindOrchestrator,
			Models: runby.ModelsDelegated,
			// 살아 있는 조상 프로세스로 환경 판정을 확증합니다.
			// 무관한 프로세스를 잘못 라벨링할 만큼 일반적인 이름
			// (node, python, java)이면 비워 두십시오.
			Executables: []string{"acme-run"},
			Detect: func(env runby.Env) (runby.Detection, bool) {
				id, ok := runby.Value(env, "ACME_RUN_ID")
				if !ok {
					return runby.Detection{}, false
				}
				return runby.Detection{
					AgentID: id,
					Axis:    runby.Axis{Evidence: runby.PresentNames(env, "ACME_RUN_ID")},
				}, true
			},
		},
	)
}
```

`Register`는 **다섯 축의 드라이버를 한 번에** 받습니다. `Driver` 인터페이스는 닫혀 있어(비공개 메서드) 이 패키지의 다섯 타입만 만족합니다 — 축 자체를 밖에서 추가할 수는 없습니다. 축은 드라이버만이 아니라 `Result`의 필드이자 조상 라벨 기여이자 매치 결합 규칙이기 때문입니다.

## 2. 쓰는 쪽

```go
import (
	"github.com/ironpark/runby"
	_ "example.com/runby-acme" // init에서 Register
)

func main() {
	if runby.IsAgent() {
		log.Printf("run by %s", runby.Current().Chain())
	}
}
```

옵션을 어디에도 넘기지 않았는데 `acme`가 잡힙니다.

## 규칙

### 등록된 드라이버가 내장보다 앞섭니다

같은 식별자의 내장 드라이버가 있으면 **교체합니다**(나란히 실행되지 않습니다). 내장 드라이버가 낡았을 때 이 패키지의 릴리스를 기다리지 않고 고칠 수 있습니다.

### 에이전트 축의 순서는 사다리가 정합니다

패키지 초기화 순서는 드라이버 작성자가 통제할 수 없습니다. 그래서 에이전트 축은 등록 순서나 옵션 전달 순서가 아니라 **`Kind`와 `Models`에서 파생된 `Level`로 정렬**됩니다 — `l3` → `l2` → `l1`.

이게 없으면 등록된 `l1` 하네스가 내장 `l3` 오케스트레이터보다 앞서서, 오케스트레이터가 구동한 런타임이 `Primary()`가 되는 정반대 결과가 납니다.

**`Kind`와 `Models`를 반드시 선언하십시오.** 둘 다 없으면 사다리에 자리가 없어 맨 뒤로 갑니다 — 이 패키지가 그것을 오케스트레이터라고 주장할 근거가 없기 때문입니다.

다른 축의 순서:

- **CI·터미널** — 첫 매치가 이깁니다. 등록된 드라이버가 앞이므로 내장보다 우선합니다. 등록된 드라이버끼리의 순서는 임포트 순서이며 **보장되지 않습니다.** 정확한 순서가 필요하면 `WithCIDrivers`로 직접 넘기십시오.
- **remote·runner** — 매치되는 전부가 보고되므로 순서는 결과 슬라이스의 순서일 뿐입니다.

### 실패는 시끄럽게

`Register`는 두 경우에 **panic**합니다. 조용한 실패는 "요청한 코드와 말없이 어긋난 감지"가 되기 때문입니다.

- 같은 축에 같은 식별자를 두 번 등록
- `Current()`가 이미 캐시를 계산한 뒤에 등록 — `init`에서 부르면 일어나지 않습니다

### 프로세스 전체 상태입니다

빌드 어딘가의 `_` 임포트 하나가 **프로그램 전체**의 감지를 바꿉니다. 부탁한 적 없는 코드까지 포함해서요. 이 패턴에 내재된 성질이고 `database/sql`이 하는 것과 같은 거래입니다.

- **`_` 임포트는 main 패키지에서 하십시오.** 라이브러리가 드라이버를 임포트하면 그 라이브러리에 의존하는 모든 프로그램에 전이적으로, 보이지 않게 강요됩니다.
- **테스트는 `WithOnlyDrivers`를 쓰십시오.** 내장도 등록된 것도 무시하고 준 것만 실행하므로, 다른 곳의 `_` 임포트에 좌우되지 않습니다.

```go
// 내 드라이버만 놓고 테스트 — 내장 드라이버가 같은 픽스처에 걸릴 여지가 없습니다.
result := runby.Detect(
	runby.WithEnviron([]string{"ACME_RUN_ID=r1"}),
	runby.WithOnlyDrivers(acmeDriver),
)
```

## 드라이버 작성 지침

`Detect` 함수는 이 패키지의 내장 드라이버와 같은 규칙을 따라야 합니다.

**`env`를 보관하지 마십시오.** 호출 이후에도 유효하다는 보장이 없습니다.

**값이 아니라 이름만 `Evidence`에 넣으십시오.** 값은 토큰이나 경로일 수 있습니다. `PresentNames(env, names...)`가 설정된 것만 골라 정렬·중복 제거해 줍니다. 본문이 통째로 들어가는 변수(스크립트 소스 등)는 이름조차 넣지 마십시오.

**설정 변수를 근거로 쓰지 마십시오.** API 키나 사용자가 넣는 설정값은 "그 제품이 이 프로세스를 실행했다"의 증거가 아닙니다. 필요한 것은 제품이 **자기가 실행한 자식에게 심는 마커**입니다.

**빈 문자열은 미설정입니다.** `Value`가 그렇게 취급합니다. `MAKEFLAGS`처럼 "항상 export되지만 비어 있을 수 있는" 변수는 마커가 될 수 없습니다.

**부재를 부정으로 쓰지 마십시오.** 마커가 없다는 것이 "그 제품이 아니다"를 뜻하는 경우는 드뭅니다.

**`Detect`가 채워 주는 것은 다시 쓰지 마십시오.** 식별자(`Agent`/`Provider`/`Program`/`Platform`/`Tool`), `Kind`, `Models`, `Level`, 그리고 비어 있는 `Confidence`·`Sandbox.Network`는 자동으로 채워집니다.

파싱 헬퍼는 내장 드라이버가 쓰는 것과 같은 것을 쓰면 됩니다 — [`api.md`의 헬퍼 표](api.md)를 보십시오.

## 조사 문서

이 저장소의 내장 드라이버는 전부 [`docs/research/`](../research/)에 공식 출처를 인용한 조사 문서를 갖고 있고, 테스트가 그 존재를 강제합니다. 사내 드라이버에 같은 것을 요구하지는 않지만, **판정의 근거를 어딘가에 적어 두는 습관**은 권합니다 — 제품이 환경변수 계약을 바꿨을 때 무엇을 다시 확인해야 하는지가 그 문서에 있기 때문입니다.

내장 제품으로 기여하실 생각이라면 조사 문서가 필수입니다. 양식은 [`docs/research/README.md`](../research/README.md)에 있습니다.
