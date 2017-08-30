/*
Copyright Marlabs. 2017 All Rights Reserved.
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"strings"
	"encoding/pem"
	"crypto/x509"
)

var logger = shim.NewLogger("FileaccessChaincode")

const indexName = `Fileaccess`

type FileaccessValue struct {
	Hash  	    	string 				`json:"hash"`
	Acl				[]string			`json:"acl"`
}

type Fileaccess struct {
	Owner	        string				`json:"owner"`
	Filename        string 				`json:"filename"`
	Hash      		string 				`json:"hash"`
	Acl				[]string			`json:"acl"`
}

type FileaccessChaincode struct {
}

func (t *FileaccessChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *FileaccessChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "put" {
		return t.put(stub, args)
	//} else if function == "access" {
	//	return t.access(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:400, Message:"unknown invoke function"}
}

// creates or modifies record of file access
func (t *FileaccessChaincode) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var fileaccess Fileaccess
	err := json.Unmarshal([]byte(args[0]), &fileaccess)
	if err != nil {
		return pb.Response{Status:400, Message:"cannot unmarshal json arg"}
	}

	owner := "oleg"

	key, err := stub.CreateCompositeKey(indexName, []string{owner, fileaccess.Filename})
	if err != nil {
		return shim.Error(err.Error())
	}

	value, err := json.Marshal(FileaccessValue{Hash: fileaccess.Hash, Acl: fileaccess.Acl})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *FileaccessChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return pb.Response{Status:400, Message:"missing arg filename"}
	}

	filename := args[0]

	owner := "oleg"

	fileaccess, err := t.findByKey(stub, owner, filename)

	if err != nil {
		return pb.Response{Status:404, Message:"cannot findByKey"}
	}

	ret, err := json.Marshal(fileaccess)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ret)
}

func (t *FileaccessChaincode) findByKey(stub shim.ChaincodeStubInterface, owner string, filename string) (Fileaccess, error) {

	key, err := stub.CreateCompositeKey(indexName, []string{owner, filename})
	if err != nil {
		return Fileaccess{}, fmt.Errorf("Cannot create composite key: %v", err)
	}

	response, err := stub.GetState(key)
	if err != nil {
		return Fileaccess{}, fmt.Errorf("Cannot read the state: %v", err)
	}
	if response == nil {
		return Fileaccess{}, fmt.Errorf("No Fileaccess found for key: %v", key)
	}

	var value FileaccessValue
	err = json.Unmarshal(response, &value)
	if err != nil {
		return Fileaccess{}, fmt.Errorf("Cannot Unmarshal Fileaccess: %v", err)
	}

	fileaccess := Fileaccess {
		Owner: owner,
		Filename: filename,
		Hash: value.Hash,
		Acl: value.Acl,
	}

	return fileaccess, nil
}

func getCommonName(certificate []byte) string {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName)

	return commonName
}

func main() {
	err := shim.Start(new(FileaccessChaincode))
	if err != nil {
		logger.Error(fmt.Errorf("Error starting chaincode: %s", err))
	}
}
