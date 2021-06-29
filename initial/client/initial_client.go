package main

import (
	"fmt"
	"log"
	initial "myProject/protos"
	"net"

	pt "myProject/initial"

	"google.golang.org/grpc"
)

//TODO 后续读取环境变量
const id int = 1

// const m int = 2

//广播ip
const (
	address = "172.23.255.255:50051"
)



func ToShard() {

}

// 作为一个客户端的角色, 去dial广播地址即可
// 每个节点同时也需要作为服务端在对应端口 8888 监听 --- 如何实现？
func main() {
	// var wg sync.WaitGroup
	node := pt.NewNode(id)
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
	pt.RunBroadPk(client, node)

	// randomness, err := GenerateRandomBytes(10) //不用seed会产生确定性结果
	// if err != nil {
	// 	log.Fatalf("generate randomness error: %v", err)
	// }
	// // sortition
	// Sotition(node, randomness, &wg)
	// if node.choosed {
	// 	runBroadProof(client, node, randomness, ip.String(), id)
	// }

	// 对IpInShard分片
}
