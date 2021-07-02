package main

import (
	// "bytes"
	// "encoding/binary"
	"fmt"
	crypto "myProject/crypto" //注意要写绝对路径，从gopath开始
)

// const denominator float64 = 18446744073709551615

func main()  {
	sotition()
	// 1. 生成公私钥
	pk, sk := crypto.VrfKeygen()
	fmt.Println("pk = ", pk)
	fmt.Println("sk = ", sk)

	pk2 := sk.Pubkey()
	fmt.Println(pk2)
	if pk == pk2 {
		fmt.Println("veritfy success")
	}
	// var success bool = false

	// msg := ([]byte)("hello")
	// proof, ok1 := sk.ProveMy(msg)
	// if !ok1 {
	// 	fmt.Println("proof generated error")
	// }
	// // fmt.Println("proof = ", proof)

	// output, ok2 := proof.Hash()
	// if !ok2 {
	// 	fmt.Println("output generated error")
	// }
	// // fmt.Println("output = ", output)
	// ok3, output2 := pk.VerifyMy(proof, msg)
	// if !ok3 {
	// 	fmt.Println("verifyMy error")
	// }
	// if output == output2 {
	// 	success = true
	// 	fmt.Println("verify success")
	// }

	// return success

	// //output转换成[]byte类型
	// bytesBuffer := bytes.NewBuffer(output[:])

	// //[64]byte 
	// var x int64
  	// binary.Read(bytesBuffer, binary.BigEndian, &x)

  	// rnd := float64(x)
	// if rnd < 0 {
	// 	rnd += denominator
	// }
	
	// p := rnd/denominator //表示概率
	// if p < 0.7 { //允许进入系统
	// 	return
	// }
}

func sotition() {
	// var success bool = false
	_, sk := crypto.VrfKeygen()

	msg := ([]byte)("hello")
	proof, ok := sk.ProveMy(msg) //多次调用仍然会得到相同的输出

	if !ok {
		fmt.Println("proof generated error")
	}
	output, ok := proof.Hash() // 对proof的hash结果是确定的
	if !ok {
		fmt.Println("output generated error")
	}
	fmt.Printf("output1 = %v\n", output)
}

/*
1. keypair. 生成公私钥，这个肯定是证明人生成的，会在全网广播自己的公钥。这样所有人都有公钥
	- 参与10个人
	- 这10个人每个人都知道其他人的公钥
	- 这10个人每个人都保存自己的公钥

2. 证明人进行加密抽签，会生成一个固定长度的 随机值rnd 和 一个proof
	- 这个proof可以证明这个rnd是证明人产生的关于输入消息msg的证明
	- 这个随机值rnd表示抽签的结果。满足某个阈值，就表示成为验证者。
	- 验证者之间用 节点发现算法，找到所有的验证者

3. 其他验证者对证明人的输出进行验证

4. 现在rnd所有满足条件的节点组成验证者，可以参与系统后续的分片操作了
- 所有节点应该知道其他验证者的存在
- 也知道要分片的个数。比方验证者有10个，就分2组。seed(10)随机数，然后取余，得到自己是哪个分片的
- 分片内部进行PBFT共识

5. 共识 --- 收集吞吐量和时延

*/