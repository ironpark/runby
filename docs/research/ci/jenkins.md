---
title: Jenkins
slug: jenkins
research_date: 2026-08-31
open_source: true
repository: https://github.com/jenkinsci/jenkins
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: 공식 `jenkinsci/jenkins` 저장소의 `jenkins.model.CoreEnvironmentContributor`에서 `CI`, `JENKINS_URL`, `JENKINS_HOME`, `HUDSON_URL`, `HUDSON_HOME`, `EXECUTOR_NUMBER`, `NODE_NAME`, `BUILD_URL`, `JOB_URL`을 주입하는 정확한 조건 분기를 확인했고, `hudson.model.Run`·`hudson.model.Job`·`hudson.model.AbstractBuild`에서 `BUILD_NUMBER`, `BUILD_ID`, `BUILD_TAG`, `JOB_NAME`, `WORKSPACE` 주입 지점을 확인함. 소스 자체가 조건을 명시하므로 별도 실행 검증 없이 문서화할 수 있음
---

# Jenkins

Jenkins는 빌드마다 `CI`, `BUILD_NUMBER`, `BUILD_ID`, `JOB_NAME` 같은 핵심 환경변수를 코어의 `EnvironmentContributor` 확장점(`jenkins.model.CoreEnvironmentContributor`)을 통해 무조건적으로 주입한다고 소스에서 보장합니다. 다만 이 변수들 중 다수는 이름 자체가 매우 일반적이어서(`BUILD_ID`, `BUILD_NUMBER`, `JOB_NAME`, `NODE_NAME` 등) 다른 빌드 도구나 사용자 스크립트가 같은 이름을 재사용할 수 있습니다. 반면 Jenkins 전용 마커(`CI=true`처럼 값이 `"true"`로 고정된 범용 신호를 제외하면)는 존재하지 않으며, `JENKINS_URL`조차 관리자가 시스템 설정에서 Jenkins URL을 지정하지 않으면 주입되지 않습니다. 이 문서는 공식 문서와 공식 소스 코드에서 확인된 사실만 다루며, 확인하지 못한 부분은 명시적으로 "미확인"으로 표시합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `CI` | 고정 문자열 `"true"` | 실행 식별 | 연속 통합 실행 환경임을 표시 (Jenkins 2.286+) | 보조 신호 — Jenkins 코어가 모든 Job 실행에 무조건 주입하지만, TravisCI·CircleCI·GitHub Actions 등 다른 CI 도구도 동일한 이름과 값을 사용하는 범용 관례라 Jenkins 고유 신호가 아님 | [Jenkins 2.286 변경 로그](https://www.jenkins.io/changelog/2.286/), [공식 소스: `CoreEnvironmentContributor.java#L43`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L43), [JENKINS-36707](https://issues.jenkins.io/browse/JENKINS-36707) |
| `JENKINS_URL` | 전체 URL 문자열, 예: `https://example.com:port/jenkins/` | 실행 식별 | 시스템 설정에 등록된 Jenkins 컨트롤러의 전체 URL | 적합 — `jenkins`라는 문자열이 이름에 박혀 있는 가장 구체적인 단일 신호. 단, 관리자가 시스템 설정에서 Jenkins URL을 지정하지 않은 인스턴스에서는 아예 주입되지 않으므로 부재가 "Jenkins 아님"을 의미하지 않음 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `CoreEnvironmentContributor.java#L46-L48`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L46-L48) |
| `BUILD_ID` | 문자열. 현재는 `BUILD_NUMBER`와 동일한 값 | 실행 식별 | 빌드 식별자 | 보조 신호 — 이름이 매우 일반적이며, `JENKINS_URL` 없이 단독으로는 다른 도구와 구분되지 않음. `JENKINS_URL`과 결합하면 강한 신호가 됨 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `Run.java#L2431`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/Run.java#L2431) |
| `BUILD_NUMBER` | 정수 문자열, 예: `"153"` | 실행 식별 | 현재 빌드 번호 | 보조 신호 — `BUILD_ID`와 동일한 이유로 이름이 일반적임. `JENKINS_URL`과 결합해 사용 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `Run.java#L2430`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/Run.java#L2430) |
| `BUILD_TAG` | 문자열, `jenkins-${JOB_NAME}-${BUILD_NUMBER}` (`/`는 `-`로 치환) | 실행 식별 | 산출물 파일명 등에 넣기 좋은 고유 식별 문자열 | 보조 신호 — 값 자체에 `jenkins-` 접두어가 박혀 있어 교차 검증에 유용하지만, 이를 신뢰하려면 변수 존재 확인이 아닌 값 패턴 파싱이 필요해 `JENKINS_URL` 조합보다 신뢰도가 낮음 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `Run.java#L2432`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/Run.java#L2432) |

