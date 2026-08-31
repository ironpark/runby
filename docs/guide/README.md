# 사용자 문서

루트 [`README.md`](../../README.md)는 설치와 5분 안에 쓰기 위한 최소한만 담습니다. 이 디렉터리는 각 축을 실제로 쓸 때 알아야 할 내용을 다룹니다.

| 문서 | 내용 |
|---|---|
| [`cli.md`](cli.md) | `runby` 명령. 셸 스크립트에서 쓰는 법 |
| [`api.md`](api.md) | `Detect`, 옵션 전체, `Result`·`Detection` 구조, 캐시된 진입점, 드라이버 확장 |
| [`agents.md`](agents.md) | 에이전트 축. 감지 대상, `Kind`, `Confidence`, 계층 해석 |
| [`ci.md`](ci.md) | CI 축. 플랫폼별 필드 정규화 규칙 |
| [`terminal.md`](terminal.md) | 터미널 축과 `TTY`. 이 축이 약한 이유 |
| [`remote.md`](remote.md) | 멀티플렉서·SSH·컨테이너 계층과 잔존 위험 |
| [`runner.md`](runner.md) | 실행 주체 축. 스크립트·훅·서비스가 실행했는지, git 훅을 감지할 수 없는 이유 |
| [`process.md`](process.md) | 상위 프로세스 체인과 교차 검증 |

각 판정의 **근거**(어떤 공식 문서·소스에서 확인했는지)는 [`docs/research/`](../research/)에 있습니다.

## 먼저 읽을 것

`runby`를 쓰기 전에 알아야 할 단 하나는 이것입니다.

**환경변수는 프로세스가 시작될 때 상속된 스냅샷입니다.** 감지 성공은 "이 프로세스가 시작될 당시 그 에이전트가 활성 상태였다"는 뜻이지, "지금도 살아 있다"는 뜻이 아닙니다. 장시간 실행되는 프로세스에서 생존 여부가 중요하다면 [`process.md`](process.md)의 교차 검증을 쓰거나 IPC 같은 별도 수단이 필요합니다.

이 한계 때문에 축마다 신뢰도가 다릅니다. 강한 순서로: `Process` > `Layers`·`CI` > `Terminal`.
