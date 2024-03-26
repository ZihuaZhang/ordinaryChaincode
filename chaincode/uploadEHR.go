/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/ZihuaZhang/chaincode/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	ehrChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})

	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := ehrChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
