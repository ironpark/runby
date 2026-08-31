---
title: Visual Studio Code
slug: vscode
research_date: 2026-08-31
open_source: true
repository: https://github.com/microsoft/vscode
product_type: terminal_emulator
executes_agents: []
runtime_test_required: true
runtime_test_reason: 공식 소스에서 `TERM_PROGRAM=vscode` 주입 코드는 확인했지만, (1) VS Code 포크(Cursor·Windsurf·Trae·Kiro 등)가 이 리터럴을 그대로 유지하는지, (2) `VSCODE_GIT_ASKPASS_*`가 통합 터미널 환경에 실제로 나타나는지(Git 확장이 터미널에 기여하는 경로), (3) 셸 통합이 꺼진 프로필에서 `VSCODE_INJECTION`이 실제로 부재하는지, (4) Remote-SSH·Dev Containers·WSL 원격 세션의 터미널에서도 같은 변수 집합이 주입되는지는 실제 실행으로 확인하지 않음
---

# Visual Studio Code

VS Code의 통합 터미널은 자식 셸 프로세스에 `TERM_PROGRAM=vscode`를 **무조건** 주입합니다. 공식 소스의 `addTerminalEnvironmentKeys`가 조건 없이 리터럴을 대입하며, 버전은 값이 있을 때만 덧붙습니다.

```ts
export function addTerminalEnvironmentKeys(env: IProcessEnvironment, version: string | undefined, locale: string | undefined, detectLocale: 'auto' | 'off' | 'on'): void {
	env['TERM_PROGRAM'] = 'vscode';
	if (version) {
		env['TERM_PROGRAM_VERSION'] = version;
	}
	if (shouldSetLangEnvVariable(env, detectLocale)) {
		env['LANG'] = getLangEnvVariable(locale);
	}
	env['COLORTERM'] = 'truecolor';
}
```

