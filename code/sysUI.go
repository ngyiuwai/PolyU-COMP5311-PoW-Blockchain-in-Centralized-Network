package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	userHost := "localhost"
	serverHost := "localhost"
	var userPort, serverPort string

	if len(os.Args) == 3 {

		// Fast Mode, input userPort and serverPort Arguments
		userPort = os.Args[1]
		serverPort = os.Args[2]

	} else {

		// Normal UI
		fmt.Println("The is a program to demonstrate some key characteristics of Blockchain.")
		fmt.Println("To run this program, you should set up at least two node and one miner.")
		fmt.Println("> One Full Node, UserID = 9999")
		fmt.Println("> At least one Normal Node")
		fmt.Println("> At least one Miner")
		fmt.Println("> UserID of Normal Node/ Miner could any integer between 1025 and 65535")
		fmt.Println()

		// Set Port
		fmt.Print("Please enter userID of you: ")
		fmt.Scanln(&userPort)
		fmt.Print("Please enter userID of Node you wish to connect: ")
		fmt.Scanln(&serverPort)
		fmt.Println()

	}
	run(userHost, userPort, serverHost, serverPort)
}

func run(userHost string, userPort string, serverHost string, serverPort string) {

	// Set address for TCP connection
	userAddr, err := net.ResolveTCPAddr("tcp", userHost+":"+userPort)
	errorMsg(err)
	serverAddr, err := net.ResolveTCPAddr("tcp", serverHost+":"+serverPort)
	errorMsg(err)

	// Choose Function - Either be a nodecontroller, or a miner
	// **Becauses Peer2Peer model (not Server & Client model) is need if a node is miner and nodecontroller at the same time.
	// **Need to make the a TCP socket "Dial" and "Listen" in simultaneously.
	// **By default, Peer2Peer socket is not supported in golang. Need to use external library.
	fmt.Printf("Self Node port %s ; Server Node port %s ; (Full Node @ Port 9999 as a Server)\n", userPort, serverPort)
	fmt.Printf("Enter 10 to become a Node\n")
	fmt.Printf("Enter 20 to become a Miner\n")
	fmt.Printf("Enter 30 to calculate a Merkle Tree Root\n")
	var input string
	fmt.Scanln(&input)
	switch input {
	case "10" /*	Node  Mode */ :

		// Choose Node's action - either show local Blockchain, or Start Server Service
		fmt.Printf("- Enter 11 to Show Blockchain in this server\n")
		fmt.Printf("- Enter 12 to Start acting as a server\n")
		fmt.Scanln(&input)

		switch input {

		case "11" /*Node - Show local Blockchain*/ :
			var selfNodeChain Blockchain
			selfNodeChain.LoadFromLocalDB(userPort)
			if len(selfNodeChain.Blocks) == 0 {
				fmt.Println("Node:	No Data in Local Database")
				break
			}
			fmt.Println("Node:	Blockchain at Local Database:")
			selfNodeChain.PrintChain()

		case "12" /*Node - As a server*/ :

			// Initialize by loading blockchain from Database
			var selfNodeChain Blockchain
			selfNodeChain.LoadFromDB(userPort)
			if len(selfNodeChain.Blocks) == 0 {
				fmt.Println("Node:	Error in loading blockchain. Exit")
				break
			}
			fmt.Println("Node:	Blockchain at local Database:")
			selfNodeChain.PrintChain()

			// Listening from Miner
			listener, err := net.ListenTCP("tcp", userAddr)
			errorMsg(err)
			fmt.Println("Node:	Server Listening on port", userPort)

			// Create new socket if a connection is accepted
			// golang allows multiple connection by default (non-blocking)
			for {
				conn, err := listener.Accept()
				errorMsg(err)
				go handleMsg(conn, selfNodeChain)
			}

		}

	case "20" /* Miner Mode */ :
		// Connect to server node
		conn, err := net.DialTCP("tcp", userAddr, serverAddr)
		errorMsg(err)
		fmt.Printf("Miner:	Connection %s <--> %s\n", userAddr.String(), serverAddr.String())

		// Choose Miner's action - either mining, or check transaction data
		fmt.Println("- Enter 21 to Mine")
		fmt.Println("- Enter 22 to Retrive all Block Hashes at server node")
		fmt.Println("- Enter 23 to Retrive block in blockchain using a block hash")
		fmt.Println("- Enter 24 to Retrive data  in blockchain using a Merkle Tree Root")
		fmt.Scanln(&input)

		switch input {

		case "21" /*Miner - Mining*/ :
			// Request PrevBlockHash
			fmt.Println("Miner:	Request PrevBlockHash from Node")
			message := []byte("addBK")
			PrevBlockHashFromNode := minerSendMsg(conn, message)
			fmt.Printf("Miner:	Received %x\n", PrevBlockHashFromNode)

			// Get Data from user, and build a new Block
			dataToPack := minerGetDataFromUI()
			fmt.Println("Miner:	...mining...")
			newBlock := CreateBlock(dataToPack, PrevBlockHashFromNode)
			if newBlock.ValidateBlock() == true {
				fmt.Println("Miner:	Success! Block information here:")
				minerPrintBlock(newBlock)
			}

			// Serialize block using "encoding/json", then add the action indicator
			fmt.Println("Miner:	Now send the Block to server node.")
			newBlockJSON, _ := json.Marshal(newBlock)
			message = bytes.Join([][]byte{[]byte("addBK"), newBlockJSON}, []byte{})
			fmt.Println("Miner:	Result - ", string(minerSendMsg(conn, message)))

			conn.Close()
			break

		case "22" /*Miner - Check Block Hashes*/ :
			// Request BlockChain
			fmt.Println("Miner:	Request Full Block Hashes from Node")
			message := []byte("getBC")
			blockHashesFromNode := minerSendMsg(conn, message)
			fmt.Printf("Miner:	Received Block Hashes\n")
			conn.Close()

			// BlockChain is in JSON. Need to decode.
			fmt.Printf("Miner:	...Decoding Block Hashes...\n")
			var blockHashes Blockchain
			err = json.Unmarshal(blockHashesFromNode, &blockHashes)

			// Print BlockChain
			fmt.Println("Miner:	Block Hashes from server Node")
			blockHashes.PrintChain()
			fmt.Println("Miner:	Is the Block Hashes valid? -", blockHashes.ValidateChain())
			break

		case "23" /*Miner - Check Single Block*/ :
			// Request BlockChain
			fmt.Print("Miner:	Please input the Block Hash here ")
			fmt.Scanln(&input)
			fmt.Printf("Miner:	Request the Block with Hashes %s\n", input)
			message, _ := hex.DecodeString(input)
			message = bytes.Join([][]byte{[]byte("getBK"), message}, []byte{})
			targetBlockFromNode := minerSendMsg(conn, message)
			fmt.Printf("Miner:	Received the Block\n")
			conn.Close()

			// BlockChain is in JSON. Need to decode.
			fmt.Printf("Miner:	...Decoding the Block...\n")
			var targetBlock Blockchain
			err = json.Unmarshal(targetBlockFromNode, &targetBlock)

			// Print BlockChain
			if len(targetBlock.Blocks) > 0 {
				fmt.Println("Miner:	Target Block is found")
				targetBlock.PrintChain()
			} else {
				fmt.Println("Miner:	Target Block is not found")
			}
			break

		case "24" /*Miner - Check Single Data*/ :
			// Request BlockChain
			fmt.Print("Miner:	Please input the Merkle Tree Root here ")
			fmt.Scanln(&input)
			fmt.Printf("Miner:	Request the Block with Merkle Tree Root %s\n", input)
			message, _ := hex.DecodeString(input)
			message = bytes.Join([][]byte{[]byte("getTX"), message}, []byte{})
			targetBlockFromNode := minerSendMsg(conn, message)
			fmt.Printf("Miner:	Received the Block\n")
			conn.Close()

			// BlockChain is in JSON. Need to decode.
			fmt.Printf("Miner:	...Decoding the Block...\n")
			var targetBlock Blockchain
			err = json.Unmarshal(targetBlockFromNode, &targetBlock)

			// Print BlockChain
			if len(targetBlock.Blocks) > 0 {
				fmt.Println("Miner:	Target Block is found")
				targetBlock.PrintChain()
			} else {
				fmt.Println("Miner:	Target Block is not found")
			}
			break
		}
		break

	case "30" /*Calculated Merkle Tree Root*/ :
		var dataRaw string
		var dataString []string
		var dataBytes [][]byte
		fmt.Println("Tree:	Enter data to be used for Merkle Tree Calculation (seperated by ',')")
		fmt.Scan(&dataRaw)
		dataString = strings.Split(dataRaw, ",")
		dataBytes = arrayConvertorStringToBytes(dataString)
		CalTree(dataBytes)
		fmt.Printf("\n")
		fmt.Printf("Tree:	The Merkle Tree Root is %x\n", CalRoot(dataBytes))
		break

	}
}

func errorMsg(err error) {
	if err != nil {
		fmt.Println("Connection Error:	", err)
		os.Exit(1)
	}
}
