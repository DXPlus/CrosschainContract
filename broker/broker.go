/*-------------------------------------------*/
/*            跨链合约入口 broker.go           */
/*-------------------------------------------*/
package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/common/util"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	interchainEventName = "interchain-event-name"
	innerMeta           = "inner-meta"
	outterMeta          = "outter-meta"
	chaincodeID         = "mycc"                     // 修改为本链链码的名称
	channelID           = "mychannel"                // 修改为本链通道的名称
	PrivateKey          = "private-key"
	PAPPIP              = "PAPP-IP-address"
)

type Broker struct{}

// 定义跨链请求的数据结构
type CrossChainRequest struct {
	DstChainID    string `json:"dstChainID"`        //目的链ID
	Func          string `json:"func"`              //请求目的
	Args        []string `json:"args"`              //请求参数
}

// 定义跨链合约与PAPP通信的消息结构
type RequestToPAPP struct {
	CCRequest   CrossChainRequest `json:"cc_request"`        //跨链请求
	SigR        []byte            `json:"sig_r"`             //对请求的签名
	SigS        []byte            `json:"sig_s"`             //对请求的签名
}

// 链码初始化函数
func (broker *Broker) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return broker.initialize(stub)
}

// 链码调用入口
func (broker *Broker) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Printf("invoke: %s\n", function)
	switch function {
	/*--------------------------------------*/
	/*               业务链调用              */
	/*--------------------------------------*/
	case "get":
		return broker.get(stub, args)
	case "set":
		return broker.set(stub, args)
	case "InterchainSingleQuery":
		return broker.InterchainSingleQuery(stub, args)
	case "InterchainMultiQuery":
		return broker.InterchainMultiQuery(stub, args)
	case "InterchainSingleModify":
		return broker.InterchainSingleModify(stub, args)
	case "InterchainDoubleModify":
		return broker.InterchainDoubleModify(stub, args)
	/*--------------------------------------*/
	/*                PAPP调用              */
	/*--------------------------------------*/
	case "InterchainVerify":
		return broker.InterchainVerify(stub, args)
	case "savePrivateKey":
		return broker.savePrivateKey(stub, args)
	case "modifyPAPPIP":
		return broker.modifyPAPPIP(stub,args)
	case "interchainGet":
		return broker.interchainGet(stub, args)
	case "interchainSet":
		return broker.interchainSet(stub, args)
	/*--------------------------------------*/
	/*        系统管理员调用-跨链历史查询       */
	/*--------------------------------------*/
	case "getInnerMeta":
		return broker.getInnerMeta(stub)
	case "getOuterMeta":
		return broker.getOuterMeta(stub)
	case "getInMessage":
		return broker.getInMessage(stub, args)
	case "getOutMessage":
		return broker.getOutMessage(stub, args)

	default:
		return shim.Error("invalid function: " + function + ", args: " + strings.Join(args, ","))
	}
}

// init
func (broker *Broker) initialize(stub shim.ChaincodeStubInterface) pb.Response {
	inCounter := make(map[string]uint64)
	outCounter := make(map[string]uint64)

	// 对innerMeta、outterMeta在链码存储里进行初始化
	if err := broker.putMap(stub, innerMeta, inCounter); err != nil {
		return shim.Error(err.Error())
	}

	if err := broker.putMap(stub, outterMeta, outCounter); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// get
func (broker *Broker) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("incorrect number of arguments, expecting 1")
	}
	key := args[0]
	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(value)
}

// set
func (broker *Broker) set(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("incorrect number of arguments, expecting 2")
	}
	key := args[0]
	value := args[1]
	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error(fmt.Errorf("set error: %w", err).Error())
	}
	return shim.Success(nil)
}

/*----------------------------------------------------------*/
/*                      业务链调用接口实现                     */
/*----------------------------------------------------------*/

// 跨链单链查询
func (broker *Broker) InterchainSingleQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("incorrect number of arguments, expecting 2")
	}

	dstChainID := args[0] // 目的链ID
	key := args[1]        // 查询的key值

	// 1 获取本链的私钥
	privateKeyText, err := stub.GetState(PrivateKey)
	if err != nil{
		return shim.Error(err.Error())
	}
	// 2 生成跨链请求
	ccRequest := CrossChainRequest{
		DstChainID: dstChainID,
		Func: "InterchainSingleQuery",
		Args: []string{dstChainID, key},
	}
	ccRJson, err := json.Marshal(ccRequest)
	if err != nil {
		fmt.Println("json marshal error: ", err)
	}
	// 3 生成签名
	Sha1Inst := sha1.New()
	Sha1Inst.Write(ccRJson)
	sourceData := Sha1Inst.Sum([]byte(""))
	signR,signS := EccSign(privateKeyText,sourceData)
	// 4 生成与PAPP通信的请求
	req := RequestToPAPP{
		CCRequest:       ccRequest,
		SigR:            signR,
		SigS:            signS,
	}

	return broker.InterchainRequestByHttp(stub, req)
}

