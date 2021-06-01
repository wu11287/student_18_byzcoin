package main

import (
	// "bytes"
	// "encoding/binary"
	"fmt"
	"math/rand"
	crypto "myProject/crypto"
	"sync"
	"time"
)

// const denominator float64 = 18446744073709551615

type node struct {
	id     int
	weight int
	pk     crypto.VrfPubkey
	sk     crypto.VrfPrivkey
	rnd    crypto.VrfOutput
	proof  crypto.VrfProof
	pkList []crypto.VrfPubkey
	// ch chan crypto.VrfPubkey //管道
}

//生成长度为 n 的随机byte数组
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Sotition(nNode *node, msg []byte) {
	proof, ok1 := nNode.sk.ProveMy(msg)
	if !ok1 {
		fmt.Println("generate proof error!")
	}

	rnd, ok2 := proof.Hash()
	if !ok2 {
		fmt.Println("generate rnd error!")
	}
	nNode.rnd = rnd
	nNode.proof = proof
}

func newNode(nNode *node, i int) {
	nNode.id = i
	nNode.weight = i
	nNode.pk, nNode.sk = crypto.VrfKeygen()
}

//每个节点把自己的公钥写到所有节点管道中 -- 此时公钥是已经生成的
func broadcastPK(ch []chan crypto.VrfPubkey, nNode *node, i int, deferFunc func()) {
	defer func() {
		deferFunc()
	}()
	// for i:= 0; i < 4; i++ {
	// 	select {
	// 	case ch[0] <- nNode.pk:
	// 		fmt.Printf("node %d sendPK to ch[0] success!", i)
	// 	case ch[1] <- nNode.pk:
	// 		fmt.Printf("node %d sendPK to ch[1] success!", i)
	// 	case ch[2] <- nNode.pk:
	// 		fmt.Printf("node %d sendPK to ch[2] success!", i)
	// 	case ch[3] <- nNode.pk:
	// 		fmt.Printf("node %d sendPK to ch[3] success!", i)
	// 	default:
	// 		fmt.Printf("broadcast pk error\n")
	// 	}
	// }
	// for i := 0; i < 4; i++ {
	// 	ch[i] <- nNode.pk //数据存在管道里
	// 	close(ch[i])
	// }
	ch[0] <- nNode.pk
	ch[1] <- nNode.pk
	ch[2] <- nNode.pk
	ch[3] <- nNode.pk
}

// func broadcastProof(ch []chan interface{}, nNode *node, i int, deferFunc func()) {
// 	defer func(){
// 		deferFunc()
// 	}()
// 	for {
// 		select {
// 		case ch[0] <- nNode.proof:
// 			fmt.Printf("node %d sendProof to ch[0] success!", i)
// 		case ch[1] <- nNode.pk:
// 			fmt.Printf("node %d send ch[1] success!", i)
// 		case ch[2] <- nNode.pk:
// 			fmt.Printf("node %d send ch[2] success!", i)
// 		case ch[3] <- nNode.pk:
// 			fmt.Printf("node %d send ch[3] success!", i)
// 		default:
// 			fmt.Printf("broadcast pk error")
// 		}
// 	}
// }

func main() {
	var wg sync.WaitGroup
	chs := make([]chan crypto.VrfPubkey, 4) //四个节点的管道切片,管道传送公钥
	nodes := [4]node{}
	rand.Seed(time.Now().Unix()) //以当前时间，更新随机种子
	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		fmt.Println("generate random byte[] failed!")
	}
	fmt.Printf("%#v\n", randomness)
	//生成所有节点的公私钥
	for i := 0; i < 4; i++ {
		chs[i] = make(chan crypto.VrfPubkey) //实例化这个管道
		wg.Add(1)
		newNode(&nodes[i], i)
	}

	// 广播公钥
	for i := 0; i < 4; i++ {
		broadcastPK(chs, &nodes[i], i, wg.Done)
	}

	//把每个管道中的公钥，持久化到每个节点的pklist中
	for i := 0; i < 4; i++ {
		val := <- chs[0]
		nodes[0].pkList = append(nodes[0].pkList, val)
	}
	for i := 0; i < 4; i++ {
		val := <- chs[1]
		nodes[1].pkList = append(nodes[1].pkList, val)
	}
	for i := 0; i < 4; i++ {
		val := <- chs[2]
		nodes[2].pkList = append(nodes[2].pkList, val)
	}
	for i := 0; i < 4; i++ {
		val := <- chs[3]
		nodes[3].pkList = append(nodes[3].pkList, val)
	}
	// for val := range chs[1] {
	// 	<-chs[1]
	// 	nodes[1].pkList = append(nodes[1].pkList, val)
	// }

	wg.Wait()

	for _, val := range nodes[0].pkList {
		fmt.Print(val, " ")
	}
	for _, val := range nodes[1].pkList {
		fmt.Print(val, " ")
	}

	// //加密抽签
	// for i := 0; i < 4; i++ {
	// 	go Sotition(&nodes[i], randomness)
	// }
	// //广播Proof
	// for i := 0; i < 4; i++ {
	// 	broadcastProof(chs, &nodes[i], i, wg.Done)
	// }

	// identified := make([]int, 0)
	// for i := 0; i < 4; i++ {
	// 	ok := isMeet(&nodes[i])
	// 	if !ok {
	// 		continue
	// 	}
	// 	everyId := nodes[i].id
	// 	identified = append(identified, everyId)
	// }

	// for _, id := range identified {
	// 	fmt.Println(id)
	// }
}

// func isMeet(nNode *node) bool {
// 	bytesBuffer := bytes.NewBuffer(nNode.rnd[:])
// 	var x int64
// 	binary.Read(bytesBuffer, binary.BigEndian, &x)

//   	rnd := float64(x)
// 	if rnd < 0 {
// 		rnd += denominator
// 	}

// 	p := rnd/denominator
// 	return p < 0.7
// }
