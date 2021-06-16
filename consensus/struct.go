package consensus

import (
	crypto "myProject/crypto"
	"gopkg.in/dedis/onet.v2/network"
	initial "myProject/initial"
	
)

type ServerToml struct {
	Address 	network.Address
	id 			int
	Public		crypto.VrfPubkey
	URL 		string //http://127.0.0.1 类似这种
	Description	string
}


type shardToml struct {
	Servers []*ServerToml
}


/*
type ServerIdentity struct {
	Public kyber.Point
	ID ServerIdentityID
	Address Address
	Description string
	private kyber.Scalar
	URL string
}
*/
type shard struct {
	Id 	int
	List []*initial.node
}