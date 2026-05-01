## os 모듈
읽어야할 것 실제 go code,system.md , summray.md, claude.md
### 전제조건
본 spec은 물리 os 의 점검을 담당한다.
절대 기존의 다른 점검 코드 쪽의 code를 수정/삭제/추가해선 안된다.
기본적인 점검 방법과 flow는 네트워크와 동일하다.
같은 프로젝트 안에있지만 k8s job을 통해서 별도의 args 태그로 담당하는 기능만 수행이 가능행해야한다.
네트워크와 동일하게 데몬셋 형식으로 각각의 노드에서 실행되어야한다.
summary.md의 기술된 내용과 변경점이 몇몇 존재하고 최신화 되지 못했다.
network 점검도구가 실제로 사용되고 있으며 즉 markdown 보다 실제 internal 내부의 code가 신뢰성이 높다.
네트워크와 동일하게 점검후 sse pod로 user-id / sse-type / data 로 이루어진 값을 보내야한다.
아마 기존에는 network 폴더에서 훨씬 많은 key를 보내고 있을 텐데 실질적으로 sse 에 서 받아볼 필요가 있는 로그는
위 3개로 이루어지면된다.
user-id = 1 (관리자의 id)
sse-type = os
data = 실제 관리자 웹에서 받아볼 시스템 점검 결과, 즉 본 기능에서 확인하고자 하는것.


### 기능 

CPU사용률/ Load AverageMemory사용률/ Disk파티션별 사용률/ inode 사용률 / Process좀비 프로세스, 비정상 프로세스
위 4가지 점검이 핵심이다.
위 4가지 점검 사항이 각각 노드별로 즉 x3 배수가 나올것이고 해당 점검 결과 실제로 받아볼 유의미한 데이터를 data에 실어서
x.sse.fanout 익스체인지로 각 단계가 끝날때마다 publish 하면된다.
network.md에서 기술된 출력 양식을 동일하게 활용한다.