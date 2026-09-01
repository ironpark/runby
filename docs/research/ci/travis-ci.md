---
title: Travis CI
slug: travis-ci
research_date: 2026-08-31
open_source: true
repository: https://github.com/travis-ci/travis-ci
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 문서(docs.travis-ci.com, config.travis-ci.com)가 기본 주입 변수와 값 형식을 구체적으로 문서화하고 있어 실제 빌드 실행 없이도 감지 규칙을 확정할 수 있음
---

# Travis CI

Travis CI 공식 문서는 모든 빌드에 주입되는 "default/built-in environment variables" 목록을 공개하며, 이 변수들은 사용자의 `.travis.yml` 설정과 무관하게 Travis CI 워커가 job 환경에 항상 설정합니다. 다만 이 변수들은 표준 환경변수이므로 자식 프로세스로 상속되거나 로컬 셸에서 동일한 이름으로 위조될 수 있고, 문서 자체도 이를 "internal" 식별자로만 설명할 뿐 위조 방지를 보장하지 않습니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `TRAVIS` | `true` | 실행 식별 | Travis CI 환경임을 나타내는 전용 마커 | 적합 — Travis CI가 항상 설정하는 제품 전용 상수이며, 가장 신뢰도 높은 존재 마커 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `CI` | `true` | 실행 식별 | 일반적인 CI 환경 표시 | 보조 신호 — Travis CI뿐 아니라 대부분의 CI 플랫폼이 동일한 관례로 설정하는 범용 변수라 단독으로는 Travis CI를 특정하지 못함 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `CONTINUOUS_INTEGRATION` | `true` | 실행 식별 | 레거시 CI 표시 변수 | 보조 신호 — `CI`와 마찬가지로 여러 CI 도구가 공유하는 관례적 변수 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_BUILD_ID` | 내부 ID 문자열 | 실행 식별 | Travis CI가 내부적으로 사용하는 현재 빌드의 식별자 | 적합 — 빌드 단위의 안정적 ID이며 빌드 URL 생성에도 쓰임 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_JOB_ID` | 내부 ID 문자열 | 실행 식별 | Travis CI가 내부적으로 사용하는 현재 job의 식별자 | 적합 — job 단위의 안정적 ID이며 job URL 생성에도 쓰임 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_BUILD_NUMBER` | 정수 문자열 (예: `"4"`) | 실행 식별 | 현재 빌드의 사람이 읽을 수 있는 순번 | 적합 — `TRAVIS_BUILD_ID`의 보조 표시용 빌드 레벨 식별자로 사용 가능 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_JOB_NUMBER` | `빌드번호.job번호` 문자열 (예: `"4.1"`) | 실행 식별 | 현재 job의 사람이 읽을 수 있는 번호 | 적합 — 빌드 번호와 job 순번을 함께 담아 build↔job 관계를 표시하는 job 레벨 식별자 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |

