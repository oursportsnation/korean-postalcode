# Deployment Guide

`korean-postalcode`를 GitHub에 배포하고 설정하는 전체 가이드입니다.

## 📦 GitHub 저장소 생성 및 배포

### 1. GitHub 저장소 생성

**Option A: GitHub CLI 사용 (권장)**

```bash
cd korean-postalcode

# GitHub CLI로 저장소 생성 및 초기 푸시
gh repo create your-org/korean-postalcode \
  --public \
  --description "한국 우편번호 및 도로명주소 데이터를 관리하는 재사용 가능한 Go 패키지" \
  --source=. \
  --remote=origin \
  --push
```

**Option B: 웹 인터페이스 사용**

1. GitHub에서 새 저장소 생성
   - Repository name: `korean-postalcode`
   - Description: `한국 우편번호 및 도로명주소 데이터를 관리하는 재사용 가능한 Go 패키지`
   - Visibility: **Public** ✅
   - Initialize: ❌ (README, .gitignore, LICENSE 이미 존재)

2. 로컬에서 Git 초기화 및 푸시

```bash
cd korean-postalcode

# Git 초기화
git init
git add .
git commit -m "Initial commit: Korean PostalCode library

- 한국 우편번호 및 도로명주소 패키지
- Repository, Service, Handler 레이어 분리
- REST API 지원 (표준 HTTP, Gin)
- 고성능 검색 (우편번호 prefix 인덱스)
- CLI 도구 및 배치 import 지원
- 완전한 문서 및 예제"

# 원격 저장소 추가
git branch -M main
git remote add origin https://github.com/your-org/korean-postalcode.git

# 푸시
git push -u origin main
```

### 2. 첫 번째 릴리즈 태그 생성

```bash
# v1.0.0 태그 생성
git tag -a v1.0.0 -m "Release v1.0.0

첫 번째 공식 릴리즈

Features:
- 우편번호 검색 API
- 도로명주소 조회 기능
- REST API 핸들러 (표준 HTTP, Gin)
- CLI Import 도구
- 31만건 데이터 고성능 처리

Performance:
- 우편번호 prefix 검색: ~1-5ms
- 정확한 우편번호 조회: ~1-3ms

Documentation:
- 완전한 API 문서
- 통합 가이드
- 마이그레이션 가이드
- 실행 가능한 예제"

# 태그 푸시 (Release 자동 생성됨)
git push origin v1.0.0
```

### 3. GitHub 저장소 설정

#### Repository Settings

**About (저장소 상단)**
- Description: `한국 우편번호 및 도로명주소 데이터를 관리하는 재사용 가능한 Go 패키지`
- Website: `https://pkg.go.dev/github.com/oursportsnation/korean-postalcode`
- Topics: `go`, `golang`, `postal-code`, `korea`, `address`, `korean-address`, `gorm`, `rest-api`

**Features (Settings → General)**
- ✅ Issues
- ✅ Projects
- ❌ Wiki (문서는 docs/ 디렉토리 사용)
- ❌ Sponsorships
- ✅ Discussions (선택)

**Pull Requests**
- ✅ Allow squash merging
- ✅ Allow merge commits
- ✅ Allow rebase merging
- ✅ Automatically delete head branches

**Actions (Settings → Actions)**
- ✅ Allow all actions and reusable workflows

**Pages (Settings → Pages)**
- Source: Deploy from a branch
- Branch: `main` / `docs`
- ✅ Enforce HTTPS

## 🔒 Secrets 설정 (선택)

GitHub Actions에서 사용할 Secrets:

```bash
# GitHub 웹 인터페이스에서
# Settings → Secrets and variables → Actions → New repository secret
```

**필요한 Secrets:**
- `CODECOV_TOKEN` (선택): Codecov 통합
- 기타 추가 Secret은 필요시 설정

## 📊 pkg.go.dev 등록

**자동 등록:**
- v1.0.0 태그가 푸시되면 자동으로 pkg.go.dev에 등록됨
- 약 10-30분 소요

**수동 요청 (빠른 등록):**

```bash
# 브라우저에서 접속
https://pkg.go.dev/github.com/oursportsnation/korean-postalcode

# 또는
curl https://sum.golang.org/lookup/github.com/oursportsnation/korean-postalcode@v1.0.0
```

## 📝 README Badges 추가

저장소의 README.md 상단에 추가 배지:

```markdown
[![Go Reference](https://pkg.go.dev/badge/github.com/oursportsnation/korean-postalcode.svg)](https://pkg.go.dev/github.com/oursportsnation/korean-postalcode)
[![Go Report Card](https://goreportcard.com/badge/github.com/oursportsnation/korean-postalcode)](https://goreportcard.com/report/github.com/oursportsnation/korean-postalcode)
[![CI](https://github.com/oursportsnation/korean-postalcode/workflows/CI/badge.svg)](https://github.com/oursportsnation/korean-postalcode/actions)
[![codecov](https://codecov.io/gh/oursportsnation/korean-postalcode/branch/main/graph/badge.svg)](https://codecov.io/gh/oursportsnation/korean-postalcode)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

## 🔄 지속적인 유지보수

### 버전 관리 (Semantic Versioning)

```bash
# 버그 픽스 (v1.0.0 → v1.0.1)
git tag -a v1.0.1 -m "Bug fix: ..."
git push origin v1.0.1

# 새 기능 (v1.0.1 → v1.1.0)
git tag -a v1.1.0 -m "Feature: ..."
git push origin v1.1.0

# Breaking changes (v1.1.0 → v2.0.0)
git tag -a v2.0.0 -m "Breaking: ..."
git push origin v2.0.0
```

### 브랜치 전략

```
main         - 안정 버전
develop      - 개발 브랜치
feature/*    - 새 기능
bugfix/*     - 버그 수정
hotfix/*     - 긴급 수정
```

## ✅ 배포 완료 체크리스트

- [ ] GitHub 저장소 생성됨
- [ ] 코드 푸시 완료
- [ ] v1.0.0 태그 생성됨
- [ ] GitHub Actions CI 통과
- [ ] pkg.go.dev에 등록됨
- [ ] README 배지 추가됨
- [ ] 빌드 및 테스트 성공
- [ ] 문서 검토 완료

## 📞 지원 및 커뮤니티

- **Issues**: https://github.com/oursportsnation/korean-postalcode/issues
- **Discussions**: https://github.com/oursportsnation/korean-postalcode/discussions
- **Pull Requests**: 기여를 환영합니다!

---

**배포 시간**: 약 15-30분
**다운타임**: 0
