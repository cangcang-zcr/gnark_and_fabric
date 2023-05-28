package service

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

func (t *ServiceSetup) StorePubKey(addressA string, addressB string, pubKey string) (string, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "storePubKey", Args: [][]byte{[]byte(addressA), []byte(addressB), []byte(pubKey)}}
	response, err := t.Client.Execute(req)
	if err != nil {
		return "", err
	}
	

	return string(response.TransactionID), nil
}

func (t *ServiceSetup) StoreIPFSAddress(addressA string, addressB string, ipfsAddress string) (string, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "storeIPFSAddress", Args: [][]byte{[]byte(addressA), []byte(addressB), []byte(ipfsAddress)}}
	response, err := t.Client.Execute(req)
	if err != nil {
		return "", err
	}

	return string(response.TransactionID), nil
}

func (t *ServiceSetup) GetPubKey(addressA string, addressB string) (string, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "getPubKey", Args: [][]byte{[]byte(addressA), []byte(addressB)}}
	response, err := t.Client.Query(req)
	if err != nil {
		return "", err
	}

	return string(response.Payload), nil
}

func (t *ServiceSetup) GetIPFSAddress(addressA string, addressB string) (string, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "getIPFSAddress", Args: [][]byte{[]byte(addressA), []byte(addressB)}}
	response, err := t.Client.Query(req)
	if err != nil {
		return "", err
	}

	return string(response.Payload), nil
}


func (t *ServiceSetup) GenerateZkParams(addressA string, addressB string) (string, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "generateZkParams", Args: [][]byte{[]byte(addressA), []byte(addressB)}}
	response, err := t.Client.Execute(req)
	if err != nil {
		return "", err
	}

	return string(response.TransactionID), nil
}

func (t *ServiceSetup) GetZkProvingKey(addressA string, addressB string) ([]byte, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "getZkProvingKey", Args: [][]byte{[]byte(addressA), []byte(addressB)}}
	response, err := t.Client.Query(req)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (t *ServiceSetup) GetZkVerifyingKey(addressA string, addressB string) ([]byte, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "getVerifyingKey", Args: [][]byte{[]byte(addressA), []byte(addressB)}}
	response, err := t.Client.Query(req)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (t *ServiceSetup) VerifyZkProof(addressA string, addressB string, proof string, publicWitness string) (string, error) {
	req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "verifyZkProof", Args: [][]byte{[]byte(addressA), []byte(addressB), []byte(proof), []byte(publicWitness)}}
	response, err := t.Client.Execute(req)
	if err != nil {
		return "", err
	}
	return string(response.TransactionID), nil
}


func (t *ServiceSetup) GetZkCS(addressA string, addressB string) ([]byte, error) {
    req := channel.Request{ChaincodeID: t.ChaincodeID, Fcn: "getZkCS", Args: [][]byte{[]byte(addressA), []byte(addressB)}}
    response, err := t.Client.Query(req)
    if err != nil {
        return nil, err
    }
    return response.Payload, nil
}
