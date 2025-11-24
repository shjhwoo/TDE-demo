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

CREATE DATABASE IF NOT EXISTS demoDB;
USE demoDB;

-- 2. 테스트 테이블 생성
CREATE TABLE CreditCards (
Id INT PRIMARY KEY AUTO_INCREMENT,
OwnerName VARCHAR(50),
CardNumber VARCHAR(20)
);

-- 3. 민감한 데이터 삽입 
INSERT INTO CreditCards (OwnerName, CardNumber) VALUES
('Cheolsu Kim', '1234-5678-9012-3456'),
('Younghee Lee', '9876-5432-1098-7654');

-- 4. 테이블 암호화
ALTER TABLE CreditCards ENCRYPTED='YES';

-- 5. 데이터가 잘 들어갔나 확인 (투명성 확인, 즉 DB 접근 권한이 있는 사용자는 암호화 여부를 몰라도 그냥 바로 SELECT UPDATE 등의 쿼리를 수행 가능)
SELECT \* FROM CreditCards;

-- 6. 암호화 적용여부 확인
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

1. ALTER TABLE {테이블이름} ENCRYPTED='YES';
2.  또는 innodb_file_per_table=ON 를 my.cnf에 지정 (대신에 innodb_encrypt_tables = FORCE 주석처리)
