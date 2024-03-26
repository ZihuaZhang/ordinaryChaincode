package chaincode

import (
	"encoding/json"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type EHR struct {
	UserID       string `json:"UserID"`
	LocationIPFS string `json:"LocationIPFS"`
	KeyCipher    string `json:"KeyCipher"`
}

var totalNum = 0

func (s *SmartContract) UploadEHR(ctx contractapi.TransactionContextInterface, locationIPFS string, keyCiper string) (string, error) {

	aEHR := EHR{
		UserID:       strconv.Itoa(totalNum),
		LocationIPFS: locationIPFS,
		KeyCipher:    keyCiper,
	}
	totalNum++
	aEHRJSON, err := json.Marshal(aEHR)
	if err != nil {
		return "", err
	}
	err = ctx.GetStub().PutState(aEHR.UserID, aEHRJSON)
	return aEHR.UserID, err
}

func (s *SmartContract) QueryEHR(ctx contractapi.TransactionContextInterface, userID string) (*EHR, error) {
	aEHRJSON, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return nil, err
	}
	var aEHR EHR
	err = json.Unmarshal(aEHRJSON, &aEHR)
	if err != nil {
		return nil, err
	}

	return &aEHR, nil
}

func (s *SmartContract) AddEHR(ctx contractapi.TransactionContextInterface, userID string, newLocationIPFS string) error {
	aEHRJSON, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return err
	}
	var aEHR EHR
	err = json.Unmarshal(aEHRJSON, &aEHR)
	if err != nil {
		return err
	}
	aNewEHR := EHR{
		UserID:       userID,
		LocationIPFS: aEHR.LocationIPFS + "," + newLocationIPFS,
		KeyCipher:    aEHR.KeyCipher,
	}
	aNewEHRJSON, err := json.Marshal(aNewEHR)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(aEHR.UserID, aNewEHRJSON)
	return nil
}
