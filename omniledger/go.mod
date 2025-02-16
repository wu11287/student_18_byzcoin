module student_18_byzcoin/omniledger

go 1.16

replace (
	github.com/coreos/bbolt v1.3.5 => go.etcd.io/bbolt v1.3.5
	github.com/dedis/fixbuf v1.0.3 => go.dedis.ch/fixbuf v1.0.3
	github.com/dedis/protobuf v1.0.11 => go.dedis.ch/protobuf v1.0.11
	go.dedis.ch/fixbuf v1.0.3 => github.com/dedis/fixbuf v1.0.3
	go.etcd.io/bbolt v1.3.5 => github.com/coreos/bbolt v1.3.5
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/coreos/bbolt v1.3.5
	github.com/daviddengcn/go-colortext v1.0.0 // indirect
	github.com/dedis/fixbuf v1.0.3 // indirect
	github.com/dedis/protobuf v1.0.11
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/prataprc/goparsec v0.0.0-20210210111032-e54bde2749dd
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.3.0
	go.etcd.io/bbolt v1.3.5 // indirect
	golang.org/x/net v0.0.0-20210505214959-0714010a04ed // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/dedis/cothority.v2 v2.0.0-20180329140330-3dbb49f06ce1
	gopkg.in/dedis/kyber.v2 v2.0.0-20180509082236-f066f8d2cd58
	gopkg.in/dedis/onet.v2 v2.0.0-20181115163211-c8f3724038a7
	gopkg.in/satori/go.uuid.v1 v1.2.0
	gopkg.in/tylerb/graceful.v1 v1.2.15 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
	rsc.io/goversion v1.2.0 // indirect
)
