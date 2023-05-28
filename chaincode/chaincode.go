package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Circuit struct {
	// struct tag on a variable is optional
	// default uses variable name and secret visibility.
	PreImage frontend.Variable
	Hash     frontend.Variable `gnark:",public"`
}

func (circuit *Circuit) Define(api frontend.API) error {
	// hash function
	mimc, _ := mimc.NewMiMC(api)

	// specify constraints
	// mimc(preImage) == hash
	mimc.Write(circuit.PreImage)
	api.AssertIsEqual(circuit.Hash, mimc.Sum())

	return nil
}

type Data struct {
	Ipfs    string `json:"ipfs"`
	Pubkey  string `json:"pubkey"`
	IsValid bool   `json:"IsValid"`
}

type ZkParams struct {
	PK []byte `json:"provingKey"`
	VK []byte `json:"verifyingKey"`
	CS []byte `json:"cs"`
}

func (d *Data) toJSONBytes() ([]byte, error) {
	return json.Marshal(d)
}

func fromJSONBytes(bytes []byte) (*Data, error) {
	var d Data
	err := json.Unmarshal(bytes, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

type MyChaincode struct {
}

func (cc *MyChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	err := stub.PutState("MyMap", []byte("{}"))
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to initialize map: %s", err.Error()))
	}
	err = stub.PutState("MyZkMap", []byte("{}"))
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to initialize zk map: %s", err.Error()))
	}

	return shim.Success(nil)
}

func (cc *MyChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	switch fn {
	case "storePubKey":
		return cc.storePubKey(stub, args)
	case "storeIPFSAddress":
		return cc.storeIPFSAddress(stub, args)
	case "getPubKey":
		return cc.getPubKey(stub, args)
	case "getIPFSAddress":
		return cc.getIPFSAddress(stub, args)
	case "generateZkParams":
		return cc.generateZkParams(stub, args)
	case "getZkProvingKey":
		return cc.getZkProvingKey(stub, args)
	case "getZkVerifyingKey":
		return cc.getZkVerifyingKey(stub, args)
	case "getZkCS":
		return cc.getZkCS(stub, args)
	case "verifyZkProof":
		return cc.verifyZkProof(stub, args)

	default:
		return shim.Error(fmt.Sprintf("unknown function: %s", fn))
	}
}

func unmarshalAndLoadDataMap(stub shim.ChaincodeStubInterface, myMapKey string) (map[string]map[string]*Data, error) {
	mapJSON, err := stub.GetState(myMapKey)
	if err != nil {
		return nil, err
	}
	dataMap := make(map[string]map[string]*Data)
	if mapJSON != nil {
		err = json.Unmarshal(mapJSON, &dataMap)
		if err != nil {
			return nil, err
		}
	}
	return dataMap, nil
}

func saveDataMap(stub shim.ChaincodeStubInterface, myMapKey string, dataMap map[string]map[string]*Data) error {
	mapJSON, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}
	err = stub.PutState(myMapKey, mapJSON)
	if err != nil {
		return err
	}
	return nil
}

func (cc *MyChaincode) getPubKey(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return cc.getProductInfo(stub, args, "pubkey")
}

func (cc *MyChaincode) getIPFSAddress(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return cc.getProductInfo(stub, args, "ipfs")
}

func (cc *MyChaincode) storePubKey(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("incorrect number of arguments, expected 3")
	}

	addressA := args[0]
	addressB := args[1]
	pubKey := args[2]

	dataMap, err := unmarshalAndLoadDataMap(stub, "MyMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal and load data map: %s", err.Error()))
	}

	if dataMap[addressA] == nil {
		dataMap[addressA] = make(map[string]*Data)
	}

	entry := dataMap[addressA][addressB]
	if entry == nil {
		entry = &Data{}
	}
	entry.Pubkey = pubKey
	dataMap[addressA][addressB] = entry

	err = saveDataMap(stub, "MyMap", dataMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to save data map: %s", err.Error()))
	}

	return shim.Success(nil)
}

