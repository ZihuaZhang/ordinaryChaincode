/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/ZihuaZhang/redactable/src"
	"github.com/fentec-project/gofe/abe"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/wumansgy/goEncrypt/aes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"math/big"
	"os"
	"path"
	"time"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "../../test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
)

func main() {
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "healme"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "healme"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	mspView, _ := abe.BooleanToMSP("((0 AND 1) OR (2 AND 3)) AND 5", false)
	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)
	pk, _, err := src.NewFAME().GenerateMasterKeys()
	// 生成AES对称密钥
	key24 := "123456781234567812345678"
	// 对称密钥AES加密病历
	plaintext := "Patient Name: John Smith\nDate of Birth: January 1, 1980\nGender: Male\nMedical Record Number: MRN123456789\nDate of Admission: March 16, 2024\nChief Complaint: Patient presents with chest pain and shortness of breath.\nHistory of Present Illness:\nThe patient reports experiencing chest pain for the past 24 hours. He describes the pain as a dull ache that radiates to his left arm. He also reports associated shortness of breath and diaphoresis. Symptoms worsened with exertion."
	_, err = aes.AesCtrEncrypt([]byte(plaintext), []byte(key24), nil)
	if err != nil {
		return
	}
	// 加密病历上传IPFS
	cid := "QmYpLtwy3gHbg9Jh8ae5g11aQeh7P69KEWttdS7p3NB12j"
	// 对对称密钥进行FAME加密
	keyBigInt, _ := new(big.Int).SetString(key24, 10)
	cipher, err := src.NewFAME().Encrypt(keyBigInt, mspView, pk)
	if err != nil {
		return
	}
	ciperJSON, err := json.Marshal(cipher)
	_, err = contract.SubmitTransaction("UploadEHR", cid, string(ciperJSON))
	if err != nil {
		return
	}
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificatePEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to read TLS certifcate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificatePEM, err := readFirstFile(certPath)
	if err != nil {
		panic(fmt.Errorf("failed to read certificate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	privateKeyPEM, err := readFirstFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func readFirstFile(dirPath string) ([]byte, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	fileNames, err := dir.Readdirnames(1)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path.Join(dirPath, fileNames[0]))
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}
