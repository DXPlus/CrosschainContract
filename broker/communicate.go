/*-------------------------------------------*/
/*            通信模块 communicate.go         */
/*-------------------------------------------*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// 通过SetEvent发送跨链请求
func (broker *Broker) InterchainRequestBySetEvent(stub shim.ChaincodeStubInterface, req RequestToPAPP) pb.Response {

	// 获取跨链记录的dstChainID和index
	destChainID := req.CCRequest.DstChainID
	outMeta, err := broker.getMap(stub, outterMeta)
	if err != nil {
		return shim.Error(err.Error())
	}

	if _, ok := outMeta[destChainID]; !ok {
		outMeta[destChainID] = 0
	}

	// index++后写入
	outMeta[destChainID]++
	if err := broker.putMap(stub, outterMeta, outMeta); err != nil {
		return shim.Error(err.Error())
	}

	// 序列化请求
	reqData, err := json.Marshal(req)
	if err != nil {
		return shim.Error(err.Error())
	}
	ccReq,err:=json.Marshal(req.CCRequest)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 生成每条跨链记录的唯一key
	key := broker.outMsgKey(destChainID,strconv.FormatUint(outMeta[destChainID], 10))
	// 保存跨链记录
	if err := stub.PutState(key, ccReq); err != nil {
		return shim.Error(fmt.Errorf("save request record error: %w", err).Error())
	}

    // setEvent触发事件
	if err := stub.SetEvent(interchainEventName, reqData); err != nil {
		return shim.Error(fmt.Errorf("set event error: %w", err).Error())
	}
	return shim.Success(nil)
}

// 通过Http发送跨链请求并接收返回数据
func (broker *Broker) InterchainRequestByHttp(stub shim.ChaincodeStubInterface, req RequestToPAPP) pb.Response {

	// 获取PAPP的IP地址
	IP, err := stub.GetState(PAPPIP)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 获取跨链记录的dstChainID和index
	destChainID := req.CCRequest.DstChainID
	outMeta, err := broker.getMap(stub, outterMeta)
	if err != nil {
		return shim.Error(err.Error())
	}

	if _, ok := outMeta[destChainID]; !ok {
		outMeta[destChainID] = 0
	}

	// index++后写入
	outMeta[destChainID]++
	if err := broker.putMap(stub, outterMeta, outMeta); err != nil {
		return shim.Error(err.Error())
	}

	// 序列化请求
	reqData, err := json.Marshal(req)
	if err != nil {
		return shim.Error(err.Error())
	}
	ccReq,err:=json.Marshal(req.CCRequest)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 生成每条跨链记录的唯一key
	key := broker.outMsgKey(destChainID,strconv.FormatUint(outMeta[destChainID], 10))
	// 保存跨链记录
	if err := stub.PutState(key, ccReq); err != nil {
		return shim.Error(fmt.Errorf("save request record error: %w", err).Error())
	}

	// 发送http.post请求
	returnData,err := broker.SendRep(string(IP), string(reqData))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(returnData))
}

//// 非跨链操作请求，通过Http发送请求并接收返回数据
//func (broker *Broker) RequestByHttp(stub shim.ChaincodeStubInterface, req RequestToPAPP) pb.Response {
//	// 获取PAPP的IP地址
//	IP, err := stub.GetState(PAPPIP)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//
//	reqData, err := json.Marshal(req)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//
//	// 发送http.post请求
//	returnData,err := broker.SendRep(string(IP), string(reqData))
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	return shim.Success([]byte(returnData))
//}
//
//// 非跨链操作请求，通过SetEvent发送请求
//func (broker *Broker) RequestBySetEvent(stub shim.ChaincodeStubInterface, req RequestToPAPP) pb.Response {
//	reqData, err := json.Marshal(req)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//
//	if err := stub.SetEvent(interchainEventName, reqData); err != nil {
//		return shim.Error(fmt.Errorf("set event error: %w", err).Error())
//	}
//	return shim.Success(nil)
//}

// 向服务器发送http.post请求
func (broker *Broker) SendRep(url string, data string) (string, error) {
	// 发送post请求
	res, err := http.Post(url, "application/json;charset=utf-8", strings.NewReader(data))
	if err != nil || res == nil {
		fmt.Println("Request error: ", err)
		return "Request error", err
	}

	defer res.Body.Close()

	// 读取服务器返回消息
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Fatal error: ", err)
		return "Fatal error", err
	}
	return string(content), err
}