### BUILD_ID의 타임스탬프 이력

`BUILD_ID`는 과거에는 빌드 시작 타임스탬프를 담는 변수였으나, 공식 문서는 "Jenkins 1.597 이상에서 생성된 빌드에서는 `BUILD_NUMBER`와 동일하다"고 명시합니다. Jenkins 1.597은 2014년에 릴리스된 버전으로, 현재 유지보수되는 모든 Jenkins 버전에서는 이미 이 동작이 적용되어 있습니다. 즉 `runby`는 `BUILD_ID`를 `BUILD_NUMBER`와 별개의 안정적 식별자로 취급할 수 없고, 사실상 `BUILD_NUMBER`의 문자열 복제본으로만 다뤄야 합니다.

## 상태·컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `JOB_NAME` | 문자열, 예: `foo` 또는 `foo/bar` | 상태·컨텍스트 | 현재 빌드가 속한 Job(프로젝트)의 이름 | 보조 신호 — 이름이 매우 일반적이라 단독 사용 불가. `JENKINS_URL`·`BUILD_NUMBER`와 결합 시 상태 정보로 유용 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `Job.java#L402`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/Job.java#L402) |
| `NODE_NAME` | 문자열. 컨트롤러(빌트인 노드)에서는 `master`(레거시) 또는 `built-in`(마이그레이션 완료 후), 에이전트에서는 에이전트 이름 | 상태·컨텍스트 | 빌드가 실행 중인 노드 이름 | 보조 신호 — 이름이 매우 일반적이고, 컨트롤러 값 자체도 인스턴스 마이그레이션 상태·버전에 따라 달라져 고정 문자열로 비교할 수 없음 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [Built-In Node Name and Label Migration](https://www.jenkins.io/doc/book/managing/built-in-node-migration/), [공식 소스: `CoreEnvironmentContributor.java#L60-L63`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L60-L63) |
| `EXECUTOR_NUMBER` | 정수 문자열 (0부터 시작) | 상태·컨텍스트 | 같은 노드 내에서 이 빌드를 수행 중인 executor 번호 | 보조 신호 — 현재 스레드가 실제 `Executor`로 실행 중일 때만 주입되므로 Pipeline에서 `node`/`agent` 블록 밖에서는 존재하지 않을 수 있음 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `CoreEnvironmentContributor.java#L58-L59`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L58-L59) |
| `WORKSPACE` | 절대 디렉터리 경로 | 상태·컨텍스트 | 빌드 작업 디렉터리 경로 | 보조 신호 — 다른 빌드 도구도 동일한 이름을 흔히 쓰는 일반적 관례라 단독 신호로는 부적합 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `AbstractBuild.java#L959`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/AbstractBuild.java#L959) |
| `BUILD_URL` | 전체 URL 문자열, 예: `http://buildserver/jenkins/job/MyJobName/17/` | 상태·컨텍스트 | 이 빌드 결과 페이지의 URL | 보조 신호 — `JENKINS_URL`이 설정된 인스턴스에서만 주입되며 값 자체가 `JENKINS_URL`을 접두어로 포함하는 파생 정보 | [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables), [공식 소스: `CoreEnvironmentContributor.java#L36-L38`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L36-L38) |
| `JENKINS_HOME` | 디렉터리 경로 | 설정 | Jenkins 컨트롤러의 루트 데이터 디렉터리 | 부적합 — 빌드 실행 여부와 무관하게 컨트롤러 머신 전역에 설정될 수 있는 설정값이며, 이 변수 하나만으로는 "이 프로세스가 Jenkins 빌드로 실행됐다"를 증명하지 못함 | [공식 소스: `CoreEnvironmentContributor.java#L53-L54`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L53-L54) |
| `HUDSON_URL` | `JENKINS_URL`과 동일한 값 | 상태·컨텍스트 | Jenkins의 전신인 Hudson 시절 이름과의 하위 호환용 별칭 | 부적합 단독 — `JENKINS_URL`이 이미 주입된 경우에만 함께 채워지는 레거시 중복 변수이므로 별도 신호로 취급할 필요가 없음 | [공식 소스: `CoreEnvironmentContributor.java#L49`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L49) |
| `HUDSON_HOME` | `JENKINS_HOME`과 동일한 값 | 설정 | Hudson 시절 이름과의 하위 호환용 별칭 | 부적합 — `JENKINS_HOME`과 동일한 이유로 단독 신호가 아님 | [공식 소스: `CoreEnvironmentContributor.java#L55`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java#L55) |

### 의도적으로 제외한 변수

- `BUILD_DISPLAY_NAME`: 사용자가 자유롭게 재정의할 수 있는 표시용 이름이라 안정적 식별에 부적합해 제외했습니다.
- `JOB_URL`, `NODE_LABELS`, `WORKSPACE_TMP`: 각각 `JENKINS_URL`/`JOB_NAME`, `NODE_NAME`, `WORKSPACE`에서 파생되는 보조 정보라 별도 표로 다루지 않았습니다.
- `JAVA_HOME`: Job에서 특정 JDK를 지정했을 때만 그 JDK 경로로 덮어써지는 도구 설정값이며, Jenkins 실행 자체와 무관해 제외했습니다.
- `BRANCH_NAME`, `CHANGE_ID`, `CHANGE_URL`, `TAG_NAME` 등 Multibranch Pipeline 변수와 `GIT_COMMIT`, `GIT_BRANCH`, `GIT_URL` 등 Git 관련 변수는 **코어가 아니라 플러그인(Pipeline Multibranch/`branch-api`, Git 플러그인)이 주입**하므로 이 문서의 핵심 감지 표에서 제외했습니다. 근거는 아래 "코어 대 플러그인 제공" 절에서 다룹니다.
- `SVN_REVISION`, `SVN_URL` 등 Subversion 플러그인 변수도 같은 이유로 제외했습니다.

### 코어 대 플러그인 제공

`jenkins.model.CoreEnvironmentContributor`(`@Extension(ordinal = -100) @Symbol("core")`)가 직접 주입하는 `CI`, `JENKINS_URL`, `HUDSON_URL`, `JOB_URL`, `JENKINS_HOME`, `HUDSON_HOME`, `EXECUTOR_NUMBER`, `NODE_NAME`, `NODE_LABELS`, `BUILD_URL`, `BUILD_DISPLAY_NAME`과, `hudson.model.Run`/`hudson.model.Job`/`hudson.model.AbstractBuild`가 직접 주입하는 `BUILD_NUMBER`, `BUILD_ID`, `BUILD_TAG`, `JOB_NAME`, `WORKSPACE`, `WORKSPACE_TMP`는 **Jenkins 코어(`jenkinsci/jenkins`) 자체가 보장**합니다. 반면 `BRANCH_NAME`/`CHANGE_*`/`TAG_*` 계열은 Multibranch Pipeline 기능(`branch-api`/Pipeline Multibranch 플러그인)이, `GIT_*` 계열은 Git 플러그인이 주입합니다. 이런 플러그인은 대부분의 Jenkins 배포판에 기본 포함되지만 **공식적으로 보장된 코어 계약이 아니므로**, `runby`는 감지 로직의 필수 조건으로 삼아서는 안 됩니다.

### Freestyle / Pipeline / 컨트롤러·에이전트 간 차이

- Job 범위 변수(`CI`, `JENKINS_URL`, `HUDSON_URL`, `JOB_URL`, `JENKINS_HOME`, `HUDSON_HOME`)는 Job 종류(Freestyle, Pipeline)와 무관하게 빌드가 시작되면 항상 계산되어 주입됩니다.
- 노드/executor 범위 변수(`EXECUTOR_NUMBER`, `NODE_NAME`, `NODE_LABELS`)는 코드상 `Thread.currentThread()`가 실제 `Executor`일 때만 채워집니다(`CoreEnvironmentContributor.java#L57-L68`). Freestyle 빌드는 항상 executor 위에서 실행되므로 이 값이 항상 존재하지만, 스크립트형(Scripted) Pipeline에서 `node {}` 블록 진입 전에 실행되는 최상위 스크립트 코드에는 이 값들이 없을 수 있습니다. 선언형(Declarative) Pipeline은 `agent` 지시자가 필수라 스테이지 내부에서는 항상 존재한다고 볼 수 있습니다.
- `WORKSPACE`는 Freestyle/Matrix 빌드에서는 `AbstractBuild`가 직접 채우고, Pipeline에서는 워크스페이스를 할당하는 스텝(`node`/`ws` 등)이 실행되어야 채워집니다. 즉 워크스페이스를 할당받기 전의 Pipeline 코드에는 존재하지 않을 수 있습니다.
- 컨트롤러에서 실행되는 빌드와 에이전트에서 실행되는 빌드는 `NODE_NAME`(컨트롤러는 `master`/`built-in`, 에이전트는 에이전트 이름)과 `WORKSPACE` 경로만 다르며, `JENKINS_URL`·`BUILD_ID`·`BUILD_NUMBER`·`CI` 등 Job 범위 변수는 컨트롤러가 계산해 원격 에이전트로 전달하므로 동일합니다.

### NODE_NAME의 컨트롤러 특수값

공식 소스(`jenkins.model.Jenkins#getSelfLabel()`)에 따르면 컨트롤러(빌트인 노드)의 `NODE_NAME`은 다음 우선순위로 결정됩니다.

1. 시스템 프로퍼티 `jenkins.model.Jenkins.nodeNameAndSelfLabelOverride`가 설정된 경우 그 값.
2. 그렇지 않고 빌트인 노드 이름 마이그레이션이 완료된 경우(새 설치는 기본 완료, 기존 설치는 관리자가 명시적으로 마이그레이션해야 함) `built-in`.
3. 마이그레이션이 아직 안 된 기존 인스턴스에서는 레거시 값 `master`.

즉 컨트롤러의 `NODE_NAME`은 `master`와 `built-in` 중 어느 쪽도 고정적으로 보장되지 않으며, 관리자가 완전히 다른 문자열로 재정의할 수도 있습니다. `runby`는 `NODE_NAME`의 값이 아니라 "존재 여부"만 상태 정보로 참고해야 합니다.

## 실행 주체 감지에 관한 결론

Jenkins에는 다른 CI 도구들의 `GITHUB_ACTIONS`, `CIRCLECI`처럼 Jenkins 전용으로 이름 붙은 불리언 마커가 없습니다. 이것이 이 문서에서 가장 어려운 지점입니다. `CI`는 Jenkins 코어가 보장하지만 범용 CI 관례라 다른 도구와 구분되지 않고, `BUILD_ID`/`BUILD_NUMBER`/`JOB_NAME`/`NODE_NAME`은 이름 자체가 지나치게 일반적이어서 Jenkins가 아닌 스크립트나 다른 빌드 도구가 같은 이름을 우연히 또는 의도적으로 설정할 수 있습니다.

이런 이유로 `runby`는 다음 우선순위를 권장합니다.

1. **`JENKINS_URL` 존재 + (`BUILD_ID` 또는 `BUILD_NUMBER`) 존재** → 적합. `JENKINS_URL`은 이름에 `jenkins`가 명시적으로 들어 있는 유일한 핵심 변수이고, 여기에 빌드 식별자가 함께 있으면 "Jenkins 인스턴스가 만든 빌드 실행 환경"이라는 조합적 신뢰도가 크게 올라갑니다.
2. **`JENKINS_URL`이 없는 경우**(관리자가 시스템 설정에서 Jenkins URL을 지정하지 않은 인스턴스) `CI=true` + `BUILD_NUMBER` + `JOB_NAME`을 함께 확인 → 보조 신호로만 취급. 개별 변수가 모두 일반적인 이름이므로 이 조합도 확정적 증거는 아닙니다.
3. `NODE_NAME`, `WORKSPACE`, `EXECUTOR_NUMBER`, `BUILD_URL`, `BUILD_TAG`, `JENKINS_HOME`, `HUDSON_URL`/`HUDSON_HOME`은 1·2번 판정이 성립한 뒤 상태 보강 정보로만 사용합니다. 단독으로는 판정에 쓰지 않습니다.

남는 오탐 위험은 두 가지입니다. 첫째, 환경변수는 자식 프로세스로 상속되므로 Jenkins 빌드 스텝이 실행한 프로세스가 또 다른 프로세스를 실행하면 그 하위 프로세스도 동일한 변수를 물려받습니다 — 즉 "Jenkins가 직접 실행" 여부가 아니라 "Jenkins 빌드 환경의 후손"임을 감지하는 것입니다. 둘째, 이 변수들은 이름이 알려져 있고 값 형식도 문서화되어 있어 사용자가 로컬 셸이나 테스트 스크립트에서 얼마든지 위조할 수 있습니다. 절대적인 신뢰 경계가 아니므로 보안이 중요한 판단에는 사용하지 말아야 합니다.

## 공식 문서

- [Using a Jenkinsfile — Environment variables](https://www.jenkins.io/doc/book/pipeline/jenkinsfile/#using-environment-variables)
- [Jenkins 2.286 변경 로그 (CI 변수 도입)](https://www.jenkins.io/changelog/2.286/)
- [JENKINS-36707 — CI=true 요청 이슈](https://issues.jenkins.io/browse/JENKINS-36707)
- [Built-In Node Name and Label Migration](https://www.jenkins.io/doc/book/managing/built-in-node-migration/)
- [Pipeline Multibranch — env 변수](https://www.jenkins.io/doc/book/pipeline/multibranch/)
- [공식 Jenkins 소스 저장소 (`jenkinsci/jenkins`)](https://github.com/jenkinsci/jenkins)
- [공식 소스: `CoreEnvironmentContributor.java`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/CoreEnvironmentContributor.java)
- [공식 소스: `Run.java`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/Run.java)
- [공식 소스: `Job.java`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/Job.java)
- [공식 소스: `AbstractBuild.java`](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/hudson/model/AbstractBuild.java)
- [공식 소스: `Jenkins.java` (`getSelfLabel`)](https://github.com/jenkinsci/jenkins/blob/982bc91d866ed90aa135b87a2cb4ac1e68c2412e/core/src/main/java/jenkins/model/Jenkins.java)
