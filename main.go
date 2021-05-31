package main

import (
	"fmt"
	"math/rand"
	crypto "myProject/crypto"
	"time"
	"bytes"
	"encoding/binary"
)

const denominator float64 = 18446744073709551615


type node struct {
	id int
	weight int
	pk crypto.VrfPubkey
	sk crypto.VrfPrivkey
	rnd crypto.VrfOutput
	proof crypto.VrfProof
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

//每个节点把自己的公钥写到所有节点管道中
func broadcastPK(ch []chan interface{}, nNode *node, i int) {
	for {
		select {
		case ch[0] <- nNode.pk:
			fmt.Printf("node %d sendPK to ch[0] success!", i)
		case ch[1] <- nNode.pk:
			fmt.Printf("node %d sendPK to ch[1] success!", i)
		case ch[2] <- nNode.pk:
			fmt.Printf("node %d sendPK to ch[2] success!", i)
		case ch[3] <- nNode.pk:
			fmt.Printf("node %d sendPK to ch[3] success!", i)
		default:
			fmt.Printf("broadcast pk error")
		}
	}
}

func broadcastProof(ch []chan interface{}, nNode *node, i int) {
	for {
		select {
		case ch[0] <- nNode.proof:
			fmt.Printf("node %d sendProof to ch[0] success!", i)
		case ch[1] <- nNode.pk:
			fmt.Printf("node %d send ch[1] success!", i)
		case ch[2] <- nNode.pk:
			fmt.Printf("node %d send ch[2] success!", i)
		case ch[3] <- nNode.pk:
			fmt.Printf("node %d send ch[3] success!", i)
		default:
			fmt.Printf("broadcast pk error")
		}
	}
}

func main() {
	chs := make([]chan interface{}, 4)
	nodes := [4]node{}
	rand.Seed(time.Now().Unix()) //以当前时间，更新随机种子
	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		fmt.Println("generate random byte[] failed!")
	}
	//生成所有节点的公私钥
	for i := 0; i < 4; i++ {
		chs[i] = make(chan interface{})
		newNode(&nodes[i], i)  
	}

	// 广播公钥
	for i := 0; i < 4; i++ {
		broadcastPK(chs, &nodes[i], i)
	}
	//加密抽签
	for i := 0; i < 4; i++ {
		go Sotition(&nodes[i], randomness)
	}
	//广播Proof
	for i := 0; i < 4; i++ {
		broadcastProof(chs, &nodes[i], i)
	}

	identified := make([]int, 0)
	for i := 0; i < 4; i++ {
		ok := isMeet(&nodes[i])
		if !ok {
			continue
		}
		everyId := nodes[i].id
		identified = append(identified, everyId)
	}

	for _, id := range identified {
		fmt.Println(id)
	}
}

func isMeet(nNode *node) bool {
	bytesBuffer := bytes.NewBuffer(nNode.rnd[:])
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)

  	rnd := float64(x)
	if rnd < 0 {
		rnd += denominator
	}
	
	p := rnd/denominator
	return p < 0.7
}