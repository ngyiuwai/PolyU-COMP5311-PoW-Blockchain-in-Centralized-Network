package main

// This program is written by assuming array of strings is packed in block.
// So need convertor to convert it to array of byte stream.
// Affected:
// sysUI.go         -> for overall user interface
// nodeMiner.go     -> for user interface when acting as a miner.   [main puropse of nodeMiner.go is data exchange in network, which is not affected.]
// nodeServer.go    -> for user interface when acting as a server.  [main puropse of nodeServer.go is data exchange in network, which is not affected.]
// blockchainSyn.go -> for Genesis Block Creation.					[main puropse of blockchainSyn.go is syndication of blockchain between nodes (in RAM) & database (in disk), which is not affected.]

func arrayConvertorStringToBytes(input []string) (output [][]byte) {
	for i := 0; i < len(input); i++ {
		output = append(output, []byte(input[i]))
	}
	return output
}
