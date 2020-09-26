/*-------------------------------------------*/
/*            跨链合约入口 broker.go           */
/*-------------------------------------------*/
package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/common/util"
	"strconv"
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
	case "setPrivateKey":
		return broker.setPrivateKey(stub, args)
	case "modifyPAPPIP":
		return broker.modifyPAPPIP(stub,args)
	case "interchainGet":
		return broker.interchainGet(stub, args)
	case "interchainSet":
		return broker.interchainSet(stub, args)
	case "interchainQueryByValue":
		return broker.interchainQueryByValue(stub, args)
	case "interchainFuncCall":
		return broker.interchainFuncCall(stub, args)
	case "pollingEvent":
		return broker.pollingEvent(stub, args)
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
	// 3 生成签名  各个节点生成的签名是不一样的，但是都会验证通过
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

// 保存PAPP给链码颁发的证书信息
func (broker *Broker) setPrivateKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

// 调用业务链归集接口
func (broker *Broker) interchainQueryByValue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("incorrect number of arguments, expecting 1")
	}

	value := args[0]// 归集关键词

	b := util.ToChaincodeArgs("queryByValue",value)
	response := stub.InvokeChaincode(chaincodeID, b, channelID)
	if response.Status != shim.OK {
		return shim.Error(fmt.Sprintf("invoke chaincode '%s' err: %s",chaincodeID , response.Message))
	}

	return shim.Success(response.Payload)
}

// 跨链合约调用业务合约函数的通用接口
func (broker *Broker) interchainFuncCall(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// args[0]   调用函数名
	// args[1:]  调用函数时的参数args
	//funcName := args[0] // 调用函数名
	//funcParameters := append(args[:0], args[1:]...)
	//b := util.ToChaincodeArgs(funcName,funcParameters)
	b := util.ArrayToChaincodeArgs(args)
	response := stub.InvokeChaincode(chaincodeID, b, channelID)
	if response.Status != shim.OK {
		return shim.Error(fmt.Sprintf("invoke chaincode '%s' err: %s",chaincodeID , response.Message))
	}

	return shim.Success(response.Payload)
}


// 根据PAPP当前储存的事件数据来获取最新事件的数据
func (broker *Broker) pollingEvent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("incorrect number of arguments, expecting 1")
	}
	// `[{"chainA":"1"},{"chainB":"3"},{"chainC":"1"}]`
	pappEvents := []byte(args[0])

	m := make(map[string]uint64)
	if err := json.Unmarshal(pappEvents, &m); err != nil {
		return shim.Error(fmt.Errorf("unmarshal out meta: %s", err).Error())
	}
	//`[{"chainA":"1"},{"chainB":"4"},{"chainC":"1"},{"chainD":"1"}]`
	outMeta, err := broker.getMap(stub, outterMeta)
	if err != nil {
		return shim.Error(err.Error())
	}

	events := make([]*CrossChainRequest, 0)
	for addr, idx := range outMeta {
		// 用outterMeta的chainID来查m对应的index，如果没有查到就是设置为0
		startPos, ok := m[addr]
		if !ok {
			startPos = 0
		}
		for i := startPos + 1; i <= idx; i++ {
			eb, err := stub.GetState(broker.outMsgKey(addr, strconv.FormatUint(i, 10)))
			if err != nil {
				fmt.Printf("get out event by key %s fail", broker.outMsgKey(addr, strconv.FormatUint(i, 10)))
				continue
			}
			e := &CrossChainRequest{}
			if err := json.Unmarshal(eb, e); err != nil {
				fmt.Println("unmarshal event fail")
				continue
			}
			events = append(events, e)
		}
	}

	ret, err := json.Marshal(events)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ret)
}


func main() {
	err := shim.Start(new(Broker))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s .....", err)
	}
}