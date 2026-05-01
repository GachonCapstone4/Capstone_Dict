## 추가사항 

기존 VPN 검증은 다음과 같을 것이다.

L3 터널링	pfSense/에이전트	AWS VPN 게이트웨이

wg show,ping -c 5

경로 검증	웹 서버	AWS 추론 서버

traceroute (터널 통과 여부 확인)

성능 품질	에이전트	AWS 추론 서버

iperf3 -c [IP] -u (대역폭 측정)

보안 정책	pfSense/AWS	보안 그룹/ACL

nc -zv [IP] [Port]

여기서 mtr 을 이용한 traceroute 점검 
iperf 3 를 통한 점검 빼고 보안정책밀 터널링 점검, nc 등의 기능을 안전히 삭제하라