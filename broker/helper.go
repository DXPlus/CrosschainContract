/*-------------------------------------------*/
/*            跨链辅助功能 helper.go           */
/*-------------------------------------------*/
package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// putMap
func (broker *Broker) putMap(stub shim.ChaincodeStubInterface, metaName string, meta map[string]uint64) error {
	if meta == nil {
		return nil
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return stub.PutState(metaName, metaBytes)
}

// getMap
func (broker *Broker) getMap(stub shim.ChaincodeStubInterface, metaName string) (map[string]uint64, error) {
	metaBytes, err := stub.GetState(metaName)
	if err != nil {
		return nil, err
	}

	meta := make(map[string]uint64)
	if metaBytes == nil {
		return meta, nil
	}

	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

/*-------------------------------------------*/
/*                  跨链历史记录模块           */
/*-------------------------------------------*/

// 获取当前链为来源链的最新跨链请求的序号，{目的链：序号}，如{B:2, C:6}
func (broker *Broker) getOuterMeta(stub shim.ChaincodeStubInterface) pb.Response {
	v, err := stub.GetState(outterMeta)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(v)
}

// 查询键值中dstChainID指定目的链，idx指定序号，查询结果为以Broker所在的区块链作为来源链的跨链请求
func (broker *Broker) getOutMessage(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("incorrect number of arguments, expecting 2.")
	}
	destChainID := args[0]
	sequenceNum := args[1]
	key := broker.outMsgKey(destChainID, sequenceNum)
	v, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(v)
}

// 获取当前链为目标链的最新跨链请求的序号，{来源链：序号}，如{B:3, C:5}
func (broker *Broker) getInnerMeta(stub shim.ChaincodeStubInterface) pb.Response {
	v, err := stub.GetState(innerMeta)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(v)
}

// 查询键值中dstChainID指定目的链，idx指定序号，查询结果为以Broker所在的区块链作为来源链的跨链请求
func (broker *Broker) getInMessage(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("incorrect number of arguments, expecting 2")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	key := broker.inMsgKey(sourceChainID, sequenceNum)
	v, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(v)
}

// 生成发起请求的key
func (broker *Broker) outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

// 生成接受请求的key
func (broker *Broker) inMsgKey(from string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", from, idx)
}

/*-------------------------------------------*/
/*            PAPP与跨链合约身份认证模块        */
/*-------------------------------------------*/

// ECC签名
func EccSign(privateKey []byte,sourceData []byte) ([]byte, []byte) {
	ECPrivateKey, err := x509.ParseECPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	hashText := sha1.Sum(sourceData)

	r, s, err := ecdsa.Sign(rand.Reader, ECPrivateKey, hashText[:])
	if err != nil {
		panic(err)
	}
	rText, err := r.MarshalText()
	if err != nil {
		panic(err)
	}
	sText, err := s.MarshalText()
	if err != nil {
		panic(err)
	}
	return rText, sText
}

//// ECC签名验证
//func EccVerify(plainText,rText,sText,publicKeyText []byte) bool {
//	pubInterface,err := x509.ParsePKIXPublicKey(publicKeyText)
//	if err != nil {
//		panic(err)
//	}
//
//	publicKey := pubInterface.(*ecdsa.PublicKey)
//
//	hashText := sha1.Sum(plainText)
//	var r,s big.Int
//	r.UnmarshalText(rText)
//	s.UnmarshalText(sText)
//	bl := ecdsa.Verify(publicKey,hashText[:],&r,&s)
//	return bl
//}


