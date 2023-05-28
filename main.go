package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"myproject/sdkInit"
	"myproject/service"
	"myproject/web"
	"os"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	shell "github.com/ipfs/go-ipfs-api"
)

const (
	cc_name    = "simplecc"
	cc_version = "1.0.0"
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

func main() {
	// init orgs information
	orgs := []*sdkInit.OrgInfo{
		{
			OrgAdminUser:  "Admin",
			OrgName:       "Org1",
			OrgMspId:      "Org1MSP",
			OrgUser:       "User1",
			OrgPeerNum:    1,
			OrgAnchorFile: os.Getenv("GOPATH") + "/src/myproject/fixtures/channel-artifacts/Org1MSPanchors.tx",
		},
	}

	// init sdk env info
	info := sdkInit.SdkEnvInfo{
		ChannelID:        "mychannel",
		ChannelConfig:    os.Getenv("GOPATH") + "/src/myproject/fixtures/channel-artifacts/channel.tx",
		Orgs:             orgs,
		OrdererAdminUser: "Admin",
		OrdererOrgName:   "OrdererOrg",
		OrdererEndpoint:  "orderer.example.com",
		ChaincodeID:      cc_name,
		ChaincodePath:    os.Getenv("GOPATH") + "/src/myproject/chaincode/",
		ChaincodeVersion: cc_version,
	}
	print(os.Getenv("GOPATH") + "/src/myproject/fixtures/channel-artifacts/channel.tx")
	// sdk setup
	sdk, err := sdkInit.Setup("config.yaml", &info)
	if err != nil {
		fmt.Println(">> SDK setup error:", err)
		os.Exit(-1)
	}

	// create channel and join
	if err := sdkInit.CreateAndJoinChannel(&info); err != nil {
		fmt.Println(">> Create channel and join error:", err)
		os.Exit(-1)
	}

	// create chaincode lifecycle
	if err := sdkInit.CreateCCLifecycle(&info, 1, false, sdk); err != nil {
		fmt.Println(">> create chaincode lifecycle error: %v", err)
		os.Exit(-1)
	}

	// invoke chaincode set status
	fmt.Println(">> 通过链码外部服务设置链码状态......")

	addressA := "张三"
	addressB := "李四"

	sh := shell.NewShell("localhost:5001")
	inputString := "hello, world"
	cid, err := sh.Add(bytes.NewReader([]byte(inputString)))

	if err != nil {
		fmt.Println("Error adding string to IPFS:", err)
		return
	}
	fmt.Println("Added string to IPFS with CID:", cid)

	PrivateA, _ := service.GeneratePrivateKey()
	PubkeyA, _ := service.GeneratePublicKey(&PrivateA)

	PubkeyAStr := base64.StdEncoding.EncodeToString(PubkeyA[:])
	serviceSetup, err := service.InitService(info.ChaincodeID, info.ChannelID, info.Orgs[0], sdk)
	if err != nil {
		fmt.Println()
		os.Exit(-1)
	}
	fmt.Println("公钥: " + PubkeyAStr)

	//计算时间
	start := time.Now()
	msg, err := serviceSetup.StorePubKey(addressA, addressB, PubkeyAStr)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("存储公钥信息发布成功, 交易编号为: " + msg)
		fmt.Println("存储公钥信息所需时间: ", elapsed)

	}

	start = time.Now()
	result, err := serviceSetup.GetPubKey(addressA, addressB)
	elapsed = time.Since(start)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("公钥查询信息成功：")
		fmt.Println("公钥查询信息所需时间: ", elapsed)
		fmt.Println(result)
	}

	decoded, err := base64.StdEncoding.DecodeString(result)
	var PubkeyAStrFromChaincode [32]byte
	copy(PubkeyAStrFromChaincode[:], decoded[:32])

	start = time.Now()
	msg, err = serviceSetup.GenerateZkParams(addressA, addressB)
	elapsed = time.Since(start)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("生成zkparams成功, 交易编号为: " + msg)
		fmt.Println("生成zkparams所需时间: ", elapsed)

	}

	start = time.Now()
	zk_pk, err := serviceSetup.GetZkProvingKey(addressA, addressB)
	elapsed = time.Since(start)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("获取zkprovingkey成功")
		fmt.Println("生成zkprovingkey所需时间: ", elapsed)
	}

	start = time.Now()
	zk_cs, err := serviceSetup.GetZkCS(addressA, addressB)
	elapsed = time.Since(start)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("获取zkcs成功")
		fmt.Println("生成zkcs所需时间: ", elapsed)
	}

	var pk_buf bytes.Buffer
	_, err = pk_buf.Write(zk_pk)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("写入zk_pk")
	}
	pk := groth16.NewProvingKey(ecc.BN254)
	pk.ReadFrom(&pk_buf)

	var cs_buf bytes.Buffer
	_, err = cs_buf.Write(zk_cs)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("写入zk_cs:")
	}
	newR1CS := groth16.NewCS(ecc.BN254)
	newR1CS.ReadFrom(&cs_buf)

	assignment := &Circuit{
		PreImage: "16130099170765464552823636852555369511329944820189892919423002775646948828469",
		Hash:     "12886436712380113721405259596386800092738845035233065858332878701083870690753",
	}

	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		fmt.Println("Error creating witness:", err)
		return
	}
	publicWitness, _ := witness.Public()

	proof, err := groth16.Prove(newR1CS, pk, witness)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("生成proof成功")
	}

	data, err := publicWitness.MarshalBinary()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("publicWitness.MarshalBinary成功")
	}

	var proof_buf bytes.Buffer
	proof.WriteTo(&proof_buf)

	start = time.Now()
	IsValid, err := serviceSetup.VerifyZkProof(addressA, addressB, proof_buf.String(), string(data))
	elapsed = time.Since(start)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("IsValid", IsValid)
		fmt.Println("VerifyZkProof所需时间: ", elapsed)
	}

	app := web.Application{
		Setup:       serviceSetup,
		PrivateKey:  PrivateA,
		PublicKey:   PubkeyA,
		IpfsAddress: cid,
	}
	web.WebStart(app)
}