func (cc *MyChaincode) storeIPFSAddress(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("incorrect number of arguments, expected 3")
	}

	addressA := args[0]
	addressB := args[1]
	ipfsAddress := args[2]

	dataMap, err := unmarshalAndLoadDataMap(stub, "MyMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal and load data map: %s", err.Error()))
	}

	if dataMap[addressA] == nil {
		dataMap[addressA] = make(map[string]*Data)
	}

	entry := dataMap[addressA][addressB]
	if entry == nil {
		entry = &Data{}
	}
	entry.Ipfs = ipfsAddress
	dataMap[addressA][addressB] = entry

	err = saveDataMap(stub, "MyMap", dataMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to save data map: %s", err.Error()))
	}

	return shim.Success(nil)
}

func (cc *MyChaincode) getProductInfo(stub shim.ChaincodeStubInterface, args []string, param string) peer.Response {
	if len(args) != 2 {
		return shim.Error("incorrect number of arguments, expected 2")
	}

	addressA := args[0]
	addressB := args[1]

	dataMap, err := unmarshalAndLoadDataMap(stub, "MyMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal and load data map: %s", err.Error()))
	}

	data := dataMap[addressA][addressB]

	if data == nil {
		return shim.Error("data not found")
	}

	var returnVal string
	if param == "pubkey" {
		returnVal = data.Pubkey
	} else if param == "ipfs" {
		returnVal = data.Ipfs
	}
	return shim.Success([]byte(returnVal))
}

func (cc *MyChaincode) generateZkParams(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("incorrect number of arguments, expected 2")
	}

	addressA := args[0]
	addressB := args[1]

	var mimcCircuit Circuit
	_r1cs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &mimcCircuit)

	pk, vk, err := groth16.Setup(_r1cs)

	if err != nil {
		return shim.Error(fmt.Sprintf("failed to generate proving and verifying keys: %s", err))
	}

	var pk_buf bytes.Buffer
	_, err = pk.WriteTo(&pk_buf)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to Write zk_pk To pk_buf: %s", err))
	}

	var vk_buf bytes.Buffer
	_, err = vk.WriteTo(&vk_buf)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to Write zk_vk To vk_buf: %s", err))
	}

	var r1cs_buf bytes.Buffer
	_, err = _r1cs.WriteTo(&r1cs_buf)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to Write zk_r1cs To r1cs_buf: %s", err))
	}

	zkParams := ZkParams{
		PK: pk_buf.Bytes(),
		VK: vk_buf.Bytes(),
		CS: r1cs_buf.Bytes(),
	}
	currentMapJSON, err := stub.GetState("MyZkMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get current map from the ledger: %s", err.Error()))
	}
	currentMap := make(map[string]map[string]*ZkParams)
	if currentMapJSON != nil {
		err = json.Unmarshal(currentMapJSON, &currentMap)
		if err != nil {
			return shim.Error(fmt.Sprintf("failed to unmarshal current map: %s", err.Error()))
		}
	}
	if currentMap[addressA] == nil {
		currentMap[addressA] = make(map[string]*ZkParams)
	}
	currentMap[addressA][addressB] = &zkParams

	updatedMapJSON, err := json.Marshal(currentMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to marshal updated map: %s", err.Error()))
	}

	// Save the updated map to the ledger
	err = stub.PutState("MyZkMap", updatedMapJSON)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to save updated map to the ledger: %s", err.Error()))
	}

	return shim.Success(nil)
}

func (cc *MyChaincode) getZkProvingKey(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("incorrect number of arguments, expected 2")
	}

	addressA := args[0]
	addressB := args[1]

	// Get the map from the ledger
	myZkMapJSON, err := stub.GetState("MyZkMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get MyZkMap from the ledger: %s", err.Error()))
	}

	myZkMap := make(map[string]map[string]*ZkParams)
	err = json.Unmarshal(myZkMapJSON, &myZkMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal MyZkMap: %s", err.Error()))
	}

	zkParams := myZkMap[addressA][addressB]
	if zkParams == nil || len(zkParams.PK) == 0 {
		return shim.Error(fmt.Sprintf("proving key not found for address pair %s, %s", addressA, addressB))
	}

	return shim.Success(zkParams.PK)
}

