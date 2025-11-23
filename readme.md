# TDE 핵심

1. 어플리케이션 코드는 수정 필요 X
2. 키를 어떻게 관리를 할것인가

# Data at Rest Encryption

# Key Management Plugin

암호화 키 관리를 어떻게 할지 결정하는 플러그인

# InnoDB Tablespace Encryption

실제 데이터 저장되는 .ibd 파일 암호화 설정

- 제공하는 기능들

- 활용 가능한 방향들 + 데모

- 레플리케이션 + 갈레라

* 암호화 키의 동기화, 운영상 모든노드가 동일한 키 파일을 바라보게 구성
* 데이터 파일뿐 아니라 네트워크를 타고 넘어가는 Binary log도 암호화해야 함..

- 적용전후 오버헤드
  sysbench

# 세팅과정

0. 디렉토리 세팅

```
mariadb-tde-demo/
├── docker-compose.yml
├── config/
│   └── my.cnf          # TDE 설정 파일
├── keys/
│   ├── keys.txt        # (임시) 평문 키 파일 (생성 후 삭제 예정)
│   ├── keys.enc        # (실제) 암호화된 키 파일
│   └── password        # 키 파일을 풀기 위한 비밀번호
└── data/               # DB 데이터 저장소 (자동 생성됨)
```

1. docker 이미지 작성

2. my.cnf 세팅하기

3. 키 파일을 생성한다

```
# 1. keys 폴더로 이동
cd keys

# 2. 암호화에 사용할 비밀번호 파일 생성 (내용: mysecretpassword)
echo -n "mysecretpassword" > password

# 3. 평문 키 파일 생성 (형식: 키ID;32바이트_Hex_키값)
# 아래는 예시 키입니다. (ID 1번, 2번 키 생성)
echo "1;5A2B3C4D5E6F708192A3B4C5D6E7F8091A2B3C4D5E6F708192A3B4C5D6E7F809" > keys.txt
echo "2;908F7E6D5C4B3A291807F6E5D4C3B2A1908F7E6D5C4B3A291807F6E5D4C3B2A1" >> keys.txt

# 4. keys.txt를 OpenSSL을 이용해 keys.enc로 암호화
# (CBC 모드 사용, 비밀번호는 password 파일 참조)
openssl enc -aes-256-cbc -md sha1 -kfile password -in keys.txt -out keys.enc

# 5. 보안을 위해 평문 키 파일 삭제 (실무 습관)
rm keys.txt
```

4. DB 실행

```
-- 1. 데이터베이스 선택
USE demoDB;

-- 2. 테스트 테이블 생성 (InnoDB 엔진 필수)
CREATE TABLE CreditCards (
Id INT PRIMARY KEY AUTO_INCREMENT,
OwnerName VARCHAR(50),
CardNumber VARCHAR(20)
);

-- 3. 민감한 데이터 삽입demoDB
INSERT INTO CreditCards (OwnerName, CardNumber) VALUES
('Cheolsu Kim', '1234-5678-9012-3456'),
('Younghee Lee', '9876-5432-1098-7654');

-- 4. 데이터가 잘 들어갔나 확인 (투명성 확인)
SELECT \* FROM CreditCards;

-- 5. 암호화 적용여부 확인
SELECT `NAME`, ENCRYPTION_SCHEME, ROTATING_OR_FLUSHING
FROM information_schema.INNODB_TABLESPACES_ENCRYPTION
WHERE `NAME` LIKE 'demoDB/CreditCards';
```

결과:

```
{
	"table": "INNODB_TABLESPACES_ENCRYPTION",
	"rows":
	[
		{
			"NAME": "demodb/CreditCards",
			"ENCRYPTION_SCHEME": 1, -- 암호화되었다는 의미이다.
			"ROTATING_OR_FLUSHING": 0
		}
	]
}
```

- 여기서 ibd 파일은 뭔가요?

```
직관적인 비유 (엑셀)
MariaDB 프로그램: 엑셀(Excel) 소프트웨어 그 자체입니다.

테이블(Table): 엑셀 안에 있는 **'시트(Sheet)'**입니다.

.ibd 파일: 내 컴퓨터 하드디스크에 저장된 2024년_가계부.xlsx 파일입니다.

우리가 엑셀을 끄면 화면에서는 사라지지만 하드디스크에 .xlsx 파일은 남아있죠? 마찬가지로 DB를 꺼도 데이터가 사라지지 않는 이유는 이 .ibd 파일에 기록되어 있기 때문입니다.
```

## 만약에 내가 특정한 테이블 데이터만 암호화 하고 싶다면?

```
-- 현재 상태 확인 (아마 FORCE로 되어 있을 겁니다)
SHOW VARIABLES LIKE 'innodb_encrypt_tables';

-- 암호화 정책을 'FORCE'에서 'ON'으로 변경
-- ON: 기본적으로 암호화하되, 원하면 뺄 수 있음 (Opt-out 가능)
SET GLOBAL innodb_encrypt_tables = 'ON';

* 또는 innodb_file_per_table=ON 를 my.cnf에 지정 (대신에 innodb_encrypt_tables = FORCE 주석처리)

```

USE demoDB;

-- 1. 암호화 테이블 (기본값 or 명시적 지정)
CREATE TABLE TSecure (
Msg VARCHAR(100)
) ENGINE=InnoDB ENCRYPTED='YES';

-- 2. 일반(평문) 테이블 (ENCRYPTED='N' 옵션 사용)
CREATE TABLE TPlain (
Msg VARCHAR(100)
) ENGINE=InnoDB;

INSERT INTO TSecure VALUES ('MySecretPassword_SECURE');
INSERT INTO TPlain VALUES ('MySecretPassword_PLAIN');

```
ALTER TABLE TPlain ENCRYPTED='YES';

⚠️ 주의사항 (오버헤드 발생)
이 과정은 테이블의 모든 데이터를 읽고 쓰는 I/O 집약적인 작업입니다.

시간 소요: 테이블 크기가 클수록 완료되는 데 시간이 오래 걸립니다.

디스크 공간: 일시적으로 기존 파일과 새 파일이 모두 존재해야 하므로, 테이블 크기의 약 2배에 해당하는 디스크 공간이 필요합니다.

락(Lock): ALTER TABLE이 실행되는 동안 해당 테이블에는 **배타적 락(Exclusive Lock)**이 걸려 테이블을 사용할 수 없는 다운타임이 발생할 수 있습니다.

따라서 실제 운영 환경에서 대용량 테이블에 암호화를 적용할 때는 서비스 영향도를 최소화하기 위한 별도의 전략(예: 온라인 DDL 툴 사용 등)을 고려해야 합니다.

* 이미 암호화된 테이블을 복호화 하는 건 단순히 ALTER TABLE로는 불가하다 .
* 그런데 비 암호화된 테이블을 암호화 하는건
ALTER TABLE {테이블이름} ENCRYPTED = 'YES'; 로도 충분히 가능하다.

```
