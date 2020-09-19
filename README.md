### 跨链合约面向业务链合约的接口

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
