package main

import (
	"fmt"
	"net"
	// "io/ioutil"
	// "os"
	// "strings"
	// initial "myProject/initial"
)

// type Shard struct {
// 	List 	[]*initial.Node
// 	Description 	map[*initial.Node]string
// }


func main() {
	// name := "shard0.txt"
	// f, err := os.Open(name)
	// if err != nil {
	// 	panic(err)
	// }
	// content, err := ioutil.ReadAll(f)

	// // content = string(content) 按照空格划分
	// s := strings.Split(string(content), " ")

	// for _, n := range s {
	// 	fmt.Println(n)
	// }

	// fmt.Println(s)

	// shard := &Shard{}
	// fmt.Println("len = ", len(*shard))
	// for _, _ = range shard.List {
	// 	fmt.Println("shaed end")
	// }


	//得到每个docker容器的ip地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	fmt.Println(addrs[1]) //addrs[0]是127.0.0.1
}