— [공식 소스: `addTerminalEnvironmentKeys`](https://github.com/microsoft/vscode/blob/c4f707a8a49b978e488fae85d1c62ef581e6a399/src/vs/workbench/contrib/terminal/common/terminalEnvironment.ts#L62-L71)

공식 문서도 셸 통합 스크립트가 `$TERM_PROGRAM`이 `vscode`인지 먼저 확인할 것을 권장하므로, 이 값은 구현 세부사항이 아니라 VS Code가 의도적으로 노출하는 식별 신호입니다 — [Terminal Shell Integration](https://code.visualstudio.com/docs/terminal/shell-integration).

## 이 문서의 핵심: `vscode`는 제품이 아니라 계열입니다

`'vscode'`는 **공유 소스에 박힌 리터럴**입니다. VS Code를 포크한 제품이 이 줄을 고치지 않으면 그 제품의 통합 터미널도 정확히 같은 값을 내보냅니다. 현재 널리 쓰이는 포크만 해도 Cursor, Windsurf, Trae, Kiro, Positron이 있고, 이들은 모두 `microsoft/vscode`를 기반으로 합니다.

따라서 `TERM_PROGRAM=vscode`가 증명하는 것은 **"VS Code 엔진의 통합 터미널"**이지 "Microsoft가 배포한 Visual Studio Code"가 아닙니다. 이는 `runby`가 이미 Konsole에 적용한 판단과 같은 구조입니다 — `KONSOLE_VERSION`은 Konsole 엔진(`konsolepart`가 Dolphin·Kate 등에 임베드된 경우 포함)을 증명할 뿐 Konsole 창을 증명하지 않으므로 `ConfidenceProbable`입니다. 같은 이유로 `runby`는 이 터미널을 `vscode`라는 **계열 이름**으로 보고하고 신뢰도를 `probable`로 둡니다.

포크를 구분할 수 있는 유일한 후보는 경로 값입니다. Git 확장이 주입하는 `VSCODE_GIT_ASKPASS_NODE`는 실행 중인 앱의 Electron 바이너리 경로(`process.execPath`)이므로 포크마다 값이 다릅니다.

```ts
			// GIT_ASKPASS
			GIT_ASKPASS: askpassScript,
			// VSCODE_GIT_ASKPASS
			VSCODE_GIT_ASKPASS_NODE: process.execPath,
			VSCODE_GIT_ASKPASS_EXTRA_ARGS: '',
			VSCODE_GIT_ASKPASS_MAIN: askpassPaths.askpassMain
```

— [공식 소스: `Askpass.getEnv`](https://github.com/microsoft/vscode/blob/92fe1dc89b929e8ef264e1f3b586d8940b3b2ff2/extensions/git/src/askpass.ts#L37-L48)

그러나 `runby`는 이 값을 **감지에 쓰지 않습니다.** 세 가지 이유입니다.

1. 값이 제품명이 아니라 **파일시스템 경로**입니다. 경로에서 제품명을 뽑아내는 것은 문자열 추측이며, 이 저장소가 `HOSTNAME`이 짧은 16진수라는 이유로 컨테이너를 판정하지 말라고 정한 것과 같은 종류의 추측입니다.
2. 내장 Git 확장이 활성화된 경우에만 존재합니다. 확장을 끄거나 `git.enabled: false`로 두면 사라집니다.
3. 경로는 사용자 홈 디렉터리를 포함할 수 있어 **민감할 수 있습니다.** `runby`의 `Evidence`는 이름만 담고 값은 절대 복사하지 않는다는 규칙이 있으므로, 값을 봐야만 성립하는 판정은 이 규칙과 어긋납니다.

포크를 구분해야 한다면 그것은 환경변수가 아니라 프로세스 조상 체인의 실행 파일 이름으로 답해야 할 질문이며, 실행 파일 이름 검증은 아직 하지 않았습니다(아래 참고).

## 터미널 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 터미널 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `TERM_PROGRAM` | 문자열, 정확히 소문자 `vscode` | 실행 식별 | 셸 통합 스크립트와 애플리케이션이 VS Code 터미널을 감지 | 적합(계열 한정) — 모든 통합 터미널 자식에 조건 없이 주입되지만, 값이 공유 소스의 리터럴이라 포크도 같은 값을 냄 | [소스 L63](https://github.com/microsoft/vscode/blob/c4f707a8a49b978e488fae85d1c62ef581e6a399/src/vs/workbench/contrib/terminal/common/terminalEnvironment.ts#L63), [Shell Integration 문서](https://code.visualstudio.com/docs/terminal/shell-integration) |
| `TERM_PROGRAM_VERSION` | 버전 문자열 | 상태·컨텍스트 | VS Code 버전 | 보조 신호 — `version`이 정의된 경우에만 주입되며 단독으로는 다른 터미널과 구분되지 않음 | [소스 L64-L66](https://github.com/microsoft/vscode/blob/c4f707a8a49b978e488fae85d1c62ef581e6a399/src/vs/workbench/contrib/terminal/common/terminalEnvironment.ts#L64-L66) |
| `COLORTERM` | `truecolor` | 설정 | 색 지원 광고 | 부적합 — 다수의 터미널이 같은 값을 설정하는 관례적 변수 | [소스 L70](https://github.com/microsoft/vscode/blob/c4f707a8a49b978e488fae85d1c62ef581e6a399/src/vs/workbench/contrib/terminal/common/terminalEnvironment.ts#L70) |
| `LANG` | 로케일 문자열 | 설정 | 로케일 자동 감지 | 부적합 — 로케일 설정값이며 식별과 무관 | [소스 L67-L69](https://github.com/microsoft/vscode/blob/c4f707a8a49b978e488fae85d1c62ef581e6a399/src/vs/workbench/contrib/terminal/common/terminalEnvironment.ts#L67-L69) |

**VS Code는 창·탭·터미널 인스턴스를 구분하는 세션 식별자를 노출하지 않습니다.** `addTerminalEnvironmentKeys`가 주입하는 전체 목록을 확인했지만 `SESSION_ID`·`WINDOW_ID`·`PANE_ID` 류의 값은 없습니다. 이 점에서 JetBrains(매 터미널마다 UUID)나 iTerm2와 다르며, Ghostty·Warp와 같은 부류입니다.

## 셸 통합 변수 — 조건부라 감지에 쓸 수 없습니다

셸 통합이 주입될 때만 붙는 변수들입니다. 프로필이 셸 통합을 지원하지 않거나 사용자가 껐으면 **아예 존재하지 않으므로**, 부재를 "VS Code가 아님"으로 읽어서는 안 됩니다.

| 환경변수 | 값 | 공식 출처 |
|---|---|---|
| `VSCODE_INJECTION` | `1` | [소스 L89](https://github.com/microsoft/vscode/blob/6e3da3e6e570b20cf2f3f664f0da54aa5f8324c7/src/vs/platform/terminal/node/terminalEnvironment.ts#L89) |
| `VSCODE_NONCE` | 셸 통합 nonce | [소스 L93](https://github.com/microsoft/vscode/blob/6e3da3e6e570b20cf2f3f664f0da54aa5f8324c7/src/vs/platform/terminal/node/terminalEnvironment.ts#L93) |
| `VSCODE_STABLE` | `1`(stable) 또는 `0`(insiders 등) | [소스 L123](https://github.com/microsoft/vscode/blob/6e3da3e6e570b20cf2f3f664f0da54aa5f8324c7/src/vs/platform/terminal/node/terminalEnvironment.ts#L123) |
| `VSCODE_SHELL_LOGIN` | `1` | [소스 L129](https://github.com/microsoft/vscode/blob/6e3da3e6e570b20cf2f3f664f0da54aa5f8324c7/src/vs/platform/terminal/node/terminalEnvironment.ts#L129) |
| `VSCODE_SHELL_ENV_REPORTING` | 보고 대상 변수 이름 목록 | [소스 L101](https://github.com/microsoft/vscode/blob/6e3da3e6e570b20cf2f3f664f0da54aa5f8324c7/src/vs/platform/terminal/node/terminalEnvironment.ts#L101) |
| `VSCODE_A11Y_MODE` | `1`/`0` | [소스 L111](https://github.com/microsoft/vscode/blob/6e3da3e6e570b20cf2f3f664f0da54aa5f8324c7/src/vs/platform/terminal/node/terminalEnvironment.ts#L111) |
| `ZDOTDIR`·`USER_ZDOTDIR` | 임시 디렉터리 경로 (zsh 전용) | 같은 파일 |

`VSCODE_NONCE`는 셸 통합 시퀀스의 위조 방지용 값이므로 **읽거나 기록해서는 안 됩니다.** `runby`는 이 변수들 중 어느 것도 마커로 쓰지 않고, `VSCODE_INJECTION`과 `VSCODE_STABLE`만 존재 여부 수준의 컨텍스트로 노출합니다.

`ZDOTDIR`은 VS Code 전용 이름이 아니라 zsh의 표준 변수이므로 컨텍스트로도 노출하지 않습니다.

## 상속과 잔존

다른 터미널 문서와 같은 네 가지 한계가 그대로 적용되며, VS Code에는 다음이 추가됩니다.

- **원격 개발이 흔합니다.** Remote-SSH·Dev Containers·WSL·Codespaces에서 VS Code 터미널을 열면 이 변수들은 **원격 쪽 셸 프로세스**에 주입됩니다. 즉 `TERM_PROGRAM=vscode`를 보는 프로세스는 VS Code가 도는 머신이 아니라 원격 머신에 있을 수 있습니다. iTerm2의 `LC_TERMINAL`처럼 SSH로 새어 나가는 것이 아니라, VS Code 자체가 원격에 서버를 띄우고 그 위에서 터미널을 여는 구조이기 때문입니다. 이 경우 `runby`의 remote 축(`RemoteCodespaces`·`RemoteDevContainer`·`RemoteWSL`)이 함께 잡히므로 두 축을 같이 읽어야 합니다.
- **tmux**: `TERM_PROGRAM`은 tmux가 pane마다 자기 이름으로 **덮어쓰므로**, tmux 안에서는 VS Code 터미널이 거짓 음성이 됩니다. `VSCODE_*` 변수들은 통과하지만 조건부라 대체 마커가 되지 못합니다.
- **GNU Screen**: `TERM_PROGRAM`을 덮어쓰지 않으므로 잔존한 값이 그대로 통과합니다.

## 실행 파일 이름

`runby`의 `Executables`(살아 있는 조상 프로세스로 환경 판정을 확증하는 목록)는 **비워 둡니다.**

통합 터미널의 실제 부모는 앱 자신이 아니라 pty 호스트 프로세스이며, 그 실행 파일 이름은 플랫폼과 배포 형태에 따라 `Code`, `Code Helper`, `code`, `electron`, `node` 등으로 달라집니다. 이 중 `code`·`node`·`electron`은 무관한 프로세스를 잘못 라벨링할 만큼 일반적인 이름이고, 나머지는 공식 문서로 확인하지 못했습니다. Apple Terminal의 바이너리가 `Terminal`이라 비워 둔 것과 같은 판단입니다.

포크를 조상 체인으로 구분하려면 `Cursor`·`Windsurf` 같은 앱별 실행 파일 이름을 각 제품의 공식 자료로 확인해야 하며, 이는 이 문서의 범위 밖입니다.

## 실행 주체 감지에 관한 결론

`TERM_PROGRAM=vscode`(정확히 소문자)는 공식 소스에서 모든 통합 터미널 자식에 조건 없이 주입되는 유일하고 신뢰도 높은 마커입니다. 다만 그 값은 공유 소스의 리터럴이라 **VS Code 계열 전체**를 가리키며 특정 제품을 가리키지 않으므로, `runby`는 `vscode`를 계열 이름으로 보고하고 `Confidence`를 `probable`로 둡니다.

세션 식별자는 없습니다. 셸 통합 변수(`VSCODE_*`)는 조건부라 마커가 될 수 없고, 포크를 가려낼 수 있는 `VSCODE_GIT_ASKPASS_NODE`는 값이 경로라 이 패키지의 "값을 읽지 않는다" 규칙과 맞지 않아 쓰지 않습니다.

그리고 **터미널을 소유한다는 사실은 에이전트 실행의 증거가 아닙니다.** VS Code는 Copilot을 비롯한 에이전트 확장을 호스팅하지만, 확장이 요청한 명령과 사람이 친 명령을 구분하는 환경변수를 노출하지 않습니다. Zed를 에이전트 축이 아니라 터미널 축으로 보고하기로 한 것과 같은 이유로, VS Code도 터미널 축에만 나타납니다. 포크인 Cursor는 별도로 `CURSOR_AGENT`를 주입하므로 에이전트 축에서 따로 잡히며([`../agents/cursor-agent.md`](../agents/cursor-agent.md)), 그 결과 Cursor Agent가 실행한 명령은 **에이전트 축과 터미널 축에 동시에** 나타날 수 있습니다.

## 공식 문서

- [Terminal Shell Integration](https://code.visualstudio.com/docs/terminal/shell-integration) — 셸 통합 스크립트가 `$TERM_PROGRAM`이 `vscode`인지 확인하도록 권장
- [공식 소스: `addTerminalEnvironmentKeys`](https://github.com/microsoft/vscode/blob/c4f707a8a49b978e488fae85d1c62ef581e6a399/src/vs/workbench/contrib/terminal/common/terminalEnvironment.ts#L62-L71) — `TERM_PROGRAM`·`TERM_PROGRAM_VERSION`·`COLORTERM`·`LANG` 주입
- [공식 소스: 셸 통합 환경 주입](https://github.com/microsoft/vscode/blob/6e3da3e6e570b20cf2f3f664f0da54aa5f8324c7/src/vs/platform/terminal/node/terminalEnvironment.ts#L89-L160) — `VSCODE_INJECTION` 외 조건부 변수
- [공식 소스: `Askpass.getEnv`](https://github.com/microsoft/vscode/blob/92fe1dc89b929e8ef264e1f3b586d8940b3b2ff2/extensions/git/src/askpass.ts#L37-L48) — `VSCODE_GIT_ASKPASS_*`
