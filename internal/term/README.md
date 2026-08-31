# internal/term

`runby`가 표준 라이브러리만으로 동작하도록, [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term)에서 `IsTerminal` 검사 하나만 옮겨온 패키지입니다.

원본은 The Go Authors의 BSD-3-Clause 라이선스이며 전문은 [`LICENSE`](LICENSE)에 있습니다. 원본이 `golang.org/x/sys`의 `unix.IoctlGetTermios`를 쓰는 부분을 표준 `syscall` 패키지 호출로 다시 작성했고, `MakeRaw`·`Restore`·`GetSize`·`ReadPassword` 등 나머지 API는 옮기지 않았습니다.

## 플랫폼별 구현

| 파일 | 대상 | 방법 |
|---|---|---|
| `term_unix.go` + `term_unix_bsd.go` | darwin, dragonfly, freebsd, netbsd, openbsd | `ioctl(TIOCGETA)` |
| `term_unix.go` + `term_unix_linux.go` | linux | `ioctl(TCGETS)` |
| `term_windows.go` | windows | `GetConsoleMode` |
| `term_plan9.go` | plan9 | `Fd2path`로 `/dev/cons` 확인 |
| `term_unsupported.go` | 그 외 | 항상 `false` |

## 지원하지 않는 플랫폼

AIX·Solaris·z/OS에서는 원본이 `golang.org/x/sys`를 통해 답할 수 있지만, 표준 `syscall` 패키지는 이 플랫폼들에 `TCGETS`와 `SYS_IOCTL`을 노출하지 않습니다(교차 컴파일로 확인). 의존성을 추가하지 않기로 한 결정의 대가로 이 플랫폼들에서는 `TTY.Attached`와 `TTY.Interactive`가 항상 `false`입니다. 환경변수만 읽는 나머지 축은 영향을 받지 않습니다.

WebAssembly(js, wasip1)에는 터미널 개념이 없어 원본도 같은 답을 냅니다.

## 왜 `os.ModeCharDevice`로 대체할 수 없는가

`/dev/null`도 문자 장치입니다. 모드 검사로 바꾸면 파이프로 실행한 모든 경우가 대화형으로 보고됩니다. 커널에 터미널 속성을 물어보는 것만이 둘을 구분합니다. `term_test.go`가 이 사실을 고정합니다.
