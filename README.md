## 跨链合约接口说明

### 1. 跨链合约面向业务链合约的接口

#### 跨链发票查询接口

InterchainSingleQuery

```go
{"InterchainSingleQuery", //type: 跨链发票查询接口
 "chainID-dajsdnfjasfasdf",//目的链的ID
 "key-12345678",//需要查询的发票的ID
}
```

#### 跨链发票归集接口

InterchainMultiQuery

```go
{"InterchainMultiQuery", //type: 跨链发票归集
 "queryByGhf",//归集方式
 "company-ghf12345678",//归集关键字
}
```

#### 跨链单链发票报销接口

InterchainSingleModify

```go
{"InterchainSingleModify", //type: 跨链单链发票报销
 "chainID-dajsdnfjasfasdf",//目的链的ID
 "key-12345678",//需要修改的发票的ID
 "value-[fphm:213123,fpdm:1238123,jym:666666]",//需要修改的发票的更新数据
}
```

#### 跨链双链发票报销接口

InterchainDoubleModify

```go
{"InterchainDoubleModify", //type: 跨链双链发票报销
 "chainID-dajsdnfjasfasdf",//目的链的ID
 "key-12345678",//本链修改的发票的ID
 "value-[fphm:213123,fpdm:1238123,jym:666666]",//本链需要修改的发票的更新数据
 "key-87654321",//目的链修改的发票的ID
 "value-[fphm:113123,fpdm:2238123,jym:777777]",//目的链
}
```

### 2. 跨链合约面向PAPP的调用接口

#### 保存链码私钥

savePrivateKey

```go
{"savePrivateKey", // type: 保存链码私钥
 "privateKey",// 私钥，用于PAPP验证链码身份
}
```

#### 修改PAPP的IP地址

modifyPAPPIP

```go
{"modifyPAPPIP", // type: 修改PAPP的IP地址
 "ip",// PAPP的IP地址
}
```

#### 跨链查询接口

interchainGet

```go
{"interchainGet", // type: 跨链查询接口
 "key",// 查询的key
}
```

#### 跨链写入接口

interchainSet

```go
{"interchainSet", // type: 跨链写入接口
 "key",  // 写入的key
 "value",// 写入的value
}
```

#### 跨链归集接口

interchainQueryByValue

```go
{"interchainQueryByValue", // type: 跨链归集接口
 "value",// 归集的关键词
}
```

#### 跨链函数调用接口

interchainFuncCall

```go
{"interchainFuncCall", // type: 跨链函数调用接口
 "funcName",  // 调用的函数名
 "args", // 调用函数时的参数
}
```

#### 跨链函数调用接口

interchainFuncCall

```go
{"interchainFuncCall", // type: 跨链函数调用接口
 "funcName",  // 调用的函数名
 "args", // 调用函数时的参数
}
```

#### 事件获取接口

pollingEvent

```go
{"pollingEvent", // type: 事件获取接口
 "pappEvents", // PAPP存储的跨链请求(事件)中每条链的最新index，格式如下：
 //      `[{"chainA":"1"},{"chainB":"3"},{"chainC":"1"}]`
}
```