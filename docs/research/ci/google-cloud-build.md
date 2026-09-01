---
title: Google Cloud Build
slug: google-cloud-build
research_date: 2026-09-02
open_source: false
repository: null
product_type: ci_platform
executes_agents: []
runtime_test_required: false
runtime_test_reason: Google이 관리하는 cloud-builders 공식 소스가 BUILDER_OUTPUT을 builder 컨테이너의 출력 디렉터리 계약으로 사용하고 Cloud Build 공식 문서가 builder step 실행 모델을 정의하므로 별도 실행 검증 없이 ci-info 호환 마커를 적용할 수 있음
---

# Google Cloud Build

Google Cloud Build는 빌드 step마다 builder 컨테이너를 실행합니다. Google이 관리하는 공식 builder 소스는 `BUILDER_OUTPUT`을 Cloud Build builder 환경에서 제공되는 출력 디렉터리로 사용하므로, ci-info와의 호환을 위해 이 변수를 존재 마커로 채택합니다.

## 실행 식별 신호

| 환경변수 | 값/자료형 | 종류 | 용도 | CI 실행 감지 | 공식 출처 |
|---|---|---|---|---|---|
| `BUILDER_OUTPUT` | builder 출력 디렉터리 경로 | 실행 식별 | Cloud Build builder가 결과를 쓰는 위치 | 적합 — Google 공식 builder 소스가 Cloud Build 환경 계약으로 사용하지만, 일반 Cloud Build step에서 항상 노출된다고 확장해 주장하지 않음 | [cloud-builders: gcs-fetcher](https://github.com/GoogleCloudPlatform/cloud-builders/blob/master/gcs-fetcher/cmd/gcs-fetcher/main.go) |
| `env` / `secretEnv` | build config가 정한 값 | 설정 | step에 사용자·비밀 환경변수 전달 | 부적합 — 사용자가 지정하는 설정이고 비밀 값일 수 있어 읽지 않음 | [Build config file schema](https://docs.cloud.google.com/build/docs/build-config-file-schema) |

`BUILDER_OUTPUT`은 이 패키지가 채택하는 ci-info 호환 기준이며, Cloud Build API의 `$BUILD_ID` 같은 substitution은 step 환경변수로 자동 매핑되지 않을 수 있으므로 일반 marker로 사용하지 않습니다. Evidence에는 `BUILDER_OUTPUT` 이름만 기록합니다.

## 공식 문서

- [Build config file schema](https://docs.cloud.google.com/build/docs/build-config-file-schema)
- [Substitute variable values](https://docs.cloud.google.com/build/docs/configuring-builds/substitute-variable-values)
- [Google Cloud official builder images](https://github.com/GoogleCloudPlatform/cloud-builders)
- [BUILDER_OUTPUT source usage](https://github.com/GoogleCloudPlatform/cloud-builders/blob/master/gcs-fetcher/cmd/gcs-fetcher/main.go)
