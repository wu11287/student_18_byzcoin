package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	crypto "myProject/crypto"
	initial "myProject/protos"
	"net"

	"google.golang.org/grpc"

	st "myProject/initial"
)


var (
	port = flag.Int("port", 50051, "The server port")
)


type MyServer struct {
	initial.UnimplementedBroadAllServer
	node       *st.Node
	randomness []byte
}


// 作为server处理其他client发过来的消息，但是不回复任何消息
// 因为广播地址是255，所以每个节点实际上和全网所有节点发消息，其他节点收到就存储消息即可
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
		fmt.Printf("server received msg: pk = %v, id = %v\n", in.Pk, in.Id)
		var data []byte = []byte(in.Pk)
		var data2 crypto.VrfPubkey
		copy(data2[:32], data)
		s.node.PkList[in.Id] = data2
	}
}



// 服务端定义的方法，供客户端调用
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
		ok := st.VerifyProof(s.node.PkList[in.Id], proof, s.randomness)
		if !ok {
			log.Fatalf("verify proof error")
			return err
		}

		tmp := st.IPAndId{Ip: in.Ip, Id: int(in.Id)}
		s.node.IpInShard = append(s.node.IpInShard, tmp)
	}
}


func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
	grpcServer := grpc.NewServer()
	initial.RegisterBroadAllServer(grpcServer, &MyServer{})
	grpcServer.Serve(lis)
}