## 상태·컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `TRAVIS_EVENT_TYPE` | `push` \| `pull_request` \| `api` \| `cron` | 상태·컨텍스트 | 빌드를 트리거한 방식 표시 | 보조 신호 — Travis 실행 여부가 아니라 트리거 유형을 알려주는 값이라 `TRAVIS` 존재를 전제로만 의미가 있음 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_REPO_SLUG` | `owner_name/repo_name` 문자열 | 상태·컨텍스트 | 빌드 대상 저장소 식별 | 보조 신호 — 실행 주체 판정이 아닌 컨텍스트 정보 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_BRANCH` | 브랜치명 문자열 | 상태·컨텍스트 | push 빌드(또는 PR이 아닌 빌드)의 브랜치명 | 보조 신호 — 컨텍스트 정보이며 사용자 셸에서도 흔히 재사용되는 이름 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_COMMIT` | 커밋 SHA 문자열 | 상태·컨텍스트 | 테스트 대상 커밋 | 보조 신호 — 컨텍스트 정보 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_PULL_REQUEST` | PR 번호 문자열 또는 `"false"` | 상태·컨텍스트 | PR 빌드 여부와 PR 번호 | 보조 신호 — `TRAVIS_EVENT_TYPE=pull_request`와 함께 트리거 컨텍스트를 보강 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_OS_NAME` | `linux` \| `windows` 등 | 상태·컨텍스트 | 멀티 OS 빌드에서 현재 워커의 OS | 보조 신호 — 실행 주체가 아닌 워커 플랫폼 정보 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_JOB_RESTARTED` | `true` \| `false` | 상태·컨텍스트 | 현재 job이 재시작(restart)된 것인지 표시 | 보조 신호 — Travis CI에는 숫자형 재시도/시도 횟수 카운터가 문서화되어 있지 않고, 이 불리언 플래그만 존재 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |
| `TRAVIS_TEST_RESULT` | `0` \| `1` | 상태·컨텍스트 | `script` 단계 전체 성공 여부 | 부적합 — job 종료 시점에 설정되는 결과값으로 실행 주체 감지와 무관 | [Default Environment Variables](https://docs.travis-ci.com/user/environment-variables/) |

이 문서는 아래 변수들을 **의도적으로 표에서 제외**했습니다.

- `TRAVIS_BUILD_DIR`, `TRAVIS_APP_HOST`, `TRAVIS_DIST`, `TRAVIS_COMPILER`, `TRAVIS_CPU_ARCH`: 워커 환경 세부 설정으로, 실행 주체 판정이나 `runby`가 노출할 핵심 실행 메타데이터에 필요하지 않습니다.
- `TRAVIS_TAG`, `TRAVIS_COMMIT_MESSAGE`, `TRAVIS_COMMIT_RANGE`, `TRAVIS_PULL_REQUEST_BRANCH`, `TRAVIS_PULL_REQUEST_SHA`, `TRAVIS_PULL_REQUEST_SLUG`, `TRAVIS_PULL_REQUEST_IS_DRAFT`: `TRAVIS_BRANCH`/`TRAVIS_COMMIT`/`TRAVIS_PULL_REQUEST`가 제공하는 컨텍스트의 세부 파생값이며, 과도한 사용자 데이터(커밋 메시지 등)를 함께 노출할 수 있어 최소 노출 원칙에 맞지 않습니다.
- `TRAVIS_SECURE_ENV_VARS`, `TRAVIS_JOB_RESTARTED_BY`, `TRAVIS_ALLOW_FAILURE`, `TRAVIS_DEBUG_MODE`, `TRAVIS_SUDO`: 실행 정책·권한 관련 보조 플래그로, 실행 주체 감지에 기여하지 않습니다.
- 언어별 버전 변수(`TRAVIS_NODE_VERSION`, `TRAVIS_GO_VERSION`, `TRAVIS_RUBY_VERSION` 등)와 `TRAVIS_MARIADB_VERSION`, Xcode 관련 변수: 설정된 언어/서비스에 따라 존재 여부가 달라지는 선택적 변수로, 범용 실행 감지 신호가 아닙니다.
- `TZ`, `HAS_JOSH_K_SEAL_OF_APPROVAL`, `USER`, `HOME`, `LANG`, `LC_ALL`, `RAILS_ENV`, `RACK_ENV`, `MERB_ENV`, `JRUBY_OPTS`, `JAVA_HOME`, `DEBIAN_FRONTEND`: Travis 워커가 셸 편의를 위해 설정하는 범용 환경값으로, 이름 자체가 Travis 전용이 아니거나(`USER`, `HOME`, `LANG` 등) 특정 언어 스택 전용이라 실행 감지에 부적합합니다.

## 실행 주체 감지에 관한 결론

`runby`는 `TRAVIS=true`를 Travis CI 실행의 1차 존재 마커로 사용해야 합니다. Travis CI 전용 이름이며 문서에서 모든 빌드에 항상 설정된다고 명시하기 때문입니다. `CI=true`와 `CONTINUOUS_INTEGRATION=true`는 다른 CI 플랫폼도 공유하는 범용 관례이므로 Travis CI 특정에는 쓰지 말고, 다른 CI 감지기와의 충돌 방지에만 참고해야 합니다.

빌드/job 계층 식별에는 다음 우선순위를 권장합니다.

1. **빌드 레벨 안정 ID**: `TRAVIS_BUILD_ID`(내부 ID, URL 생성용)를 우선하고, 사람이 읽을 수 있는 표시용으로 `TRAVIS_BUILD_NUMBER`를 보조로 사용합니다.
2. **job 레벨 안정 ID**: `TRAVIS_JOB_ID`(내부 ID)를 우선하고, `TRAVIS_JOB_NUMBER`(`빌드번호.job번호` 형식)를 보조 표시값으로 사용합니다.
3. **재시도/시도 카운터**: Travis CI 공식 문서에는 숫자형 attempt/retry 카운터가 없습니다. `TRAVIS_JOB_RESTARTED`(불리언)만 문서화되어 있어, job이 재시작되었는지 여부만 판단할 수 있고 몇 번째 시도인지는 알 수 없습니다.
4. **트리거 유형**: `TRAVIS_EVENT_TYPE`이 `push`, `pull_request`, `api`, `cron` 중 하나로 문서화되어 있으며, 이 값이 이 네 가지 외의 값을 가질 수 있다는 공식 언급은 확인하지 못했습니다.

**travis-ci.com과 travis-ci.org, Enterprise 간 차이**: `travis-ci.org`는 2020년 말 `travis-ci.com`으로 통합·폐쇄되었으므로 현재는 `travis-ci.com`만 남아 있습니다. 두 도메인을 구분하는 전용 환경변수는 공식 문서에서 확인되지 않았고, 커뮤니티에서도 이런 변수를 요청하는 이슈([travis-ci/travis-ci#9124](https://github.com/travis-ci/travis-ci/issues/9124))가 있었을 뿐 공식적으로 추가되었다는 근거는 찾지 못했습니다. Travis CI Enterprise(온프레미스) 설치는 동일한 `docs.travis-ci.com` 문서 체계 아래 설명되며, 기본 주입 변수 이름이 다르다는 공식 언급도 확인되지 않았습니다. 따라서 이 문서의 변수는 세 배포 형태 모두에서 동일하게 존재한다고 가정하되, 공식적으로 명시적 확인 문구는 없다는 점을 유보로 남깁니다.

**제품 상태(2026-08-31 기준)**: 공식 상태 페이지([traviscistatus.com](https://www.traviscistatus.com/))는 이 조사 시점에 모든 시스템이 "All Systems Operational"이며 최근 15일간 보고된 장애가 없다고 표시했습니다. `docs.travis-ci.com`이나 Travis CI 공식 채널에서 서비스 종료(EOL)를 공식 발표한 근거는 확인하지 못했습니다. 다만 일부 서드파티 호스팅 사업자(예: WordPress VIP)가 자사 플랫폼에서 Travis CI 지원을 2026년 3월 31일에 종료하고 GitHub Actions로 이전하도록 공지한 사례가 있는데, 이는 해당 사업자의 자체 정책이지 Travis CI 자체의 공식 서비스 종료 발표가 아닙니다. `runby`는 이 문서화된 변수들을 현재 신뢰할 수 있는 신호로 취급하되, Travis CI 채택률이 다른 CI 플랫폼 대비 낮아지는 추세를 감안해 낮은 우선순위의 CI 감지기로 등록하는 편이 합리적입니다.

환경변수는 자식 프로세스로 상속되거나 로컬에서 동일한 이름으로 설정될 수 있으므로, 이 표의 `적합` 판정도 절대적인 신뢰 경계는 아닙니다.

## 공식 문서

- [Environment Variables](https://docs.travis-ci.com/user/environment-variables/)
- [Travis CI Build Config Reference — env](https://config.travis-ci.com/ref/env)
- [API Developer Documentation — env_vars](https://developer.travis-ci.com/resource/env_vars)
- [Travis CI Status](https://www.traviscistatus.com/)
- [공식 저장소 (`travis-ci/travis-ci`)](https://github.com/travis-ci/travis-ci)

## Pull request 감지

`TRAVIS_PULL_REQUEST`는 일반 빌드에서 문자열 `false`, PR 빌드에서 PR 번호를 제공한다고 공식 문서가 명시합니다. 따라서 `runby`는 값이 비어 있지 않고 대소문자 무관하게 `false`가 아니면 `PullRequest=true`로 보고하고, 그 값을 `PullRequestID`로 옮깁니다. 판정에 사용한 변수 이름은 `Evidence`에 기록하며 번호 값은 기록하지 않습니다.
