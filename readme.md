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
