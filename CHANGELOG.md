# Changelog

이 프로젝트는 [유의적 버전](https://semver.org/lang/ko/)을 따르며, 형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.1.0/)를 따릅니다.

## [v0.2.0] - 2026-09-02

### Added

- 에이전트 감지 6종 추가: pi, Charm Crush, Qwen Code, Roo Code, OpenHands, DeepSeek Harness.
  각 감지 규칙의 공식 근거는 [`docs/research/agents/`](docs/research/agents/)에 있습니다.
- 조사 후 제외한 에이전트 후보(Aider, Kilo Code, Continue CLI, Factory Droid,
  Warp Agent, Replit Agent, Jules, Windsurf Cascade, Devin)와 보류(Trae),
  `AGENT`/`AI_AGENT` 표준화 흐름을 연구 문서로 기록.

### Notes

- Roo Code는 Cline과 같은 터미널 마커 방식이라 `probable`로 보고됩니다.
- pi·DeepSeek Harness·Qwen Code는 세션 식별자를 `SessionID`로, 부가 컨텍스트를
  `Extra`로 노출합니다.

## [v0.1.0] - 2026-09-02

첫 릴리스.

- 다섯 축 감지: 에이전트(14종) · CI(37개 제공자, PR 감지 포함) · 터미널 ·
  원격 환경 · 실행 도구, 그리고 TTY·프로세스 조상 체인.
- 드라이버 기반 확장(`Register`, `WithDrivers`, `WithOnlyDrivers`)과
  `EnvReader` 기반 evidence 자동 기록.
- 멀티플렉서(tmux 등) 감지 시 환경 파생 계층의 신뢰도 강등, `Unattended()`의
  probable 계층 제외 규칙.
- `runby` CLI: `is <축> [제품]`, `is unattended`, `chain`, `-json`, `-v`.
- 외부 의존성 없음, MIT.

[v0.2.0]: https://github.com/ironpark/runby/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/ironpark/runby/releases/tag/v0.1.0
