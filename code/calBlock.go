package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"time"
)

// Block : Define object Block
type Block struct {
	// Block Header
	Timestamp     uint32
	PrevBlockHash []byte
	Root          []byte
	Nonce         uint32
	// Block Data
	Data [][]byte
	// Block hash,  can be computed using header
	CurrBlockHash []byte
	// Byte Stream : Serialized Block Header
	//	Block is defines as
	//	8	bytes:	MagicNumber		(16-digit hexadecimal integer, 00004B61726C4E67)
	//	4	bytes:	Timestamp		(10-digit decimal positive integer)
	//	32	bytes:	PrevBlockHash	(64-digit hexadecimal integer)
	//	32	bytes:	MerkleTreeRoot	(64-digit hexadecimal integer)
	//	4	bytes:	Nonce			(10-digit decimal positive integer)
	//	Variable :	Data			(UTF-8)
	// Length of Header = 78 bytes
	ByteStream []byte
}

// CreateBlock : Create new Block
func CreateBlock(dataInput [][]byte, PrevBlockHash []byte) *Block {

	time.Sleep(1 * time.Second)

	block := &Block{
		Timestamp:     uint32(time.Now().Unix()),
		PrevBlockHash: PrevBlockHash,
		Root:          CalRoot(dataInput),
		Data:          dataInput,
	}

	block.CalNoncePOW()

	return block
}

// Serialize : Serialize the Block Header as defined in "type Block struct {}".
//			   The serialized Block Header is used in Proof of Work
func (bk *Block) Serialize() {

	// Convert everythings in header to []byte
	byteMagicNumber := make([]byte, 8)
	byteMagicNumber, _ = hex.DecodeString("00004B61726C4E67")

	byteTimestamp := make([]byte, 4)
	binary.BigEndian.PutUint32(byteTimestamp, bk.Timestamp)

	byteNonce := make([]byte, 4)
	binary.BigEndian.PutUint32(byteNonce, bk.Nonce)

	// Generate serialized header
	bk.ByteStream = bytes.Join(
		[][]byte{
			byteMagicNumber,
			byteTimestamp,
			bk.PrevBlockHash,
			bk.Root,
			byteNonce,
		},
		[]byte{},
	)
}

// CalCurrHash : Calculation CurrBlockHash using SHA256
func (bk *Block) CalCurrHash() {
	h := sha256.New()
	h.Write(bk.ByteStream[8:])
	bk.CurrBlockHash = h.Sum(nil)
}

// targetPOW : The target number of "0" in Proof of Work.
//			   The first n-digit of CurrBlockHash (in hexadecimal integer) should be "0".
const targetPOW = 4

// CalNoncePOW : Start Proof of Work. Change Nonce unit targetPOW is archieved.
func (bk *Block) CalNoncePOW() {
	var tryNonce uint32
	var tryFlag bool
	tryNonce = 0

	for {
		// Step 1 : Convert Block Header to byte stearm, and calculation its CurrBlockHash
		bk.Nonce = tryNonce
		bk.Serialize()
		bk.CalCurrHash()
		tryFlag = true
		// Step 2 : If the 1st n-digit of CurrBlockHash == "0", miningin is success
		for i := 0; i < targetPOW; i = i + 1 {
			if hex.EncodeToString(bk.CurrBlockHash)[i:i+1] != "0" {
				tryFlag = false
			}
		}
		// Step 3: Exit if mining is success, else try nonce = nonce + 1
		if tryFlag == false {
			tryNonce = tryNonce + 1
		} else {
			break
		}
	}

}

// ValidateBlock : Check if the block is valid.
func (bk *Block) ValidateBlock() bool {

	// Step 1 : Extract basic Header information
	chkBk := &Block{
		Timestamp:     bk.Timestamp,
		PrevBlockHash: bk.PrevBlockHash,
		Root:          bk.Root,
		Nonce:         bk.Nonce,
	}

	// Step 2 : Calculation CurrBlockHash
	chkBk.Serialize()
	chkBk.CalCurrHash()

	// Step 3 : Check if CurrBlockHash is valid
	var chkFlag bool
	chkFlag = true
	for i := 0; i < targetPOW; i = i + 1 {
		if hex.EncodeToString(chkBk.CurrBlockHash)[i:i+1] != "0" {
			chkFlag = false
		}
	}

	return chkFlag
}
