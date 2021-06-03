package testmi
// package main

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

var lock sync.Mutex

type node struct {
	id        int
	weight    int
	pk        crypto.VrfPubkey
	sk        crypto.VrfPrivkey
	rnd       crypto.VrfOutput
	proof     crypto.VrfProof
	pkList    [4]crypto.VrfPubkey
	proofList [4]crypto.VrfProof
	idList    []int
}

type pkAndId struct {
	id int
	pk crypto.VrfPubkey
}

type ProofAndId struct {
	id    int
	proof crypto.VrfProof
}

type RndAndId struct {
	id  int
	rnd crypto.VrfOutput
	p   float64
	in  bool
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

func Sotition(nNode *node, msg []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()
	proof, ok := nNode.sk.ProveMy(msg)
	if !ok {
		fmt.Println("generate proof error!")
	}

	rnd, ok := proof.Hash()
	if !ok {
		fmt.Println("generate rnd error!")
	}
	nNode.rnd = rnd
	nNode.proof = proof
	lock.Unlock()
}

func newNode(nNode *node, i int) {
	nNode.id = i
	nNode.weight = i+1
	nNode.pk, nNode.sk = crypto.VrfKeygen()
}

func broadcastPK(ch []chan *pkAndId, nNode *node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()

	tmp := &pkAndId{id, nNode.pk}
	go func() {
		for i:=0; i< 4; i++ {
			ch[i] <- tmp
		}
	}()

	lock.Unlock()
}

// 广播特定节点的proof
func broadcastProof(ch []chan *ProofAndId, nNode *node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()
	tmp := &ProofAndId{id, nNode.proof}

	go func() {
		for i:=0; i < 4; i++ {
			ch[i] <- tmp
		}
	}()

	lock.Unlock()
}

func broadcastRnd(ch []chan *RndAndId, nNode *node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()
	var p float64
	for i := 0; i < nNode.weight; i++ {
		p = isMeet(nNode)
		if p > 0.7 {
			continue
		} else {
			break;
		}
	}
	// p := isMeet(nNode)
	var isin bool

	if p > 0.7 {
		isin = false
	} else {
		isin = true
	}

	tmp := &RndAndId{id, nNode.rnd, p, isin}

	go func() {
		for i:=0; i < 4; i++ {
			ch[i] <- tmp
		}
	}()

	lock.Unlock()
}

func isMeet(nNode *node) float64 {
	bytesBuffer := bytes.NewBuffer(nNode.rnd[:])
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	rnd := float64(x)
	if rnd < 0 {
		rnd += denominator
	}

	p := rnd / denominator
	return p
}

//持久化公钥到节点struct
func storePk(ch chan *pkAndId, nNode *[4]node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 4; i++ {
		tmp := <-ch
		nNode[id].pkList[tmp.id] = tmp.pk
	}
}

func storeProof(ch chan *ProofAndId, nNode *[4]node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i :=0; i < 4; i++ {
		tmp := <-ch
		nNode[id].proofList[tmp.id] = tmp.proof
	}
}

func storeRnd(ch chan *RndAndId, nNode *[4]node, id int, randomness []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 4; i++ {
		tmp := <-ch
		if tmp.in {
			ok := verifyRnd(nNode[id].pkList[tmp.id], nNode[id].proofList[tmp.id], tmp.rnd, randomness, id)
			if ok {
				nNode[id].idList = append(nNode[id].idList, tmp.id)
			}
		}
	}
}

func verifyRnd(pk crypto.VrfPubkey, proof crypto.VrfProof, output crypto.VrfOutput, msg []byte, id int) bool {
	ok, output2 := pk.VerifyMy(proof, msg)

	if !ok {
		fmt.Println("verified error")
		return false
	}
	return output == output2
}

func count() bool {
// func main() {
	var wg sync.WaitGroup
	chsPK := make([]chan *pkAndId, 4)
	chsProof := make([]chan *ProofAndId, 4)
	chsRnd := make([]chan *RndAndId, 4)
	nodes := [4]node{}

	rand.Seed(time.Now().Unix())
	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		fmt.Println("generate random byte[] failed!")
	}

	for i := 0; i < 4; i++ {
		chsPK[i] = make(chan *pkAndId)
		chsProof[i] = make(chan *ProofAndId)
		chsRnd[i] = make(chan *RndAndId)
		newNode(&nodes[i], i)
	}

	for i := 0; i < 4; i++ {
		wg.Add(2)
		go broadcastPK(chsPK, &nodes[i], i, &wg)
		go storePk(chsPK[i], &nodes, i, &wg)
	}

	wg.Wait()

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go Sotition(&nodes[i], randomness, &wg)
	}

	wg.Wait()

	for i := 0; i < 4; i++ {
		wg.Add(2)
		go broadcastProof(chsProof, &nodes[i], i, &wg)
		go storeProof(chsProof[i], &nodes, i, &wg)
	}
	wg.Wait()


	for i := 0; i < 4; i++ {
		wg.Add(2)
		go broadcastRnd(chsRnd, &nodes[i], i, &wg)
		go storeRnd(chsRnd[i], &nodes, i, randomness, &wg)
	}

	var istrue bool
	size := len(nodes[0].idList)
	var cnt int = 1
	for i := 1; i < 4; i++ {
		tmp := len(nodes[i].idList)
		if tmp == size {
			cnt++
		} else {
			break
		}
	}
	if cnt == 4 {
		istrue = true
	}
	// fmt.Println(istrue)
	return istrue
}