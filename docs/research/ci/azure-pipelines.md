---
title: Azure Pipelines
slug: azure-pipelines
research_date: 2026-08-31
open_source: true
repository: https://github.com/microsoft/azure-pipelines-agent
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Microsoft Learn 공식 predefined variables 문서와 microsoft/azure-pipelines-agent 공식 소스만으로 실행 식별·상태 변수의 존재와 값 형식을 확인할 수 있었고, 이 변수들은 호스트형·자체 호스트형 에이전트 모두에서 에이전트가 설정하는 공개 계약이라 별도 실행 검증 없이도 판정 근거로 충분함
---

# Azure Pipelines

Azure Pipelines는 SaaS로 제공되는 Azure DevOps 서비스이며, 서비스 자체는 오픈소스가 아닙니다. 다만 파이프라인 작업을 실제로 실행하고 predefined variable을 프로세스 환경에 주입하는 구성요소인 빌드 에이전트는 [`microsoft/azure-pipelines-agent`](https://github.com/microsoft/azure-pipelines-agent)로 공개되어 있습니다. Microsoft Learn의 predefined variables 문서는 이 변수들을 "자동으로 설정되며 읽기 전용"이라고 명시하므로, 표에 실은 변수들은 파이프라인 작성자가 임의로 이름을 바꿀 수 없는 안정적 공개 계약으로 취급할 수 있습니다. 단, 값 자체는 여느 환경변수와 마찬가지로 자식 프로세스에 상속되므로 하위 프로세스에서 관찰한 값이 반드시 "현재" 파이프라인 실행을 가리킨다고 단정할 수는 없습니다.

## 변수명 매핑 규칙

Microsoft Learn 문서는 YAML에서 변수를 `Build.BuildId`, `System.JobId`처럼 점(dot) 표기로 정의하지만, 스크립트나 하위 프로세스에는 **점을 밑줄로 바꾸고 전체를 대문자로 변환한** 환경변수로 전달된다고 명시합니다. 예를 들어 `Build.ArtifactStagingDirectory`는 `BUILD_ARTIFACTSTAGINGDIRECTORY`가 되고, `System.AccessToken`은 `SYSTEM_ACCESSTOKEN`이 됩니다. 이 문서의 표는 이 규칙에 따라 실제 환경변수명(`BUILD_BUILDID`, `SYSTEM_JOBID` 등)을 기준으로 정리했으며, 괄호 안에 YAML 표기를 병기합니다. `Agent.JobStatus` → `AGENT_JOBSTATUS`처럼 일부는 구버전 소문자 표기(`agent.jobstatus`)도 하위 호환으로 남아 있다고 문서에 명시되어 있지만, `runby`는 대문자·밑줄 표준형만 신뢰 가능한 계약으로 삼는 것이 안전합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `TF_BUILD` (`TF_BUILD`) | 불리언 문자열, 정확히 `True` (대문자 T) | 실행 식별 | 빌드 태스크가 스크립트를 실행 중임을 표시 | 적합 — 문서가 명시하는 가장 단순하고 범용적인 Azure Pipelines 실행 마커. `runby`는 대소문자 무관 불리언 파싱을 쓰므로 `True`/`true` 차이는 문제되지 않지만, 문서상 실제 값은 `True`로 고정 표기됨 | [Predefined variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `BUILD_BUILDID` (`Build.BuildId`) | 정수 ID 문자열 | 실행 식별 | 현재 빌드(파이프라인 실행) 레코드의 고유 ID | 적합 — 파이프라인 실행(run) 단위의 안정적 식별자 | [Predefined variables — Build variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `SYSTEM_JOBID` (`System.JobId`) | GUID 문자열 | 실행 식별 | 단일 잡의 단일 시도(attempt)에 대한 고유 식별자. 현재 파이프라인 내에서만 유일함 | 적합 — job 단위 실행 식별에 가장 구체적인 신호. 템플릿 확장 시점에는 사용할 수 없음(런타임에만 값 존재) | [Predefined variables — System variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `BUILD_BUILDNUMBER` (`Build.BuildNumber`) | 사용자 정의 형식의 문자열(run number) | 실행 식별 | 완료된 빌드의 표시 이름. 공백이나 라벨에 부적합한 문자를 포함할 수 있음 | 보조 신호 — 실행을 사람이 읽을 수 있게 표시하지만 사용자 지정 형식이라 고유성이 `Build.BuildId`만큼 보장되지 않음 | [Run and build numbers](https://learn.microsoft.com/en-us/azure/devops/pipelines/process/run-number?view=azure-devops) |
| `SYSTEM_JOBATTEMPT` (`System.JobAttempt`) | 정수, 최초 시도 시 `1` | 실행 식별 | 잡이 재시도될 때마다 증가하는 시도 횟수 | 보조 신호 — job 자체를 식별하지는 않지만 `SYSTEM_JOBID`와 함께 재시도 여부를 판단하는 데 사용 | [Predefined variables — System variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |

`System.StageId`는 Microsoft Learn의 predefined variables 문서에 **존재하지 않습니다.** 문서가 실제로 제공하는 스테이지 수준 식별자는 문자열 기반의 `System.StageName`(`SYSTEM_STAGENAME`)이며, GUID 형태의 안정적 스테이지 ID는 공식적으로 문서화되어 있지 않습니다. 아래 표에서 `SYSTEM_STAGENAME`을 스테이지 컨텍스트 변수로 다루되, GUID 성격의 `SYSTEM_STAGEID`를 기대하고 파싱하는 코드는 작성하지 않아야 합니다. 스테이지 재시도 횟수는 `System.StageAttempt`(`SYSTEM_STAGEATTEMPT`, 최초 시도 `1`부터 증가)로 문서화되어 있습니다. 레거시 개념인 phase에도 동일한 형태의 `System.PhaseAttempt`(`SYSTEM_PHASEATTEMPT`)가 존재하지만, 문서는 "phase는 matrix·multi-config 잡에서만 job과 구분되는 대체로 중복된 개념"이라고 명시합니다.

## 상태·컨텍스트 변수

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `BUILD_REASON` (`Build.Reason`) | 고정 문자열 집합: `Manual`, `IndividualCI`, `BatchedCI`, `Schedule`, `ValidateShelveset`, `CheckInShelveset`, `PullRequest`, `BuildCompletion`, `ResourceTrigger` | 상태·컨텍스트 | 빌드를 유발한 트리거 종류 | 보조 신호 — Azure Pipelines 실행이라는 전제하에 트리거 유형을 알려주지만 그 자체로 실행 주체를 증명하지는 않음 | [Predefined variables — Build variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `SYSTEM_STAGENAME` (`System.StageName`) | 문자열 식별자 | 상태·컨텍스트 | 현재 스테이지의 문자열 기반 ID(의존성·출력 변수 참조용) | 보조 신호 — 스테이지 컨텍스트를 제공하지만 GUID 수준의 고유성은 보장되지 않음 | [Predefined variables — System variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `SYSTEM_JOBDISPLAYNAME` (`System.JobDisplayName`) | 사람이 읽는 이름 문자열 | 상태·컨텍스트 | 잡의 표시 이름 | 보조 신호 — 표시용이며 고유 식별자가 아님 | [Predefined variables — System variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `BUILD_DEFINITIONNAME` (`Build.DefinitionName`) | 문자열 | 상태·컨텍스트 | 파이프라인 정의 이름 | 보조 신호 — 어떤 파이프라인 정의가 실행 중인지 알려주지만 실행 여부 자체의 증거는 아님 | [Predefined variables — Build variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `SYSTEM_DEFINITIONID` (`System.DefinitionId`) | 정수 ID | 상태·컨텍스트 | 파이프라인 정의의 안정적 ID | 보조 신호 — 정의 자체는 실행마다 고정이므로 단독으로는 실행 여부를 증명하지 못함 | [Predefined variables — System variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `BUILD_SOURCEBRANCH` (`Build.SourceBranch`) | ref 문자열, 예: `refs/heads/main`, `refs/pull/1/merge` | 상태·컨텍스트 | 빌드 대상 브랜치/ref | 보조 신호 — 실행 컨텍스트 정보이며 실행 주체 판정에는 쓰이지 않음 | [Predefined variables — Build variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `BUILD_SOURCEVERSION` (`Build.SourceVersion`) | Git commit ID 또는 TFVC changeset 번호 | 상태·컨텍스트 | 이 빌드에 포함된 최신 소스 버전 | 보조 신호 — 커밋 컨텍스트 정보 | [Predefined variables — Build variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `BUILD_REPOSITORY_NAME` / `BUILD_REPOSITORY_PROVIDER` (`Build.Repository.Name` / `Build.Repository.Provider`) | 문자열 / 고정 집합(`TfsGit`, `TfsVersionControl`, `Git`, `GitHub`, `Svn`) | 상태·컨텍스트 | 트리거한 저장소 이름과 종류 | 보조 신호 — 저장소 컨텍스트 정보 | [Predefined variables — Build variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `SYSTEM_HOSTTYPE` (`System.HostType`) | 고정 집합: 빌드는 `build`, 릴리스는 `deployment`/`gates`/`release` | 상태·컨텍스트 | 현재 실행이 빌드인지 클래식 릴리스의 어느 단계인지 구분 | 보조 신호 — 빌드/릴리스 계열을 구분하는 데 유용하지만 그 자체가 존재 여부를 증명하지는 않음 | [Predefined variables — System variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `AGENT_NAME` (`Agent.Name`) | 문자열 | 상태·컨텍스트 | 풀에 등록된 에이전트 이름. 자체 호스트형은 사용자가 직접 지정 | 보조 신호 — 자체 호스트형에서는 임의 문자열이라 위조·중복 가능성이 있어 신원 증명에는 약함 | [Predefined variables — Agent variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `AGENT_ID` (`Agent.Id`) | 정수 ID | 상태·컨텍스트 | 에이전트의 ID | 보조 신호 — 에이전트 자체 식별자이며 파이프라인 실행 식별자는 아님 | [Predefined variables — Agent variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |
| `SYSTEM_PULLREQUEST_PULLREQUESTID` (`System.PullRequest.PullRequestId`) | 정수 ID | 상태·컨텍스트 | 이 빌드를 유발한 PR ID. Git PR 브랜치 정책이 적용된 빌드에서만 초기화됨 | 보조 신호 — `BUILD_REASON=PullRequest`인 경우에 한해 존재 | [Predefined variables — System variables](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops) |

다음 변수들은 의도적으로 표에서 제외했습니다.

- **`Build.QueuedBy` / `Build.RequestedFor` 계열(사용자 신원 변수)**: 문서에 따르면 트리거 방식(스케줄, 게이트 체크인 등)에 따라 시스템 계정 문자열로 채워질 수 있어 실제 사용자 신원을 신뢰성 있게 보장하지 않으며, PII 성격이 강해 `runby`가 다루는 "실행 주체 감지"의 범위를 벗어납니다.
- **`Agent.BuildDirectory`, `Build.ArtifactStagingDirectory`, `Build.SourcesDirectory` 등 경로류 변수**: 실행 환경의 파일 시스템 배치일 뿐 실행 식별과 무관하며, 자체 호스트형 에이전트에서는 사용자가 구성을 바꿀 수 있습니다.
- **`System.AccessToken`**: 보안 토큰이며 값 자체를 로그나 감지 로직에 노출하면 안 되는 민감 정보입니다.
- **`System.PullRequest.SourceBranch`/`TargetBranch` 등 PR 상세 컨텍스트**: `SYSTEM_PULLREQUEST_PULLREQUESTID`로 PR 트리거 여부를 판단하기에 충분하며, 나머지는 `runby`가 노출할 실행 메타데이터의 범위를 넘는 세부 정보로 판단해 생략했습니다.
- **`Checks.StageAttempt`**: 환경(Environment) 승인·체크 컨텍스트 내부에서만 사용 가능한 변수로, 일반 잡 실행 감지와는 다른 별도 실행 맥락입니다.

## 실행 주체 감지에 관한 결론

`runby`가 Azure Pipelines 실행을 감지할 때는 `TF_BUILD == True`(대소문자 무관 비교 시 `true`)를 1차 판정 신호로 삼는 것이 가장 안전합니다. 문서가 이 변수를 "빌드 태스크가 스크립트를 실행 중"이라는 좁고 명확한 의미로 정의하고 있고, 값 형태도 단순한 불리언이기 때문입니다. `BUILD_BUILDID`와 `SYSTEM_JOBID`의 존재는 보조 확인 신호로 함께 확인하면 오탐 가능성을 더 낮출 수 있습니다 — 세 변수가 함께 존재하면 실제 Azure Pipelines 에이전트가 주입한 실행 컨텍스트일 가능성이 높습니다.

정밀도가 필요한 계층 구조는 다음 우선순위로 읽어야 합니다.

1. **실행(run) 수준**: `BUILD_BUILDID` — 파이프라인 실행 전체를 가리키는 안정적 ID.
2. **잡(job) 수준**: `SYSTEM_JOBID` — 단일 잡의 단일 시도를 가리키는 GUID.
3. **스테이지 수준**: `SYSTEM_STAGENAME` — GUID가 아닌 문자열 식별자만 공식 문서화되어 있음(`SYSTEM_STAGEID`는 존재하지 않음).
4. **재시도 횟수**: `SYSTEM_JOBATTEMPT`(잡 재시도), `SYSTEM_STAGEATTEMPT`(스테이지 재시도), 레거시 `SYSTEM_PHASEATTEMPT`(matrix/multi-config 잡의 phase 재시도) — 모두 최초 시도 시 `1`에서 시작해 재시도마다 증가.

트리거 종류는 `BUILD_REASON`으로 확인할 수 있으며, 문서가 명시하는 값 집합은 `Manual`, `IndividualCI`, `BatchedCI`, `Schedule`, `ValidateShelveset`, `CheckInShelveset`, `PullRequest`, `BuildCompletion`, `ResourceTrigger` 입니다.

호스팅 환경에 따른 차이도 고려해야 합니다.

- **자체 호스트형(self-hosted) 에이전트**: `AGENT_NAME`은 사용자가 직접 지정하므로 여러 에이전트가 같은 이름을 가질 수 있고, 실행 주체 증명이 아닌 정보용 라벨로만 취급해야 합니다.
- **컨테이너 잡**: `microsoft/azure-pipelines-agent` 저장소의 과거 이슈([`#2039`](https://github.com/microsoft/azure-pipelines-agent/issues/2039))는 컨테이너로 실행되는 잡에 `TF_BUILD`가 전달되지 않은 사례를 보고했습니다. 이는 공식 predefined variables 문서에 기술된 보장은 아니며 에이전트 버전에 따라 달라질 수 있는 구현 세부사항이므로, 컨테이너 잡에서는 `TF_BUILD` 단독 판정 대신 `BUILD_BUILDID`/`SYSTEM_JOBID` 존재 여부도 함께 확인하는 편이 안전합니다.
- **배포(deployment) 잡**: `System.HostType`이 `deployment`로 설정되고 `Environment.Name` 등 별도의 환경 변수 계열이 추가되지만, 이 문서가 다루는 실행 식별 변수(`TF_BUILD`, `BUILD_BUILDID`, `SYSTEM_JOBID` 등)의 존재 자체는 동일하게 유지됩니다.
- **클래식(비-YAML) 파이프라인**: 클래식 빌드 정의는 이 문서의 build 변수들을 그대로 사용하지만, 클래식 릴리스 파이프라인은 별도의 `Release.*` 변수 체계를 사용하며 이 문서에서는 다루지 않았습니다.

마지막으로, 이 모든 변수는 일반 환경변수이므로 자식 프로세스에 상속되거나 로컬 셸에서 사용자가 동일한 이름으로 재정의(위조)할 수 있습니다. `TF_BUILD=True`를 셸 프로파일에 직접 export한 로컬 환경에서는 `runby`가 오탐할 수 있으므로, 이 신호는 "Azure Pipelines 에이전트가 실행했다는 강한 정황 증거"로 취급하되 위조 불가능한 절대적 신뢰 경계로는 취급하지 않아야 합니다.

## 공식 문서

- [Predefined variables — Azure Pipelines](https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops)
- [Define variables — Azure Pipelines](https://learn.microsoft.com/en-us/azure/devops/pipelines/process/variables?view=azure-devops)
- [Run and build numbers](https://learn.microsoft.com/en-us/azure/devops/pipelines/process/run-number?view=azure-devops)
- [Deployment jobs — Azure Pipelines](https://learn.microsoft.com/en-us/azure/devops/pipelines/process/deployment-jobs?view=azure-devops)
- [Use variables in Classic release pipelines](https://learn.microsoft.com/en-us/azure/devops/pipelines/release/variables?view=azure-devops)
- [microsoft/azure-pipelines-agent (공식 소스)](https://github.com/microsoft/azure-pipelines-agent)

## Pull request 감지

`BUILD_REASON`의 공식 값이 `PullRequest`이면 `runby`는 `PullRequest=true`로 정규화합니다. PR ID가 초기화되는 경우에는 `SYSTEM_PULLREQUEST_PULLREQUESTID`를 `PullRequestID`로 사용하지만, 이 변수는 모든 PR 실행 형태에서 보장되는 것은 아니므로 값이 없어도 요청 판정은 유지합니다. `BUILD_REASON`과 ID 변수는 `Evidence`에 이름만 기록합니다.
