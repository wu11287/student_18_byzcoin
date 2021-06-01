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

type node struct {
	id     		int
	weight 		int
	pk     		crypto.VrfPubkey
	sk     		crypto.VrfPrivkey
	rnd    		crypto.VrfOutput
	proof  		crypto.VrfProof
	pkList 		[4]crypto.VrfPubkey
	proofList 	[4]crypto.VrfProof
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
	// fmt.Println("newNode", i, nNode.pk, nNode.sk)
}

//每个节点把自己的公钥、id 写到管道
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
	fmt.Println()
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


func storeRnd(ch chan *RndAndId, nNode *[4]node, id int, deferFunc func(), randomness []byte) {
	defer func(){
		deferFunc()
	}()

	for i:=0; i < 4; i++ {
		tmp := <- ch
		// 如果满足条件，就验证后加入idlist
		if tmp.in {
			go verifyRnd(nNode[id].pkList[tmp.id], nNode[id].proofList[tmp.id], tmp.rnd, randomness, id, nNode, tmp.id)
		}
	}
}

// 验证rnd是否正确，判断节点是否说谎
func verifyRnd(pk crypto.VrfPubkey, proof crypto.VrfProof, output crypto.VrfOutput,  msg []byte, id int, nNode *[4]node, tmpId int) {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	ok, output2 := pk.VerifyMy(proof, msg)
	if !ok {
		fmt.Println("verified error")
		return
	}
	if output == output2 {
		nNode[id].idList = append(nNode[id].idList, tmpId)
	}
}


func main() {
	var wg sync.WaitGroup
	// chsPK := make([]chan crypto.VrfPubkey, 4) //四个节点的管道切片,管道传送公钥
	chsPK := make([]chan *pkAndId, 4) 
	chsProof := make([]chan *ProofAndId, 4)
	chsRnd := make([]chan *RndAndId, 4)
	nodes := [4]node{}
	rand.Seed(time.Now().Unix())
	randomness, err := GenerateRandomBytes(10)

	if err != nil {
		fmt.Println("generate random byte[] failed!")
	}
	//生成所有节点的公私钥
	for i := 0; i < 4; i++ {
		chsPK[i] = make(chan *pkAndId)
		chsProof[i] = make(chan *ProofAndId)
		chsRnd[i] = make(chan *RndAndId)
		wg.Add(1)
		go newNode(&nodes[i], i)
	}

	time.Sleep(time.Second * 1)

	// 广播公钥
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go broadcastPK(chsPK, &nodes[i], i, wg.Done)
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go storePk(chsPK[i], &nodes, i, wg.Done)
	}
	
	time.Sleep(time.Second * 1) //使得程序运行到此的时候已经广播公钥完毕

	//加密抽签
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go Sotition(&nodes[i], randomness)
	}

	//实际上会在加密抽签后，先判断自己是不是被选中，被选中就打包区块。并将rnd和proof随着区块一起广播
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
		wg.Add(1)
		go storeProof(chsProof[i], &nodes, i, wg.Done)
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go storeRnd(chsRnd[i], &nodes, i, wg.Done, randomness)
	}
	fmt.Println("now")

	for _, v := range nodes[0].idList {
		fmt.Println(v)
	}


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