func (cc *MyChaincode) getZkVerifyingKey(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("incorrect number of arguments, expected 2")
	}

	addressA := args[0]
	addressB := args[1]

	// Get the map from the ledger
	myZkMapJSON, err := stub.GetState("MyZkMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get MyZkMap from the ledger: %s", err.Error()))
	}

	myZkMap := make(map[string]map[string]*ZkParams)
	err = json.Unmarshal(myZkMapJSON, &myZkMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal MyZkMap: %s", err.Error()))
	}

	zkParams := myZkMap[addressA][addressB]
	if zkParams == nil || len(zkParams.VK) == 0 {
		return shim.Error(fmt.Sprintf("verifying key not found for address pair %s, %s", addressA, addressB))
	}

	return shim.Success(zkParams.VK)
}

func (cc *MyChaincode) getZkCS(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("incorrect number of arguments, expected 2")
	}

	addressA := args[0]
	addressB := args[1]

	// Get the map from the ledger
	myZkMapJSON, err := stub.GetState("MyZkMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get MyZkMap from the ledger: %s", err.Error()))
	}

	myZkMap := make(map[string]map[string]*ZkParams)
	err = json.Unmarshal(myZkMapJSON, &myZkMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal MyZkMap: %s", err.Error()))
	}

	zkParams := myZkMap[addressA][addressB]
	if zkParams == nil || len(zkParams.CS) == 0 {
		return shim.Error(fmt.Sprintf("verifying key not found for address pair %s, %s", addressA, addressB))
	}

	return shim.Success(zkParams.CS)
}

func (cc *MyChaincode) verifyZkProof(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("incorrect number of arguments, expecting 2 (addressA, addressB)")
	}

	addressA := args[0]
	addressB := args[1]
	proof_str := args[2]
	public_witness_str := args[3]

	// Get the map from the ledger
	myZkMapJSON, err := stub.GetState("MyZkMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get MyZkMap from the ledger: %s", err.Error()))
	}

	myZkMap := make(map[string]map[string]*ZkParams)
	err = json.Unmarshal(myZkMapJSON, &myZkMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal MyZkMap: %s", err.Error()))
	}

	zkParams := myZkMap[addressA][addressB]
	if zkParams == nil || len(zkParams.VK) == 0 {
		return shim.Error(fmt.Sprintf("VK not found for address pair %s, %s", addressA, addressB))
	}

	var vk_buf bytes.Buffer
	_, err = vk_buf.Write(zkParams.VK)
	if err != nil {
		return shim.Error(fmt.Sprintf("zkParams.VK failed to write to vk_buf: %s", err.Error()))
	}

	var proof_buf bytes.Buffer
	_, err = proof_buf.Write([]byte(proof_str))
	if err != nil {
		return shim.Error(fmt.Sprintf("proof_str failed to write to proof_buf: %s", err.Error()))
	}

	proof := groth16.NewProof(ecc.BN254)
	_, err = proof.ReadFrom(&proof_buf)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to read proof from proof_buf: %s", err.Error()))
	}

	vk := groth16.NewVerifyingKey(ecc.BN254)
	_, err = vk.ReadFrom(&vk_buf)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to read verifying key from vk_buf: %s", err.Error()))
	}

	public_witness, _ := witness.New(ecc.BN254.ScalarField())
	err = public_witness.UnmarshalBinary([]byte(public_witness_str))
	if err != nil {
		return shim.Error(fmt.Sprintf(" failed to public_witness.UnmarshalBinary: %s", err.Error()))
	}

	err = groth16.Verify(proof, vk, public_witness)
	if err != nil {
		return shim.Error(fmt.Sprintf("proof failed to verify: %s", err.Error()))
	}

	// Update the IsValid value in the Data map
	dataMap, err := unmarshalAndLoadDataMap(stub, "MyMap")
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal and load data map: %s", err.Error()))
	}

	data := dataMap[addressA][addressB]
	if data == nil {
		data = &Data{}
	}
	data.IsValid = true
	dataMap[addressA][addressB] = data

	err = saveDataMap(stub, "MyMap", dataMap)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to save data map: %s", err.Error()))
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(MyChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
