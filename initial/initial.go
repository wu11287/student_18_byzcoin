package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	crypto "myProject/crypto"
	"net"
	"sync"
	"time"

	initial "myProject/protos"

	"google.golang.org/grpc"
)

//TODO 后续读取环境变量
const id int = 1
const denominator float64 = 18446744073709551615
const m int = 2

//广播ip
const (
	address = "10.112.255.255:50051"
)

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

func newNode(i int) *Node {
	pk, sk := crypto.VrfKeygen()
	return &Node{
		Id:     i,
		Weight: rand.Intn(3) + 1,
		Pk:     pk,
		Sk:     sk,
	}
}

type MyServer struct {
	initial.UnimplementedBroadAllServer
	node       *Node
	randomness []byte
}

// 作为server处理其他client发过来的消息
func (s *MyServer) BroadPK(stream initial.BroadAll_BroadPKServer) error {
	waitc := make(chan struct{})
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			close(waitc)
			return nil
		}
		if err != nil {
			log.Fatalf("Failed to receive msg : %v", err)
			return err
		}
		var data []byte = []byte(in.Pk)
		var data2 crypto.VrfPubkey
		copy(data2[:32], data)
		s.node.PkList[in.Id] = data2
	}
}

func (s *MyServer) BroadProof(stream initial.BroadAll_BroadProofServer) error {
	waitc := make(chan struct{})
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// read done.
			close(waitc)
			return nil
		}
		if err != nil {
			log.Fatalf("Failed to receive ProofAndId : %v", err)
		}
		log.Printf("Got message %s, %d", in.Proof, in.Id)

		//verify proof
		var data []byte = []byte(in.Proof)
		var proof crypto.VrfProof
		copy(proof[:], data)
		ok := verifyProof(s.node.PkList[in.Id], proof, s.randomness)
		if !ok {
			log.Fatalf("verify proof error")
			return err
		}

		tmp := IPAndId{Ip: in.Ip, Id: int(in.Id)}
		s.node.IpInShard = append(s.node.IpInShard, tmp)
	}
}

func runBroadPk(client initial.BroadAllClient, node *Node) {
	msgs := []*initial.PkAndId{
		{Pk: string(node.Pk[:]), Id: int32(node.Id)}, //[32]byte 转换成string
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.BroadPK(ctx)
	if err != nil {
		log.Fatalf("%v.BroadPK(_) = _, %v", client, err)
	}

	for _, msg := range msgs {
		if err := stream.Send(msg); err != nil {
			log.Fatalf("Failed to sent msg: %v", err)
		}
	}
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
func verifyProof(Pk crypto.VrfPubkey, proof crypto.VrfProof, msg []byte) bool {
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

func ToShard() {

}

// 作为一个客户端的角色, 去dial广播地址即可
// 每个节点同时也需要作为服务端在对应端口 8888 监听 --- 如何实现？
func main() {
	var wg sync.WaitGroup

	node := newNode(id)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	ip := addrs[1]
	fmt.Println(ip) //10.112 网段

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal("create conn err :", err)
	}
	defer conn.Close()

	client := initial.NewBroadAllClient(conn)
	runBroadPk(client, node)

	randomness, err := GenerateRandomBytes(10) //不用seed会产生确定性结果
	if err != nil {
		log.Fatalf("generate randomness error: %v", err)
	}
	// sortition
	Sotition(node, randomness, &wg)
	if node.choosed {
		runBroadProof(client, node, randomness, ip.String(), id)
	}

	// 对IpInShard分片
}
