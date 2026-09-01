# 에이전트 조사 규칙과 미채택 후보

이 디렉터리는 `runby` agent 축의 제품별 판정 근거를 모은다. 감지 드라이버는 제품이 **자기가 실행하는 프로세스에** 설정하는 환경변수만 실행 마커로 채택한다. 사용자가 미리 설정하는 API 키·모델·경로·기능 플래그, 터미널이나 workspace 전체의 변수는 agent 실행 주체의 증거로 보지 않는다. 드라이버의 `Evidence`에는 변수 값이 아니라 변수 이름만 남긴다.

## `AGENT`·`AI_AGENT` 표준화 흐름

[agents.md issue #136](https://github.com/agentsmd/agents.md/issues/136)은 AI coding agent가 명령을 실행할 때 `AGENT`에 `[agent-name]`, `1`, `true` 중 하나를 넣자는 **제안**이다. 아직 단일 표준으로 확정된 것은 아니며, 같은 논의가 제품별 변수도 함께 기록한다. 현재 조사에서 공식적으로 확인되는 사례는 다음과 같다.

- Goose는 `AGENT=goose`를 설정하고, Amp도 `AGENT=amp`를 설정한다. 두 제품의 조사 문서에 각각 기록되어 있다.
- Crush는 전용 `CRUSH=1`과 함께 `AGENT=crush`·`AI_AGENT=crush`를 모든 spawn 셸에 설정한다. Crush 드라이버는 전용 마커로 먼저 판정하고, 두 일반 변수는 값이 `crush`일 때만 보조 evidence로 기록한다.
- OpenHands는 terminal·hook subprocess의 정리된 환경에 `AI_AGENT=openhands`를 기본 설정한다. 이처럼 제품이 자기 이름을 값으로 보장하는 경우에만 정확한 값 매칭으로 판정한다.
- Claude Code의 `AI_AGENT=claude-code...`는 제품명을 지목할 때만 evidence로 기록하는 별도 규칙이다. `AI_AGENT`가 있다는 사실만으로 generic agent를 판정하지 않는다.

알 수 없는 값에 대한 generic 폴백 드라이버는 이번에 구현하지 않았다. `runby`의 agent 축은 매치된 모든 드라이버를 `Agents`에 보고한다. generic `AI_AGENT` 드라이버를 추가하면 `AI_AGENT=openhands` 같은 환경에서 OpenHands와 generic 결과가 동시에 생겨 특정 드라이버와 이중 보고되고, `AGENT`·`AI_AGENT`가 함께 있는 제품은 더 복잡해진다. 표준 변수의 소유권·우선순위를 축 수준에서 설계한 뒤에야 안전하게 추가할 수 있다.

## 조사했으나 감지하지 않는 제품

| 제품 | 결론 |
|---|---|
| [Aider](aider.md) | `run_cmd.py`는 환경을 상속할 뿐 marker를 주입하지 않고, `AIDER_*`는 사용자 설정 입력 |
| [Kilo Code](kilo-code.md) | bash tool이 marker를 주입하지 않으며 `ROO_ACTIVE`를 Kilo 전용으로 보장하지 않음 |
| [Continue CLI](continue-cli.md) | `runTerminalCommand.ts`는 부모 환경·색상 변수만 사용하고 Continue marker가 없음 |
| [Factory Droid](factory-droid.md) | 비공개 구현이고 공식 문서에는 `FACTORY_API_KEY` 등 설정·인증 변수만 있음 |
| [Warp Agent](warp-agent.md) | Agent Mode 실행 marker가 없고 Warp 터미널 자체는 terminal 축에서 감지 |
| [Replit Agent](replit-agent.md) | `REPL_ID`·`REPLIT_*`는 workspace/app 런타임 변수로 remote 축 성격 |
| [Jules](jules.md) | 공식 환경 문서에 VM/자식 실행 marker가 없고 VM은 remote 축 성격 |
| [Windsurf Cascade](windsurf-cascade.md) | 비공개 제품이며 공식 marker 문서가 없음 |
| [Devin](devin.md) | 후보 `/opt/.devin`은 env가 아닌 파일 경로라 env 기반 범위 밖 |

[Trae](trae.md)의 `TRAE_AI_SHELL_ID`는 공식 근거가 없어 보류한다. 이 변수와 Devin의 파일 경로는 agents.md 논의에 후보로 나타나지만, 제3자 관찰만으로 제품 계약을 만들지 않는다.
