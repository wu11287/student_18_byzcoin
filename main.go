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
	rndList		[4]map[int]float64 //key是id，value是概率p
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
func broadcastPK(ch []chan crypto.VrfPubkey, nNode *node, deferFunc func()) {
	defer func() {
		deferFunc()
	}()
	ch[0] <- nNode.pk
	ch[1] <- nNode.pk
	ch[2] <- nNode.pk
	ch[3] <- nNode.pk
}


func broadcastProof(ch []chan crypto.VrfProof, nNode *node, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	ch[0] <- nNode.proof
	ch[1] <- nNode.proof
	ch[2] <- nNode.proof
	ch[3] <- nNode.proof
}

func broadcastRnd(ch []chan float64, nNode *node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	p := isMeet(nNode)
	fmt.Printf("broadcastRnd id = %d, p = %f", id, p)
	if p > 0.7 {
		return //表示没有被选中
	}
	ch[0] <- p
	ch[1] <- p
	ch[2] <- p
	ch[3] <- p
}

//持久化公钥到节点struct
func storePk(ch chan crypto.VrfPubkey, nNode *[4]node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	for i := 0; i < 4; i++ {
		val0 := <- ch
		// nNode[id].pkList = append(nNode[id].pkList, val0)
		nNode[id].pkList[id] = val0
	}
}

func storeProof(ch chan crypto.VrfProof, nNode *[4]node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	for {
		val0 := <- ch
		// nNode[id].proofList = append(nNode[id].proofList, val0)
		nNode[id].proofList[id] = val0
	}
}

func storeRnd(ch chan float64, nNode *[4]node, id int, deferFunc func()) {
	defer func(){
		deferFunc()
	}()
	for i := 0; i < 4; i++ {
		val0 := <- ch
		// nNode[id].proofList = append(nNode[id].proofList, val0)
		nNode[id].rndList[id] = val0
	}
}

// 验证rnd是否正确，判断节点是否说谎
func verifyRnd() {

}


func main() {
	var wg sync.WaitGroup
	chsPK := make([]chan crypto.VrfPubkey, 4) //四个节点的管道切片,管道传送公钥
	chsProof := make([]chan crypto.VrfProof, 4)
	chsRnd := make([]chan float64, 4)
	nodes := [4]node{}
	rand.Seed(time.Now().Unix()) //以当前时间，更新随机种子
	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		fmt.Println("generate random byte[] failed!")
	}
	//生成所有节点的公私钥
	for i := 0; i < 4; i++ {
		chsPK[i] = make(chan crypto.VrfPubkey) //实例化这个管道
		chsProof[i] = make(chan crypto.VrfProof)
		chsRnd[i] = make(chan float64)
		wg.Add(1)
		go newNode(&nodes[i], i)
	}

	// 广播公钥
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go broadcastPK(chsPK, &nodes[i], wg.Done)
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
		go broadcastProof(chsProof, &nodes[i], wg.Done)
		go broadcastRnd(chsRnd, &nodes[i], i, wg.Done)
	}


	for i := 0; i < 4; i++ {
		wg.Add(2)
		go storeProof(chsProof[i], &nodes, i, wg.Done)
		go storeRnd(chsRnd[i], &nodes, i, wg.Done)
	}

	fmt.Println("now")

	for i, v := range nodes[0].rndList {
		if v == 0.0 {
			continue
		}
		fmt.Printf("node %d have probability %f\n", i, v)
	}
	wg.Wait()

	


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
