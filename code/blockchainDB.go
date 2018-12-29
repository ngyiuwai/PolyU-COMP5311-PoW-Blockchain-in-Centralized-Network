package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// SaveBlock : Save Blockchain in a JSON
func SaveBlock(newBlock *Block, userID string) {

	dbPath := "./database/blocks_" + userID
	os.MkdirAll("./database", os.ModePerm)

	chain := LoadChain(userID)

	// Add new block to this chain (array of block), and replace the whole json with this updated chain
	chain = append(chain, newBlock)
	chainjson, _ := json.Marshal(chain)

	// Write the JSON on disk

	jsonFile, err := os.Create(dbPath + ".json")
	if err != nil {
		fmt.Println("Database:	Cannot write JSON,", err)
	}
	_ = ioutil.WriteFile(dbPath+".json", chainjson, os.ModePerm)
	jsonFile.Close()

}

// LoadChain : Loan Blockchain from JSON
func LoadChain(userID string) []*Block {
	dbPath := "./database/blocks_" + userID

	// Read JSON from disk.
	jsonFile, err := os.Open(dbPath + ".json")
	if err != nil {
		return []*Block{}
	}
	jsonReader, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close()

	// Convert JSON to a chain (array of block)
	var chain []*Block
	err = json.Unmarshal(jsonReader, &chain)

	// Return chain
	return chain
}