// 跨链多链查询
func (broker *Broker) InterchainMultiQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("incorrect number of arguments, expecting 2")
	}

	queryBy := args[0]    //归集方式
	queryKey := args[1]   //归集关键词

	// 1 获取本链的私钥
	privateKeyText, err := stub.GetState(PrivateKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	// 2 生成跨链请求
	ccRequest := CrossChainRequest{
		DstChainID: "",
		Func: "InterchainMultiQuery",
		Args: []string{queryBy, queryKey},
	}
	ccRJson, err := json.Marshal(ccRequest)
	if err != nil {
		fmt.Println("json err: ", err)
	}
	// 3 生成签名
	Sha1Inst := sha1.New()
	Sha1Inst.Write(ccRJson)
	sourceData := Sha1Inst.Sum([]byte(""))
	signR,signS := EccSign(privateKeyText,sourceData)
	// 4 生成与PAPP通信的请求
	req := RequestToPAPP{
		CCRequest:       ccRequest,
		SigR:            signR,
		SigS:            signS,
	}

	return broker.InterchainRequestByHttp(stub, req)
}

// 跨链单链写入
func (broker *Broker) InterchainSingleModify(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		return shim.Error("incorrect number of arguments, expecting 3")
	}

	dstChainID:= args[0]
	key := args[1]
	value := args[2]

	ccRequest := CrossChainRequest{
		DstChainID: dstChainID,
		Func: "InterchainSingleModify",
		Args: []string{dstChainID, key, value},
	}

	req := RequestToPAPP{
		CCRequest:       ccRequest,
		SigR:            []byte(""),
		SigS:            []byte(""),
	}

	return broker.InterchainRequestBySetEvent(stub, req)
}

// 跨链双链同步写入
func (broker *Broker) InterchainDoubleModify(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 4 {
		return shim.Error("incorrect number of arguments, expecting 4")
	}

	dstChainID:= args[0]
	key1 := args[1]
	value1 := args[2]
	key2 := args[3]
	value2 := args[4]

	// 生成跨链请求
	ccRequest := CrossChainRequest{
		DstChainID: dstChainID,
		Func: "InterchainDoubleModify",
		Args: []string{dstChainID, key1, value1, key2, value2},
	}
	// 生成与PAPP通信的请求
	req := RequestToPAPP{
		CCRequest:       ccRequest,
		SigR:            []byte(""),
		SigS:            []byte(""),
	}

	return broker.InterchainRequestBySetEvent(stub, req)
}

/*----------------------------------------------------------*/
/*                       PAPP调用接口实现                     */
/*----------------------------------------------------------*/

// A-PAPP验证B-PAPP交易的真实性
func (broker *Broker) InterchainVerify(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	return shim.Success(nil)
}

// 保存PAPP给链码颁发的证书信息
func (broker *Broker) savePrivateKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("incorrect number of arguments, expecting 1")
	}

	privateKey := args[0]

	if err := stub.PutState(PrivateKey, []byte(privateKey)); err != nil {
		return shim.Error(fmt.Errorf("save private key error: %w", err).Error())
	}
	return shim.Success(nil)
}

// 修改PAPP的IP地址
func (broker *Broker) modifyPAPPIP(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("incorrect number of arguments, expecting 1")
	}

	ip := args[0]

	if err := stub.PutState(PAPPIP, []byte(ip)); err != nil {
		return shim.Error(fmt.Errorf("modify PAPPIP error: %w", err).Error())
	}
	return shim.Success(nil)
}

// 查询业务链数据
func (broker *Broker) interchainGet(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("incorrect number of arguments, expecting 1")
	}

	key := args[0]

	b := util.ToChaincodeArgs("interchainGet", key)
	response := stub.InvokeChaincode(chaincodeID, b, channelID)
	if response.Status != shim.OK {
		return shim.Error(fmt.Sprintf("invoke chaincode '%s' err: %s",chaincodeID , response.Message))
	}

	return shim.Success(response.Payload)
}

// 修改业务链数据
func (broker *Broker) interchainSet(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("incorrect number of arguments, expecting 2")
	}

	key := args[0]
	value := args[1]

	b := util.ToChaincodeArgs("interchainSet", key,value)
	response := stub.InvokeChaincode(chaincodeID, b, channelID)
	if response.Status != shim.OK {
		return shim.Error(fmt.Sprintf("invoke chaincode '%s' err: %s",chaincodeID , response.Message))
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(Broker))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s .....", err)
	}
}