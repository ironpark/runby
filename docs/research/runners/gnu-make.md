---
title: GNU Make
slug: gnu-make
research_date: 2026-08-31
open_source: true
repository: https://git.savannah.gnu.org/cgit/make.git
product_type: task_runner
executes_agents: []
runtime_test_required: false
runtime_test_reason: 2026-08-31 GNU Make 3.81로 최상위 recipe와 하위 make recipe의 MAKELEVEL·MAKEFLAGS 실값을 직접 관측함. -j 유무에 따른 MAKEFLAGS 차이도 확인함
---

# GNU Make

make가 recipe(레시피)의 셸 명령을 실행할 때 주입하는 변수입니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | 실행 주체 식별 | 공식 출처 |
|---|---|---|---|---|---|
| `MAKELEVEL` | 10진 정수 문자열 | 실행 식별 | 재귀 깊이 | **적합** — recipe에는 항상 1 이상이 전달되며 값이 비는 일이 없음 | [GNU Make 매뉴얼: Variables/Recursion](https://www.gnu.org/software/make/manual/html_node/Variables_002fRecursion.html) |
| `MAKEFLAGS` | 플래그 문자열 | 상태·컨텍스트 | 하위 make로 플래그 전달 | **부적합** — 항상 export되지만 플래그가 없으면 **빈 문자열**이라 마커가 될 수 없음 | 같은 문서 |
| `MFLAGS` | 플래그 문자열 | 상태·컨텍스트 | 하위 호환용 | 부적합 — 같은 이유이며 공식 문서가 하위 호환용으로만 설명 | 같은 문서 |

공식 매뉴얼의 문장입니다.

> As a special feature, the variable `MAKELEVEL` is changed when it is passed down from level to level. This variable's value is a string which is the depth of the level as a decimal number. The value is '0' for the top-level make; '1' for a sub-make, '2' for a sub-sub-make, and so on. **The incrementation happens when make sets up the environment for a recipe.**

> The special variable `MAKEFLAGS` is always exported (unless you unexport it).

## 2026-08-31 실측 (GNU Make 3.81)

```
$ make
top recipe: MAKELEVEL=[1] MAKEFLAGS=[]
sub recipe: MAKELEVEL=[2] MAKEFLAGS=[ --no-print-directory]

$ make -j2
top recipe: MAKELEVEL=[1] MAKEFLAGS=[ --no-print-directory - --jobserver-fds=3,4 -j]
```

두 가지가 확인됩니다.

1. **최상위 make의 recipe도 `MAKELEVEL=1`을 봅니다.** 매뉴얼이 말하는 "최상위는 0"은 make 프로세스 자신의 값이고, recipe를 위한 환경을 구성하는 시점에 증가하므로 `runby`가 관측하는 값은 언제나 1 이상입니다. 따라서 `MAKELEVEL`의 **존재**만으로 판정할 수 있고 값을 파싱할 필요가 없습니다.
2. **`MAKEFLAGS`는 플래그 없이 실행하면 빈 문자열입니다.** `runby`는 빈 값을 미설정으로 취급하므로(`Value`의 규칙), "항상 export된다"는 공식 문장에도 불구하고 마커로 쓸 수 없습니다. 매뉴얼의 서술과 실제 사용 가능성이 갈리는 지점이라 실측이 필요했습니다.

## BSD make

BSD make(`bmake`)도 `MAKELEVEL`을 설정하는 것으로 알려져 있으나 이 조사에서는 확인하지 않았습니다. 사실이라면 같은 마커에 함께 걸리므로 `runby`가 BSD make를 `gnu-make`로 오보고하게 됩니다. 두 구현을 구분하려면 `MAKE_VERSION` 같은 BSD 전용 변수나 조상 프로세스를 추가로 조사해야 하며, 현재는 이 한계를 문서로만 남깁니다.

## 실행 파일 이름

`make`는 이름이 충분히 특정적이므로 `Executables`에 채웁니다. 여러 배포판이 GNU Make를 `gmake`로도 설치하므로 그 이름도 함께 넣습니다.

## 결론

`MAKELEVEL`의 존재가 판정 기준입니다. 값은 재귀 깊이라 컨텍스트로만 노출하고, `MAKEFLAGS`는 빈 문자열일 수 있어 쓰지 않습니다.
