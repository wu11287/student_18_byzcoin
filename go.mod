module myProject

go 1.16

require (
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/algorand/go-algorand v0.0.0-20210527003728-3fa882c7802e
	github.com/algorand/go-deadlock v0.2.1
	github.com/algorand/msgp v1.1.47
	github.com/daviddengcn/go-colortext v1.0.0 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c
	github.com/dedis/fixbuf v1.0.3 // indirect
	github.com/dedis/protobuf v1.0.11 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/shirou/gopsutil v3.21.5+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.6 // indirect
	gopkg.in/dedis/kyber.v2 v2.0.0-20180509082236-f066f8d2cd58 // indirect
	gopkg.in/dedis/onet.v2 v2.0.0-20181115163211-c8f3724038a7
	gopkg.in/satori/go.uuid.v1 v1.2.0 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace (
	github.com/dedis/fixbuf v1.0.3 => go.dedis.ch/fixbuf v1.0.3
	github.com/dedis/protobuf v1.0.11 => go.dedis.ch/protobuf v1.0.11
	go.dedis.ch/fixbuf v1.0.3 => github.com/dedis/fixbuf v1.0.3
)
