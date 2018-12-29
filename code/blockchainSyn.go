package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
)

//Blockchain : Define object Blockchain
type Blockchain struct {
	UserID string
	Blocks []*Block
}

//Set Full Node Information here
const fullNodeHost = "localhost"
const fullNodePort = "9999"

// LoadFromDB :	Load Blockchain from Database - Both Full Node Database and Local Database.
// 				If Full Node Database is longer then Local Database, add Full Node's Block headers to Local Blockchain.
func (bc *Blockchain) LoadFromDB(userID string) {

	// Get blockchain from Local Blockchain and Full Node Blockchain
	bc.UserID = userID
	bc.LoadFromLocalDB(bc.UserID)
	var bcFullNode Blockchain
	bcFullNode.LoadFromFullNode(bc.UserID)

	// If Full Node Blockchain is longer then Local Database, save Full Node's Block headers into Database.
	targetBCLen := len(bcFullNode.Blocks)
	initialBCLen := len(bc.Blocks)
	for initialBCLen < targetBCLen {
		// Add directly to Local Database without validating the Block.
		// because it is downloaded from Full Node, should be fine.
		bc.AddBlockDirect(bcFullNode.Blocks[initialBCLen])
		initialBCLen = len(bc.Blocks)
	}
}

// LoadFromLocalDB :	Load Blockchain from Local Database. Return an array of Blocks in memory.
func (bc *Blockchain) LoadFromLocalDB(userID string) {
	bc.UserID = userID
	bc.Blocks = LoadChain(bc.UserID)
	return
}

// LoadFromFullNode : Load Blockchain from Full Node using TCP. Return an array of Blocks in memory.
func (bc *Blockchain) LoadFromFullNode(userID string) {

	// For Full Node : Just load from Local Database.
	// If Local Database is empty, create the genesisBlock in memory. Save the block to Database in LoadFromDB())
	if userID == fullNodePort {
		bc.LoadFromLocalDB(userID)
		if len(bc.Blocks) == 0 {
			genesisHash, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
			genesisBlock := CreateBlock(arrayConvertorStringToBytes([]string{"New", "Genesis", "Block"}), genesisHash)
			bc.Blocks = []*Block{genesisBlock}
		}
		return
	}

	// For Normal Node : Establish TCP Connection with Full Node. Download block headers.
	fullNodeAddr, err := net.ResolveTCPAddr("tcp", fullNodeHost+":"+fullNodePort)
	if err != nil {
		fmt.Println("Chain:	Cannot connect to Full Node. Fail to load Blockchain.")
		fmt.Println(err)
		return
	}
	fullNodeConn, err := net.DialTCP("tcp", nil, fullNodeAddr)
	if err != nil {
		fmt.Println("Chain:	Cannot connect to Full Node. Fail to load Blockchain.")
		fmt.Println(err)
		return
	}
	fmt.Printf("Chain:	Connected to Full Node for Synchronization of Blockchain : %s\n", fullNodeAddr.String())

	// Full Node is connected, now
	// 1. Send "getBC"
	// 2. Full Node return Block headers, i.e. Blockchain with header only. Deserialize it.

	message := []byte("getBC")
	_, err = fullNodeConn.Write(message)
	buf := make([]byte, 8192)
	_, err = fullNodeConn.Read(buf)
	fullNodeConn.Close()
	// End of Step 1. buf should be a serialized blockchain []*Block in JSON.

	err = json.Unmarshal(bytes.TrimRight(buf, "\x00"), &bc)
	// End of Step 2. Deserialize done.
	return

}

// AddBlock :	Add a new generated block into blockchain - array of block & save in database (Both Local & Full Node)
// 				First will try to add the block to Full Node Database . Then add the block to Local Database.
func (bc *Blockchain) AddBlock(newBlock *Block) bool {

	bc.LoadFromDB(bc.UserID)

	// First add block to Full Node Blockchain (Skip if this step is Full Node)
	if bc.AddBlockFullNode(newBlock) == false {
		return false
	}

	// Then add the block to Local Database. Verify its hash before adding.
	preBlock := bc.Blocks[len(bc.Blocks)-1]
	if string(newBlock.PrevBlockHash) == string(preBlock.CurrBlockHash) {
		bc.Blocks = append(bc.Blocks, newBlock)
		SaveBlock(newBlock, bc.UserID)
		fmt.Println("Chain:	Success in adding Block to Local Database.")
		return true
	}

	fmt.Println("Chain:	Failed to add block. Invalid Hash.")
	return false

}

