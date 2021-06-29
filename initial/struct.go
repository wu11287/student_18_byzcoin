package initial

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	crypto "myProject/crypto"
	"sync"
	"time"

	initial "myProject/protos"
)

const denominator float64 = 18446744073709551615

var lock sync.Mutex

type Node struct {
	Id        int
	Weight    int
	Pk        crypto.VrfPubkey
	Sk        crypto.VrfPrivkey
	PkList    []crypto.VrfPubkey
	Rnd       crypto.VrfOutput
	Proof     crypto.VrfProof
	choosed   bool
	IpInShard []IPAndId
}


type IPAndId struct {
	Ip string
	Id int
}

type PkAndId struct {
	Id int
	Pk crypto.VrfPubkey
}

type ProofAndId struct {
	Id    int
	Proof crypto.VrfProof
}

func NewNode(i int) *Node {
	pk, sk := crypto.VrfKeygen()
	return &Node{
		Id:     i,
		Weight: rand.Intn(3) + 1,
		Pk:     pk,
		Sk:     sk,
	}
}


// 客户端其实没必要用流，因为只发送一次，应该设置成 服务端单向
func RunBroadPk(client initial.BroadAllClient, node *Node) {
	msgs := []*initial.PkAndId{
		{Pk: string(node.Pk[:]), Id: int32(node.Id)}, //[32]byte 转换成string
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.BroadPK(ctx)
	if err != nil {
		log.Fatalf("%v.BroadPK(_) = _, %v", client, err)
	}

	waitc := make(chan struct{})
	for _, msg := range msgs {
		if err := stream.Send(msg); err != nil {
			log.Fatalf("Failed to sent msg: %v", err)
		}
	}
	<- waitc
}

func runBroadProof(client initial.BroadAllClient, node *Node, randomness []byte, ip string, id int) {
	msgs := []*initial.ProofMsg{
		{Proof: string(node.Proof[:]), Id: int32(node.Id), Ip: ip},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.BroadProof(ctx)
	if err != nil {
		log.Fatalf("%v.BroadProof(_) = _, %v", client, err)
	}

	for _, msg := range msgs {
		if err := stream.Send(msg); err != nil {
			log.Fatalf("Failed to sent msg: %v", err)
		}
	}
}

// 传播的时候没有传播rnd
func VerifyProof(Pk crypto.VrfPubkey, proof crypto.VrfProof, msg []byte) bool {
	ok, _ := Pk.VerifyMy(proof, msg)

	if !ok {
		fmt.Println("verified error")
		return false
	}
	return ok
}

// msg 是全局一致的随机值
func Sotition(node *Node, msg []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	lock.Lock()

	proof, ok := node.Sk.ProveMy(msg)
	if !ok {
		log.Fatal("generate proof error")
	}

	rnd, ok := proof.Hash()
	if !ok {
		log.Fatal("generate rnd error")
	}
	node.Rnd = rnd
	node.Proof = proof

	for i := 0; i < node.Weight; i++ {
		ok = VerifyRnd(node)
		if ok {
			node.choosed = true
			break
		}
	}
}

func VerifyRnd(node *Node) bool {
	bytesBuffer := bytes.NewBuffer(node.Rnd[:])
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	rnd := float64(x)
	if rnd < 0 {
		rnd += denominator
	}

	p := rnd / denominator //得到一个概率值

	if p < 0.7 {
		return true
	}

	return false
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
