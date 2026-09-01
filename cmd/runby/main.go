// Command runby reports what launched the current process.
//
// The package this wraps is a library first; this command exists so a shell
// script can ask the same questions without a Go program, and so a bug report
// can carry one paste instead of a description.
//
// It never prints an environment variable's value: every axis reports the
// variable NAMES it matched on, because a value may hold a token. The -json
// mode is still not safe to paste blindly, because the Result itself carries
// session identifiers and executable paths that the text report omits.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/ironpark/runby"
)

const usage = `runby — 이 프로세스를 무엇이 실행했는지 보고합니다.

  runby [-json] [-v]     사람이 읽는 요약, 또는 Result 전체 JSON
  runby is <축> [제품]   종료 코드로만 답합니다
  runby chain            "paseo>codex" 한 줄. 감지 실패 시 "unknown"
  runby help             이 도움말 (runby -h와 같습니다)

축:
  agent ci terminal remote runner   제품 이름을 덧붙여 좁힐 수 있습니다
  tty unattended                    제품 차원이 없습니다

  runby is agent          에이전트가 실행했는가
  runby is agent codex    그게 codex인가
  runby is ci github-actions
  runby is remote tmux
  runby is unattended     아무도 출력을 보고 있지 않은가 — 스피너·프롬프트를
                          끌지 정하는 질문으로, 라이브러리의 Unattended()와
                          같은 규칙(probable 에이전트는 제외)으로 답합니다

  제품 이름은 -json에 나오는 슬러그와 같습니다. 오타는 거짓이 아니라
  사용법 오류(2)로 답하므로, 스크립트가 잘못된 분기를 타지 않습니다.

플래그:
  -json   Result 전체를 JSON으로 출력합니다. 필드는 라이브러리와 동일합니다.
  -v      각 축이 근거로 삼은 환경변수 이름을 함께 출력합니다.

종료 코드:
  0  정상. "is"에서는 참
  1  "is"에서 거짓, 또는 내부 오류
  2  사용법 오류 (알 수 없는 축·제품 포함)

환경변수 값은 어떤 모드에서도 출력하지 않습니다. -v는 변수 이름만 보여줍니다.
다만 -json에는 제품이 광고한 식별자와 실행 파일 경로가 값으로 들어가므로,
붙여넣어 공유할 때는 -v 쪽이 안전합니다.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is separated from main so the tests can drive it with fixed arguments
// and capture its output.
func run(args []string, stdout, stderr io.Writer) int {
	// The subcommands take no flags, so they are dispatched before flag
	// parsing rather than after it.
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "help":
			fmt.Fprint(stdout, usage)
			return 0
		case "is":
			return runIs(args[1:], stderr)
		case "chain":
			if len(args) > 1 {
				fmt.Fprintf(stderr, "runby: chain은 인자를 받지 않습니다\n")
				return 2
			}
			fmt.Fprintln(stdout, runby.Current().Chain())
			return 0
		default:
			fmt.Fprintf(stderr, "runby: 알 수 없는 명령 %q\n\n%s", args[0], usage)
			return 2
		}
	}

	flags := flag.NewFlagSet("runby", flag.ContinueOnError)
	flags.SetOutput(stderr)
	// Parse prints the offending flag itself; the usage is printed below on
	// the stream the outcome calls for — stdout for -h, stderr for a mistake.
	flags.Usage = func() {}
	asJSON := flags.Bool("json", false, "Result 전체를 JSON으로 출력")
	verbose := flags.Bool("v", false, "근거가 된 환경변수 이름도 출력")
	if err := flags.Parse(args); err != nil {
		// Asking for help is a documented invocation, not a mistake.
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage)
			return 0
		}
		fmt.Fprintf(stderr, "\n%s", usage)
		return 2
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "runby: 예상하지 못한 인자 %q\n\n%s", flags.Arg(0), usage)
		return 2
	}

	result := runby.Current()
	if *asJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			fmt.Fprintf(stderr, "runby: %v\n", err)
			return 1
		}
		return 0
	}

	report(stdout, result, *verbose)
	return 0
}

// axis is one question the "is" subcommand can ask.
//
// any answers "did this axis detect anything", and named answers "did it detect
// this one product". Both come from the library rather than from a rule of this
// command's own, which is what TestIsMatchesTheLibrary and
// TestIsNamedProductMatchesTheLibrary pin.
//
// products lists the names named accepts. It exists so that a typo is a usage
// error rather than a silent false: a script that asks for "codexx" and is told
// "no" would take the wrong branch forever. The list is the built-in one, which
// is complete for this binary because it registers no drivers of its own.
type axis struct {
	any      func(runby.Result) bool
	named    func(runby.Result, string) bool
	products []string
}

// slugs widens an axis's identity list into the plain strings the command
// compares against, since each identity is already its stable slug.
func slugs[T ~string](values []T) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

// axes maps the "is" subcommand's arguments to the questions each asks. They
// are the same questions the library's Result answers.
//
// The tty axis has no named form: it reports whether the standard streams can
// carry a prompt, which is not a product anyone can name.
var axes = map[string]axis{
	"agent": {
		any: func(r runby.Result) bool { return r.IsAgent() },
		named: func(r runby.Result, name string) bool {
			_, ok := r.Agent(runby.AgentName(name))
			return ok
		},
		products: slugs(runby.AgentNames()),
	},
	"ci": {
		any: func(r runby.Result) bool { return r.IsCI() },
		named: func(r runby.Result, name string) bool {
			return r.CI.Provider == runby.CIProvider(name)
		},
		products: slugs(runby.CIProviders()),
	},
	"terminal": {
		any: func(r runby.Result) bool { return r.HasTerminal() },
		named: func(r runby.Result, name string) bool {
			return r.Terminal.Program == runby.TerminalProgram(name)
		},
		products: slugs(runby.TerminalPrograms()),
	},
	"remote": {
		any: func(r runby.Result) bool { return r.IsRemote() },
		named: func(r runby.Result, name string) bool {
			_, ok := r.Remote(runby.RemotePlatform(name))
			return ok
		},
		products: slugs(runby.RemotePlatforms()),
	},
	"runner": {
		any: func(r runby.Result) bool { return r.HasRunner() },
		named: func(r runby.Result, name string) bool {
			_, ok := r.Runner(runby.RunnerTool(name))
			return ok
		},
		products: slugs(runby.RunnerTools()),
	},
	"tty": {
		any: func(r runby.Result) bool { return r.TTY.Interactive },
	},
	// unattended is not an axis of Result but the one cross-axis question the
	// library answers; exposing it here keeps shell scripts from restating the
	// rule — and drifting from it — with is agent/is ci/is tty combinations.
	"unattended": {
		any: func(r runby.Result) bool { return r.Unattended() },
	},
}

func runIs(args []string, stderr io.Writer) int {
	if len(args) == 0 || len(args) > 2 {
		fmt.Fprintf(stderr, "runby: is는 축 하나와 선택적인 제품 하나를 받습니다: is <축> [제품]\n축: %s\n", axisNames())
		return 2
	}
	ask, ok := axes[args[0]]
	if !ok {
		fmt.Fprintf(stderr, "runby: 알 수 없는 축 %q. 가능한 값: %s\n", args[0], axisNames())
		return 2
	}

	if len(args) == 1 {
		return exitCode(ask.any(runby.Current()))
	}

	product := args[1]
	if ask.named == nil {
		fmt.Fprintf(stderr, "runby: %s 축은 제품 이름을 받지 않습니다\n", args[0])
		return 2
	}
	// An unknown name is refused rather than answered, so a typo cannot read as
	// a confident "no".
	if !slices.Contains(ask.products, product) {
		fmt.Fprintf(stderr, "runby: %s 축에 알 수 없는 제품 %q. 가능한 값: %s\n",
			args[0], product, strings.Join(ask.products, " "))
		return 2
	}
	return exitCode(ask.named(runby.Current(), product))
}

// exitCode is the "is" contract in one place: true is 0, false is 1.
func exitCode(answer bool) int {
	if answer {
		return 0
	}
	return 1
}

func axisNames() string {
	names := make([]string, 0, len(axes))
	for name := range axes {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, " ")
}

func report(w io.Writer, result runby.Result, verbose bool) {
	line := func(label, value string, evidence []string) {
		fmt.Fprintf(w, "%-9s %s\n", label, value)
		if verbose && len(evidence) > 0 {
			fmt.Fprintf(w, "%-9s   ← %s\n", "", strings.Join(evidence, " "))
		}
	}

	// Agent.
	if !result.IsAgent() {
		line("agent", "감지되지 않음", nil)
	} else {
		line("agent", result.Chain(), nil)
		for _, layer := range result.Agents {
			live := ""
			if layer.AncestorPID != 0 {
				live = fmt.Sprintf("  살아 있는 조상 pid=%d", layer.AncestorPID)
			}
			fmt.Fprintf(w, "%-9s   %-14s %-13s %-13s %s%s\n", "",
				layer.Name, layer.Kind, layer.Models, layer.Confidence, live)
			if verbose && len(layer.Evidence) > 0 {
				fmt.Fprintf(w, "%-9s     ← %s\n", "", strings.Join(layer.Evidence, " "))
			}
		}
	}

	// CI.
	if !result.IsCI() {
		line("ci", "-", nil)
	} else {
		value := string(result.CI.Provider)
		if result.CI.JobName != "" {
			value += "  job=" + result.CI.JobName
		}
		if result.CI.Attempt > 1 {
			value += fmt.Sprintf("  attempt=%d", result.CI.Attempt)
		}
		line("ci", value, result.CI.Evidence)
	}

	// Terminal.
	if !result.HasTerminal() {
		line("terminal", "-", nil)
	} else {
		value := fmt.Sprintf("%s (%s)", result.Terminal.Program, result.Terminal.Confidence)
		if result.Terminal.Version != "" {
			value += "  " + result.Terminal.Version
		}
		line("terminal", value, result.Terminal.Evidence)
	}

	// Remote. The order is detection order, not nesting order, so the report
	// deliberately does not draw it as a stack.
	if !result.IsRemote() {
		line("remote", "-", nil)
	} else {
		parts := make([]string, 0, len(result.Remotes))
		var evidence []string
		for _, layer := range result.Remotes {
			parts = append(parts, fmt.Sprintf("%s (%s)", layer.Platform, layer.Kind))
			evidence = append(evidence, layer.Evidence...)
		}
		line("remote", strings.Join(parts, ", "), evidence)
	}

	// Runner. Nesting is normal here, so every layer is listed.
	if !result.HasRunner() {
		line("runner", "-", nil)
	} else {
		parts := make([]string, 0, len(result.Runners))
		var evidence []string
		for _, r := range result.Runners {
			part := fmt.Sprintf("%s (%s)", r.Tool, r.Kind)
			if r.Task != "" {
				part += " " + r.Task
			}
			parts = append(parts, part)
			evidence = append(evidence, r.Evidence...)
		}
		line("runner", strings.Join(parts, ", "), evidence)
	}

	// TTY. This is a syscall, not a variable, so it has no evidence to show.
	line("tty", ttyText(result.TTY), nil)

	// Process.
	line("process", processText(result.Process), nil)
	if verbose && len(result.Process.Ancestors) > 0 {
		for _, p := range result.Process.Ancestors {
			// The name is unreadable for processes owned by another user,
			// which is routine once the chain reaches init.
			name := p.Name
			if name == "" {
				name = "(읽을 수 없음)"
			}
			label := ""
			switch {
			case p.Agent != "":
				label = "  agent=" + string(p.Agent)
			case p.Terminal != "":
				label = "  terminal=" + string(p.Terminal)
			case p.Remote != "":
				label = "  remote=" + string(p.Remote)
			case p.Runner != "":
				label = "  runner=" + string(p.Runner)
			}
			fmt.Fprintf(w, "%-9s   pid=%-8d %s%s\n", "", p.PID, name, label)
		}
	}

	if multiplexer, ok := result.Multiplexer(); ok {
		fmt.Fprintf(w, "\n주의: %s 안에서 실행 중입니다. 멀티플렉서는 이미 열린 pane의 환경을\n"+
			"갱신하지 않으므로 환경에서 파생된 축(터미널·에이전트·러너)의 값이 낡았을 수 있습니다.\n", multiplexer.Platform)
	}
}

func ttyText(tty runby.TTY) string {
	if !tty.Inspected {
		return "검사하지 않음"
	}
	switch {
	case tty.Interactive:
		return "대화형 (stdin과 출력이 터미널)"
	case tty.Attached:
		return "일부만 터미널에 연결됨"
	default:
		return "터미널 아님"
	}
}

func processText(tree runby.ProcessTree) string {
	switch {
	case !tree.Inspected:
		return "검사하지 않음"
	case !tree.Supported:
		return "이 플랫폼에서는 읽을 수 없음"
	default:
		return fmt.Sprintf("조상 %d개", len(tree.Ancestors))
	}
}
