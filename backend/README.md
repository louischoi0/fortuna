# Fortuna Backend API

FastAPI 기반의 백엔드 API 서버입니다.

## 프로젝트 구조

```
backend/
├── app/
│   ├── __init__.py
│   ├── config.py          # 설정 관리
│   ├── database/
│   │   ├── __init__.py
│   │   ├── connection.py  # 데이터베이스 연결
│   │   ├── session.py     # 세션 관리
│   │   └── init_db.py     # DB 초기화
│   ├── models/
│   │   ├── __init__.py
│   │   ├── schemas.py     # Pydantic 모델
│   │   └── database_models.py # SQLAlchemy 모델
│   ├── routers/
│   │   ├── __init__.py
│   │   └── api.py         # API 라우터
│   └── services/
│       ├── __init__.py
│       ├── base.py        # 기본 서비스
│       ├── user_service.py
│       ├── space_service.py
│       └── transaction_service.py
├── alembic/               # 데이터베이스 마이그레이션
│   ├── versions/
│   ├── env.py
│   └── script.py.mako
├── tests/
│   ├── __init__.py
│   └── test_main.py       # 테스트 파일
├── main.py                # FastAPI 애플리케이션 진입점
├── run.py                 # 애플리케이션 실행 스크립트
├── init_database.py       # 데이터베이스 초기화 스크립트
├── setup.sh              # 자동 설정 스크립트
├── alembic.ini           # Alembic 설정
├── requirements.txt       # Python 의존성
├── env.example           # 환경변수 예시
└── README.md
```

## 설치 및 실행

### 1. 가상환경 생성 및 활성화

```bash
# 가상환경 생성
python -m venv venv

# 가상환경 활성화 (macOS/Linux)
source venv/bin/activate

# 가상환경 활성화 (Windows)
venv\Scripts\activate
```

### 2. 의존성 설치

```bash
pip install -r requirements.txt
```

### 3. 환경변수 설정

```bash
# 환경변수 파일 복사
cp env.example .env

# 필요에 따라 .env 파일 수정
```

### 4. 데이터베이스 설정

```bash
# PostgreSQL 설치 (macOS)
brew install postgresql
brew services start postgresql

# 데이터베이스 생성
createdb -U fortuna fortuna_db

# 또는 PostgreSQL에 직접 접속하여 생성
psql -U postgres
CREATE DATABASE fortuna_db;
CREATE USER fortuna WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE fortuna_db TO fortuna;
\q
```

### 5. 데이터베이스 초기화

```bash
# 데이터베이스 테이블 생성
python init_database.py
```

### 6. 애플리케이션 실행

```bash
# 방법 1: main.py 직접 실행
python main.py

# 방법 2: run.py 사용
python run.py

# 방법 3: uvicorn 직접 사용
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

## API 문서

서버 실행 후 다음 URL에서 API 문서를 확인할 수 있습니다:

- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc

## 주요 엔드포인트

- `GET /` - 루트 엔드포인트
- `GET /health` - 헬스 체크
- `GET /api/v1/status` - API 상태
- `GET /api/v1/info` - API 정보

## 테스트 실행

```bash
# 모든 테스트 실행
pytest

# 특정 테스트 파일 실행
pytest tests/test_main.py

# 상세 출력과 함께 실행
pytest -v
```

## 개발 가이드

### 새로운 API 엔드포인트 추가

1. `app/routers/` 폴더에 새로운 라우터 파일 생성
2. `app/models/schemas.py`에 필요한 Pydantic 모델 추가
3. `main.py`에서 새로운 라우터 등록

### 설정 관리

`app/config.py`에서 애플리케이션 설정을 관리합니다. 환경변수를 통해 설정값을 오버라이드할 수 있습니다.

## 데이터베이스 모델

### 주요 엔티티
- **User**: 사용자 정보
- **Space**: Fortuna 스페이스
- **Transaction**: 트랜잭션 기록
- **Event**: 이벤트 로그
- **Replica**: 복제본 정보
- **OracleNode**: 오라클 노드 정보

### 관계
- User → Space (1:N)
- Space → Transaction (1:N)
- Space → Event (1:N)
- Space → Replica (1:N)
- Transaction → Event (1:N)

## 기술 스택

- **FastAPI**: 웹 프레임워크
- **Uvicorn**: ASGI 서버
- **PostgreSQL**: 관계형 데이터베이스
- **SQLAlchemy**: ORM
- **Alembic**: 데이터베이스 마이그레이션
- **Pydantic**: 데이터 검증 및 직렬화
- **Passlib**: 비밀번호 해싱
- **pytest**: 테스트 프레임워크
- **python-dotenv**: 환경변수 관리
