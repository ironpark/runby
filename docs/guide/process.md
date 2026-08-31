# Process 축

**환경변수가 아닌 유일한 증거이며, 가장 강한 증거입니다.**

환경변수는 자손이 모두 상속하고, 멀티플렉서가 몇 시간 전 값을 물려주며, 누구나 `export`로 위조할 수 있습니다. 상위 프로세스는 다릅니다 — `export`로 만들 수 없고, 상속되지 않으며, **보인다는 것 자체가 지금 살아 있다는 뜻**입니다. 환경변수가 "이 프로세스가 시작될 때 참이었던 것"만 말할 수 있는 반면, 조상 프로세스는 **지금** 참인 것을 말합니다.

```go
tree := runby.Detect().Process
tree.Supported            // Linux·macOS·Windows에서만 true
tree.Ancestors            // 가까운 부모부터
tree.FindAgent(runby.AgentCodex)
```

```
env chain : paseo>claude-code
  pid=3066    zsh
  pid=11904   claude           -> agent=claude-code
  pid=2540    paseo            -> agent=paseo
corroboration:
  paseo          CONFIRMED by live ancestor pid=2540
  claude-code    CONFIRMED by live ancestor pid=11904
```

## 교차 검증

에이전트가 감지되고 그 실행 파일이 조상으로 살아 있으면 `Detection.AncestorPID`가 채워집니다.

```go
for _, layer := range runby.Current().Layers {
	if layer.AncestorPID != 0 {
		// 환경변수가 이 에이전트를 말했고, 그 에이전트가 지금도 조상으로 살아 있음
	}
}
```

세 축 모두 확증을 받습니다 — `Detection.AncestorPID`, `Terminal.AncestorPID`, `Remote[].AncestorPID`.

### 살아 있는 터미널 조상은 멀티플렉서 강등을 취소합니다

멀티플렉서가 감지되면 터미널 신뢰도를 `probable`로 낮추지만, **그 터미널이 살아 있는 조상으로 잡히면 낮추지 않습니다.**

근거는 멀티플렉서 서버가 데몬화되면서 자신을 시작한 터미널로부터 재부모화된다는 점입니다. tmux 3.7c로 실측한 결과 pane 안의 조상 체인은 `zsh → tmux → pid 1`로 끝나고 **원래 터미널이 나타나지 않습니다.** 따라서 터미널이 조상에 있다는 것은 낡은 pane 뒤가 아니라는 뜻입니다.

### 사내 드라이버도 확증을 받습니다

드라이버에 `Executables`를 채우면 내장 제품과 똑같이 조상 확인을 받습니다.

```go
runby.AgentDriver{
	Agent:       "acme-orchestrator",
	Executables: []string{"acme-run"},
	Detect:      detectAcme,
}
```

라벨은 `Detect` 호출에 설정된 드라이버들로부터 만들어지므로, `Register`나 `WithOnlyDrivers`로 넘긴 드라이버의 실행 파일 이름도 조상 체인 라벨링에 참여합니다.

이름은 소문자 base name에 `.exe`를 뗀 형태로 맞춰야 합니다(`internal/proc`가 그렇게 정규화합니다). Linux `/proc/<pid>/comm`은 15바이트에서 잘리므로, 잘린 이름은 접두사로 대조하되 후보가 둘 이상이면 아무것도 라벨링하지 않습니다.

**`AncestorPID == 0`은 부정이 아닙니다.** 체인은 다른 사용자 소유 프로세스에서 멈추고, 일부 플랫폼에서는 아예 읽을 수 없으며, 에이전트가 조상으로 남지 않는 방식으로 프로세스를 띄울 수도 있습니다. **긍정을 강화하는 데만 쓰고, 부정의 근거로 쓰면 안 됩니다.** 이 규칙을 테스트로 고정해 두었습니다.

## 플랫폼

| 플랫폼 | 방법 |
|---|---|
| Linux | `/proc/<pid>/stat`(ppid), `/proc/<pid>/exe`, `/proc/<pid>/comm` |
| macOS | `sysctl(KERN_PROC_PID)` + `KERN_PROCARGS2` |
| Windows | `CreateToolhelp32Snapshot` 스냅샷 1회 |
| 그 외 | `Supported == false`, 빈 체인 |

macOS에는 `kinfo_proc` 헤더가 없어 ppid 오프셋을 상수로 두어야 합니다. 그래서 **시작 시 `os.Getppid()`와 대조해 검증**하고, 어긋나면 잘못된 필드를 읽는 대신 기능을 끕니다.

`TTY`와 마찬가지로 이 축은 **이 프로세스**를 설명하므로 `WithEnviron` 계열에서는 읽지 않습니다. 비용이 다른 축보다 크므로 `WithoutProcessTree()`로 끌 수 있습니다.
