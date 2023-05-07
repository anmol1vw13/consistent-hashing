package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

type Node struct {
	Name string
	Ip   string
	data []Data
	hash int
}

type Data struct {
	value string
	hash  int
}

type ConsistentHasher struct {
	total_slots int
	nodes       []Node
	keys        []int
}

func searchIndex(arr []int, left int, right int, number int) (int, bool) {
	if left >= right {
		return right, false
	}
	if left < 0 {
		return 0, false
	}

	mid := (right + left) / 2

	if arr[mid] == number {
		return mid, true
	} else if arr[mid] > number {
		return searchIndex(arr, left, right-1, number)
	}

	return searchIndex(arr, left+1, right, number)

}

func (hasher *ConsistentHasher) assignData(value string) Node {
	hash := hasher.hash_fn(value)
	pos, _ := searchIndex(hasher.keys, 0, len(hasher.keys), hash)
	if pos >= len(hasher.keys) {
		pos = 0
	}
	hasher.nodes[pos].data = append(hasher.nodes[pos].data, Data{value, hash})
	return hasher.nodes[pos]
}

func (hasher *ConsistentHasher) moveData(position int, newNode *Node) string {
	sourceNodeIndex := position % len(hasher.nodes)

	if sourceNodeIndex < 0 {
		return ""
	}
	var prevNodeIndex int
	if sourceNodeIndex-1 < 0 {
		prevNodeIndex = len(hasher.nodes) - 1
	} else {
		prevNodeIndex = sourceNodeIndex - 1
	}

	if prevNodeIndex == -1 {
		return ""
	}

	sourceNode := &hasher.nodes[sourceNodeIndex]
	prevNode := hasher.nodes[prevNodeIndex]

	for index, data := range sourceNode.data {
		if (data.hash > prevNode.hash && data.hash <= newNode.hash) || (position == 0 && data.hash > sourceNode.hash) {
			newNode.data = append(newNode.data, data)
			sourceNode.data = append(sourceNode.data[:index], sourceNode.data[index+1:]...)
		}
	}

	return fmt.Sprintf("Attempted moving data from %s", sourceNode.Name)
}

func (hasher *ConsistentHasher) hash_fn(val string) int {
	h := sha256.New()
	h.Write([]byte(val))
	md := h.Sum(nil)
	bigInt := new(big.Int).SetBytes(md)
	return int(int64(math.Abs(float64(bigInt.Int64() % int64(hasher.total_slots)))))
}

func (hasher *ConsistentHasher) addNode(name string, ip string) (bool, string){
	node_length := len(hasher.nodes)
	hash := hasher.hash_fn(ip)
	node := Node{Name: name, Ip: ip, hash: hash}
	pos, isPresent := searchIndex(hasher.keys, 0, node_length, hash)

	if isPresent {
		return false, "Cannot allocate this node"
	}

	keys := make([]int, node_length+1)
	nodes := make([]Node, node_length+1)
	index := 0
	message := fmt.Sprintf("Added the node to index %d", pos)
	if node_length != 0 {
		message = fmt.Sprintf("%s | %s",  message, hasher.moveData(pos, &node))
	}

	for ; index < pos; index++ {
		keys[index] = hasher.keys[index]
		nodes[index] = hasher.nodes[index]
	}
	keys[index] = hash
	nodes[index] = node

	for i := pos; i < node_length; i++ {
		index++
		keys[index] = hasher.keys[i]
		nodes[index] = hasher.nodes[i]
	}

	hasher.keys = keys
	hasher.nodes = nodes
	return true, message
}

func (hasher *ConsistentHasher) removeNode(name string) (bool, string) {

	nodeIndex := -1
	for index, node := range hasher.nodes {
		if node.Name == name {
			nodeIndex = index
			break
		}
	}

	if nodeIndex == -1 {
		return false, "Node not found"
	}

	nextNodeIndex := (nodeIndex + 1) % len(hasher.nodes)
	nextNode := &hasher.nodes[nextNodeIndex]
	currentNode := hasher.nodes[nodeIndex]
	nextNode.data = append(nextNode.data, currentNode.data...)

	hasher.keys = append(hasher.keys[:nodeIndex], hasher.keys[nodeIndex+1:]...)
	hasher.nodes = append(hasher.nodes[:nodeIndex], hasher.nodes[nodeIndex+1:]...)
	return true, fmt.Sprintf("Node has been removed and data has been moved to %s", nextNode.Name)
}

func main() {
	hasher := ConsistentHasher{total_slots: 50}
	hasher.addNode("A", "10.131.213.12")
	hasher.addNode("B", "10.121.213.10")
	hasher.addNode("C", "10.122.212.11")
	hasher.addNode("D", "10.124.222.15")
	hasher.addNode("E", "10.131.213.19")

	fmt.Println("Existing nodes", hasher.nodes)

	for true {
		fmt.Println("What do you want to do? Select a number")
		fmt.Println("1) Add a new node 2) Add data 3) Remove a node 4) exit 9) Check details")
		var val int
		fmt.Scan(&val)

		switch val {
		case 1:
			var name string
			var ip string
			fmt.Print("Enter name: ")
			fmt.Scan(&name)
			fmt.Print("Enter IP: ")
			fmt.Scan(&ip)
			_, message:= hasher.addNode(name, ip)
			fmt.Println(message)
		case 2:
			var data string
			fmt.Print("Enter data: ")
			fmt.Scan(&data)
			node := hasher.assignData(data)
			fmt.Println("Assigned to the node ", node.Name)
		case 3:
			var name string
			fmt.Println("Here are the node details")
			for _, node := range hasher.nodes {
				fmt.Printf("Name %s | IP %s\n", node.Name, node.Ip)
			}
			fmt.Print("Enter the name of the node you want to delete: ")
			fmt.Scan(&name)
			_, message := hasher.removeNode(name)
			fmt.Println(message)
		case 9: 
			fmt.Println(hasher)	
			
		}
		if val == 4 {
			break
		}
	}
}
