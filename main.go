package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	crypto "myProject/crypto"
	"sync"
	"time"
)

const denominator float64 = 18446744073709551615

var randomness []byte

type node struct {
	id     		int
	weight 		int
	pk     		crypto.VrfPubkey
	sk     		crypto.VrfPrivkey
	rnd    		crypto.VrfOutput
	proof  		crypto.VrfProof
	pkList 		[4]crypto.VrfPubkey
	proofList 	[4]crypto.VrfProof
	// rndList		[4]float64 //存放所有满足条件的节点的rnd ； rnd不需要记录
	idList 		[]int
}


type pkAndId struct {
	id int
	pk crypto.VrfPubkey
}

type ProofAndId struct {
	id int
	proof crypto.VrfProof
}

type RndAndId struct {
	id int
	rnd crypto.VrfOutput
	p float64
	in bool
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
func broadcastPK(ch []chan *pkAndId, nNode *node, id int, deferFunc func()) {
	defer func() {
		deferFunc()
	}()
	tmp := &pkAndId{id, nNode.pk}
	ch[0] <- tmp
	ch[1] <- tmp
	ch[2] <- tmp
	ch[3] <- tmp
}

// 广播特定节点的proof
func broadcastProof(ch []chan *ProofAndId, nNode *node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	tmp := &ProofAndId{id, nNode.proof}
	ch[0] <- tmp
	ch[1] <- tmp
	ch[2] <- tmp
	ch[3] <- tmp
}

func broadcastRnd(ch []chan *RndAndId, nNode *node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	p := isMeet(nNode)
	fmt.Printf("broadcastRnd id = %d, p = %f", id, p)
	var isin bool
	if p > 0.7 {
		isin = false
	} else {
		isin = true
	}
	tmp := &RndAndId{id, nNode.rnd, p, isin}
	
	ch[0] <- tmp
	ch[1] <- tmp
	ch[2] <- tmp
	ch[3] <- tmp
}

//持久化公钥到节点struct
func storePk(ch chan *pkAndId, nNode *[4]node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	for {
		tmp := <- ch
		nNode[id].pkList[tmp.id] = tmp.pk
	}
}

func storeProof(ch chan *ProofAndId, nNode *[4]node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()

	for {
		tmp := <- ch
		nNode[id].proofList[tmp.id] = tmp.proof
	}
}

// 对于不满足要求的，p、rnd都是0
func storeRnd(ch chan *RndAndId, nNode *[4]node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	for {
		tmp := <- ch
		// 验证是否该节点真的满足条件
		if tmp.in {
			ok := verifyRnd(nNode[id].pkList[tmp.id], nNode[id].proofList[tmp.id], tmp.rnd, randomness)
			if !ok {
				fmt.Println("the node bad!")
				continue
			}
			nNode[id].idList = append(nNode[id].idList, tmp.id)
		}
	}
}

// 验证rnd是否正确，判断节点是否说谎
func verifyRnd(pk crypto.VrfPubkey, proof crypto.VrfProof, output crypto.VrfOutput,  msg []byte) bool {
	ok, output2 := pk.VerifyMy(proof, msg)
	if !ok {
		fmt.Println("verified error")
		return false
	}
	if output == output2 {
		return true
	}
	return false
}


func main() {
	var wg sync.WaitGroup
	// chsPK := make([]chan crypto.VrfPubkey, 4) //四个节点的管道切片,管道传送公钥
	chsPK := make([]chan *pkAndId, 4) 
	chsProof := make([]chan *ProofAndId, 4)
	chsRnd := make([]chan *RndAndId, 4)
	nodes := [4]node{}
	rand.Seed(time.Now().Unix()) //以当前时间，更新随机种子
	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		fmt.Println("generate random byte[] failed!")
	}
	//生成所有节点的公私钥
	for i := 0; i < 4; i++ {
		// chsPK[i] = make(chan crypto.VrfPubkey)
		chsPK[i] = make(chan *pkAndId)
		chsProof[i] = make(chan *ProofAndId)
		chsRnd[i] = make(chan *RndAndId)
		wg.Add(1)
		go newNode(&nodes[i], i)
	}

	// 广播公钥
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go broadcastPK(chsPK, &nodes[i], i, wg.Done)
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go storePk(chsPK[i], &nodes, i, wg.Done)
	}
	
	//加密抽签
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go Sotition(&nodes[i], randomness)
	}
	//广播Proof、rnd
	for i := 0; i < 4; i++ {
		proof, ok1 := nodes[i].sk.ProveMy(randomness)
		if !ok1 {
			fmt.Println("proof generated error")
		}
		rnd, ok2 := proof.Hash()
		if !ok2 {
			fmt.Println("output generated error")
		}
		nodes[i].proof = proof
		nodes[i].rnd = rnd
		wg.Add(2)
		go broadcastProof(chsProof, &nodes[i], i, wg.Done)
		go broadcastRnd(chsRnd, &nodes[i], i, wg.Done)
	}


	for i := 0; i < 4; i++ {
		wg.Add(2)
		go storeProof(chsProof[i], &nodes, i, wg.Done)
		go storeRnd(chsRnd[i], &nodes, i, wg.Done)
	}

	fmt.Println("now")

	wg.Wait()
}

func isMeet(nNode *node) float64 {
	bytesBuffer := bytes.NewBuffer(nNode.rnd[:])
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)

  	rnd := float64(x)
	if rnd < 0 {
		rnd += denominator
	}

	p := rnd/denominator
	return p
}
