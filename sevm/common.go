package sevm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"context"
	"log"

	"encoding/json"
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"
)

var RPC_URL string

func init() {
	filePath := "./sevm/config.json"

	file, err := os.Open(filePath)
	fmt.Println("Current working directory:", filePath)

	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	configData := make(map[string]string)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configData)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	url, ok := configData["rpc_url"]
	if !ok {
		log.Fatalf("rpc_url not found in the config file")
	}

	RPC_URL = url
}

func StringToBytes(s string) []byte {
	if len(s)%2 != 0 {
		panic(fmt.Sprintf("StringToBytes: invalid input string %q", s))
	}

	var result []byte
	for i := 0; i < len(s); i += 2 {
		b, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			panic(err)
		}
		result = append(result, byte(b))
	}
	return result
}

func PrintOperationLog(operation byte, operands z3.Value) {

	// if operands != nil {
	// 	fmt.Println("Operation: ", vm.OpCode(operation).String(), " ", operands)
	// } else {
	// 	fmt.Println("Operation: ", vm.OpCode(operation).String())
	// }

}

func GetContractCodeByAddress(contractAddress string) []byte {
	client, err := ethclient.Dial(
		RPC_URL,
	)
	if err != nil {
		log.Fatal(err)
	}

	// bigInt := big.NewInt(14684307)
	// color.Println(contractAddress)
	bytecode, err := client.CodeAt(context.Background(), common.HexToAddress(contractAddress), nil) // nil is latest block
	if err != nil {
		log.Fatal(err)
	}

	return bytecode

}

func IsSmartContractAddress(contractAddress string) bool {
	code := GetContractCodeByAddress(contractAddress)
	return len(code) > 0

}

func SimplifyZ3BV(value z3.BV) z3.BV {
	return value.Context().Simplify(value, value.Context().Config()).(z3.BV)
}

func PauseForSeconds(seconds int) {
	fmt.Printf("Sleep %d S....\n", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)

}

func GetStorageAtHash(account common.Address, key common.Hash, blockHash common.Hash) []byte {
	ethClient, err := ethclient.Dial(
		RPC_URL,
	)
	if err != nil {
		panic(err)
	}

	result, err := ethClient.StorageAt(context.Background(), account, key, blockHash.Big())
	if err != nil {
		panic(err)
	}

	return result

}

func IsConcreteValue(value z3.BV) bool {
	_, ok := SimplifyZ3BV(value).AsBigUnsigned()
	if !ok { // can't be converted
		return false
	} else {
		return true
	}

}

func removeBackslashesAndPipes(input string) string {
	result := strings.ReplaceAll(input, "\\", "")
	result = strings.ReplaceAll(result, "|", "")
	return result
}