// AddBlockDirect :	Add a block to Local Database directly, not verify it using by sending to Full Node.
//					Use this function when synchronize with Full Node only.
func (bc *Blockchain) AddBlockDirect(newBlock *Block) {

	// Allows to add Genesis Block if Blockchain Length == 0
	if len(bc.Blocks) == 0 {
		bc.Blocks = append(bc.Blocks, newBlock)
		SaveBlock(newBlock, bc.UserID)
		return
	}

	// Check the hash before adding block.
	preBlock := bc.Blocks[len(bc.Blocks)-1]
	if string(newBlock.PrevBlockHash) == string(preBlock.CurrBlockHash) {
		bc.Blocks = append(bc.Blocks, newBlock)
		SaveBlock(newBlock, bc.UserID)
		return
	}
	fmt.Println("Chain:	Failed to add block when synchronize with full node.")
	return

}

// AddBlockFullNode : Add a block to Full Node by establish a TCP connection
//					  Return true (success) or false (fail)
func (bc *Blockchain) AddBlockFullNode(newBlock *Block) bool {

	// Always return true if the node is Full Node. Because Full node doesn't need to verify block with nearby node.
	if bc.UserID == fullNodePort {
		return true
	}
	// Else Send Block to Full Node
	fullNodeAddr, err := net.ResolveTCPAddr("tcp", fullNodeHost+":"+fullNodePort)
	if err != nil {
		fmt.Println("Chain:	Cannot connect to Full Node. Fail to add block.")
		fmt.Println(err)
		return false
	}
	fullNodeConn, err := net.DialTCP("tcp", nil, fullNodeAddr)
	if err != nil {
		fmt.Println("Chain:	Cannot connect to Full Node. Fail to add block.")
		fmt.Println(err)
		return false
	}
	fmt.Printf("Chain:	Connected to Full Node for Adding Block: %s\n", fullNodeAddr.String())

	// Full Node is connected, now
	// 1. Send "addBK". By Default Full Node return PrevBlockHash if connection is success. Ignore this message.
	// 2. Send the new block to Full Node.
	// 3. Full Node return either "Success..." or "Fail...". Node can determine whether broadcasting is successfully added to Full Node.

	// Step 1. Send "addBK". Ignore returned message.
	message := []byte("addBK")
	_, err = fullNodeConn.Write(message)
	buf := make([]byte, 8192)
	_, err = fullNodeConn.Read(buf)

	// Step 2. Send the new block to Full Node.
	newBlockJSON, _ := json.Marshal(newBlock)
	message = bytes.Join([][]byte{[]byte("addBK"), newBlockJSON}, []byte{})
	_, err = fullNodeConn.Write(message)

	// Step 3. Receive Result from Full Node.
	buf = make([]byte, 8192)
	_, err = fullNodeConn.Read(buf)
	fullNodeConn.Close()
	result := string(bytes.TrimRight(buf, "\x00"))[0:7]

	if result == "Success" {
		fmt.Println("Chain:	Result - ", result, "in adding Block to Full Node")
		return true
	} else if result == "Fail  " {
		fmt.Println("Chain:	Result - ", result, "in adding Block to Full Node")
		return false
	}

	return false

}

// PrintChain :	Print all blocks in blockchain
func (bc *Blockchain) PrintChain() {
	for i := 0; i < len(bc.Blocks); i++ {
		fmt.Printf("Chain:	Block #%d\n", i)
		fmt.Printf("	> Timestamp	: %010d\n", bc.Blocks[i].Timestamp)
		fmt.Printf("	> PrevBlockHash	: %x\n", bc.Blocks[i].PrevBlockHash)
		fmt.Printf("	> Root		: %x\n", bc.Blocks[i].Root)
		fmt.Printf("	> Nonce		: %010d\n", bc.Blocks[i].Nonce)
		fmt.Printf("	> CurrBlockHash	: %x\n", bc.Blocks[i].CurrBlockHash)
		fmt.Printf("	> Data		: %s\n", bc.Blocks[i].Data)
		fmt.Printf("      	Block #%d Header in Byte Stream (80 bytes, equals to 160 digits in hex)\n", i)
		fmt.Printf("	[Magic#        ][TS    ][PrevBlockHash                                                 ][Root                                                          ][Nonce ]\n")
		fmt.Printf("	%x\n\n", bc.Blocks[i].ByteStream)
	}
}

// ValidateChain :	Check if the whole chain is valid. i.e. all CurrBlockHash & PrevBlockHash match.
func (bc *Blockchain) ValidateChain() bool {

	validFlag := true
	if len(bc.Blocks) == 0 {
		validFlag = false
	} else {

		// Check Genesis Block first:	Only Check CurrBlockHash is Valid
		if bc.Blocks[0].ValidateBlock() == false {
			validFlag = false
		}
		// Then check other Blocks :	Check CurrBlockHash is Valid & PrevBlockHash Matches
		if len(bc.Blocks) > 1 {
			for i := 1; i < len(bc.Blocks); i++ {
				if bc.Blocks[i].ValidateBlock() == false {
					validFlag = false
				}
				if string(bc.Blocks[i-1].CurrBlockHash) != string(bc.Blocks[i].PrevBlockHash) {
					validFlag = false
				}
			}
		}
	}
	return validFlag
}
