package main

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

// minerSendMsg : Just send message to a node. Get reply.
func minerSendMsg(conn net.Conn, msg []byte) (reply []byte) {
	_, err := conn.Write(msg)
	if err != nil {
		fmt.Println("Miner:	Error writing:")
		fmt.Println("Miner:	", err)
		return
	}

	fmt.Println("Miner:	...sending message to nearby node")

	buf := make([]byte, 8192)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Miner:	...Error Reading:")
		fmt.Println("Miner:	...", err)
		return
	}

	fmt.Println("Miner:	...received message from nearby node")
	return bytes.TrimRight(buf, "\x00")
}

// minerGetDataFromUI : Receive data, in format of string, from user.
func minerGetDataFromUI() [][]byte {
	var dataRaw string
	var dataString []string
	fmt.Println("Miner:	Enter data to be packed in blockchain (seperated by ',')")
	fmt.Scan(&dataRaw)
	dataString = strings.Split(dataRaw, ",")
	return arrayConvertorStringToBytes(dataString)
}

// minerPrintBlock : Print a block in command line interface.
func minerPrintBlock(block *Block) {
	fmt.Printf("Miner:	Block Information\n")
	fmt.Printf("	> Timestamp	: %010d\n", block.Timestamp)
	fmt.Printf("	> PrevBlockHash	: %x\n", block.PrevBlockHash)
	fmt.Printf("	> Root		: %x\n", block.Root)
	fmt.Printf("	> Nonce		: %010d\n", block.Nonce)
	fmt.Printf("	> CurrBlockHash	: %x\n", block.CurrBlockHash)
	fmt.Printf("	> Data		: %s\n", block.Data)
	fmt.Printf("      	Header in Byte Stream (80 bytes, equals to 160 digits in hex)\n")
	fmt.Printf("	[Magic#        ][TS    ][PrevBlockHash                                                 ][Root                                                          ][Nonce ]\n")
	fmt.Printf("	%x\n\n", block.ByteStream)

	return
}
