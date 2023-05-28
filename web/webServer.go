package web

import (
	"encoding/base64"
	"fmt"
	"myproject/service"
	"net/http"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark-crypto/ecc"
	"bytes"

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


type Application struct {
	Setup *service.ServiceSetup
	PrivateKey [32]byte
	PublicKey	[32]byte
	IpfsAddress	string
}


func (app *Application)storePubKey(w http.ResponseWriter, r *http.Request) {
	AddressA := r.URL.Query().Get("addressA")//李四
	AddressB := r.URL.Query().Get("addressB")//张三
	PubKey := r.URL.Query().Get("PubKey")
	msg, err := app.Setup.StorePubKey(AddressA, AddressB, PubKey)
	if err != nil {
		returnValue := fmt.Sprintf(err.Error())
		fmt.Fprintf(w, returnValue)
	} else {
		returnValue := fmt.Sprintf("The information has been published successfully, the transaction number is: " + msg)
		fmt.Fprintf(w, returnValue)
	}
	var PubKeystr = PubKey
	Pubkey_from, err := base64.StdEncoding.DecodeString(PubKeystr)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("解析公钥成功:")
		fmt.Println(len(Pubkey_from))

	}
	var Pubkey [32]byte
	
	if len(Pubkey_from) >= 32 {
		copy(Pubkey[:], Pubkey_from[:32])
		sharedSecret, err := service.ComputeSharedSecret(&app.PrivateKey, &Pubkey)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("share密钥: " + base64.StdEncoding.EncodeToString(sharedSecret[:]))
		}
		cipher,err := service.EncryptWithSharedSecret([]byte(app.IpfsAddress), &sharedSecret)
		//张三   李四
		cipher_b64 := base64.StdEncoding.EncodeToString(cipher[:])
		msg, err = app.Setup.StoreIPFSAddress(AddressB, AddressA, cipher_b64)
		fmt.Println("存储的加密ipfs地址:", cipher_b64)
		if err != nil {
		fmt.Println(err.Error())
		} else {
		fmt.Println("存储加密ipfs地址成功，交易id: " + msg)
		}

	} else {
		fmt.Println("Invalid public key length" + string(len(Pubkey_from)))
	}
}

func (app *Application) generateZkParams(w http.ResponseWriter, r *http.Request) {
	AddressA := r.URL.Query().Get("addressA")
	AddressB := r.URL.Query().Get("addressB")
	msg, err := app.Setup.GenerateZkParams(AddressA, AddressB)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("生成zkparams成功, 交易编号为: " + msg)
	}
}
 

func (app *Application) verify(w http.ResponseWriter, r *http.Request) {
	AddressA := r.URL.Query().Get("addressA")
	AddressB := r.URL.Query().Get("addressB")
	hashkey := r.URL.Query().Get("hashkey")
	hhkey := service.Mimchash(hashkey)


	zk_pk, err := app.Setup.GetZkProvingKey(AddressA, AddressB)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("获取zkprovingkey成功")
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


	zk_cs, err := app.Setup.GetZkCS(AddressA, AddressB)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("获取zkcs成功")
	}

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
		PreImage: hashkey,
		Hash:     hhkey,
	}


	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		fmt.Println("Error creating witness:", err)
		return
	}
	proof, err := groth16.Prove(newR1CS, pk, witness)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("生成proof成功")
	}

	publicWitness, _ := witness.Public()
	data, err := publicWitness.MarshalBinary()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("publicWitness.MarshalBinary成功")
	}

	var proof_buf bytes.Buffer
	proof.WriteTo(&proof_buf)

	IsValid, err := app.Setup.VerifyZkProof(AddressA, AddressB, proof_buf.String(), string(data))
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Fprintf(w, IsValid)
	}

}



func (app *Application)getPubKey(w http.ResponseWriter, r *http.Request) {
	AddressA := r.URL.Query().Get("addressA")
	AddressB := r.URL.Query().Get("addressB")
	result, err := app.Setup.GetPubKey(AddressA, AddressB)
	if err != nil {
		returnValue := fmt.Sprintf(err.Error())
		fmt.Fprintf(w, returnValue)
	} else {
		returnValue := fmt.Sprintf(result)
		fmt.Fprintf(w, returnValue)
	}
}

func (app *Application)storeIPFSAddress(w http.ResponseWriter, r *http.Request) {
	AddressA := r.URL.Query().Get("addressA")
	AddressB := r.URL.Query().Get("addressB")
	IPFSAddress := r.URL.Query().Get("IPFSAddress")
	msg, err := app.Setup.StoreIPFSAddress(AddressA, AddressB, IPFSAddress)
	if err != nil {
		returnValue := fmt.Sprintf(err.Error())
		fmt.Fprintf(w, returnValue)
	} else {
		returnValue := fmt.Sprintf("The information has been published successfully, the transaction number is: " + msg)
		fmt.Fprintf(w, returnValue)
	}
}




func (app *Application)getIPFSAddress(w http.ResponseWriter, r *http.Request) {
	AddressA := r.URL.Query().Get("addressA")
	AddressB := r.URL.Query().Get("addressB")
	result, err := app.Setup.GetIPFSAddress(AddressA, AddressB)
	if err != nil {
		returnValue := fmt.Sprintf(err.Error())
		fmt.Fprintf(w, returnValue)
	} else {
		returnValue := fmt.Sprintf(result)
		fmt.Fprintf(w, returnValue)
	}
}

func WebStart(app Application) {
	http.HandleFunc("/storepubkey", app.storePubKey) // 修改了这里
	http.HandleFunc("/getpubkey", app.getPubKey) // 修改了这里
	http.HandleFunc("/storeipfsaddress", app.storeIPFSAddress) // 修改了这里
	http.HandleFunc("/getipfsaddress", app.getIPFSAddress) // 修改了这里

	
	fmt.Println("Starting server on port 9000...")
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}