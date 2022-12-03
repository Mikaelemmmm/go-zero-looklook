<h1>Table of Contents</h1>

- [IX. distributed transactions](#ix-distributed-transactions)
	- [1. First of all, you need to pay attention to](#1-first-of-all-you-need-to-pay-attention-to)
	- [2. clone dtm](#2-clone-dtm)
	- [3. Configuration file](#3-configuration-file)
	- [4. Start dtm server](#4-start-dtm-server)
	- [5. Using go-zero's grpc to dock dtm](#5-using-go-zeros-grpc-to-dock-dtm)
		- [5.1. order-api](#51-order-api)
		- [5.2. order-srv](#52-order-srv)
			- [5.2.1. Create](#521-create)
			- [5.2.2. CreateRollback](#522-createrollback)
		- [5.3. stock-srv](#53-stock-srv)
			- [5.3.1. Deduct](#531-deduct)
			- [5.3.2. DeductRollback](#532-deductrollback)
	- [6. Sub-transaction barrier](#6-sub-transaction-barrier)
	- [7. Notes in go-zero docking](#7-notes-in-go-zero-docking)
		- [7.1. dtm's rollback compensation](#71-dtms-rollback-compensation)
		- [7.2. barrier's empty compensation, suspension, etc](#72-barriers-empty-compensation-suspension-etc)
		- [7.3, barrier in the rpc local transactions](#73-barrier-in-the-rpc-local-transactions)
	- [8. using go-zero http docking](#8-using-go-zero-http-docking)

# IX. distributed transactions

Because the service division of this project is relatively independent, so currently no use to distributed transactions, but go-zero combined with dtm using distributed transactions best practices, I have organized demos, here to introduce the use of go-zero combined with dtm, the project address go-zero combined with dtm best practices repository address : <https://github>. com/Mikaelemmmm/gozerodtm

[Note] The following is not go-zero-looklook project, is this project <https://github.com/Mikaelemmmm/gozerodtm>

## 1. First of all, you need to pay attention to

> go-zero 1.2.4 version or more, this must be noted

> dtm you use the latest on the line

## 2. clone dtm

```shell
git clone https://github.com/yedf/dtm.git
```

## 3. Configuration file

1.Find conf.sample.yml under the project and folder

2, cp conf.sample.yml conf.yml

3, using etcd, open the following comment in the configuration (if you do not use etcd, it is even easier, this is saved, direct link to the dtm server address on it)

```yaml
MicroService:
 Driver: 'dtm-driver-gozero' # name of the driver to handle register/discover
 Target: 'etcd://localhost:2379/dtmservice' # register dtm server to this url
 EndPoint: 'localhost:36790'
```

 Explain.

MicroService this do not move, this represents the registration of dtm to that microservice service cluster, so that the microservice cluster internal services can interact directly with dtm through grpc

Driver: 'dtm-driver-gozero', use go-zero's registered service discovery driver, support go-zero

Target: 'etcd://localhost:2379/dtmservice' register the current dtm server directly to the etcd cluster where the microservice is located, if go-zero is used as a microservice, it can directly get the server grpc link of dtm through etcd and directly interact with dtm server

 EndPoint: 'localhost:36790', which represents the connection address + port of dtm's server, the microservices in the cluster can get this address directly through etcd and interact with dtm.

If you change the dtm source grpc port yourself, remember to change the port here

## 4. Start dtm server

In the dtm project root directory

```shell
go run app/main.go dev
```

## 5. Using go-zero's grpc to dock dtm

This is an example of a quick order to deduct product inventory

### 5.1. order-api

order-api is an http service portal to create orders

```go
service order {
   @doc "create order"
   @handler create
   post /order/quickCreate (QuickCreateReq) returns (QuickCreateResp)
}
```

Next, look at the logic

```go
func (l *CreateLogic) Create(req types.QuickCreateReq,r *http.Request) (*types.QuickCreateResp, error) {

 orderRpcBusiServer, err := l.svcCtx.Config.OrderRpcConf.BuildTarget()
 if err != nil{
  return nil,fmt.Errorf("create order timeout")
 }
 stockRpcBusiServer, err := l.svcCtx.Config.StockRpcConf.BuildTarget()
 if err != nil{
  return nil,fmt.Errorf("create order timeout")
 }

 createOrderReq:= &order.CreateReq{UserId: req.UserId,GoodsId: req.GoodsId,Num: req.Num}
 deductReq:= &stock.DecuctReq{GoodsId: req.GoodsId,Num: req.Num}

 // Here is only the saga example, tcc and other examples basically no difference specific can see dtm official website

 gid := dtmgrpc.MustGenGid(dtmServer)
 saga := dtmgrpc.NewSagaGrpc(dtmServer, gid).
  Add(orderRpcBusiServer+"/pb.order/create", orderRpcBusiServer+"/pb.order/createRollback", createOrderReq).
  Add(stockRpcBusiServer+"/pb.stock/deduct", stockRpcBusiServer+"/pb.stock/deductRollback", deductReq)

 err = saga.Submit()
 dtmimp.FatalIfError(err)
 if err != nil{
  return nil,fmt.Errorf("submit data to  dtm-server err  : %+v \n",err)
 }

 return &types.QuickCreateResp{}, nil
}
```

When you enter the order logic, get the order order, stock stock service rpc address in etcd respectively, and use the method BuildTarget()

Then create order, stock corresponding request parameters

Request dtm to get the global transaction id, based on this global transaction id open grpc saga distributed transactions, create orders, deduct inventory requests into the transaction, here using the grpc form request, each business to have a forward request, a rollback request, and request parameters, when the implementation of any one of the business forward request error will automatically call the transaction of all business rollback request to achieve the rollback effect.

### 5.2. order-srv

order-srv is the order rpc service that interacts with the order table in the dtm-gozero-order database

```protobuf
//service
service order {
   rpc create(CreateReq)returns(CreateResp);
   rpc createRollback(CreateReq)returns(CreateResp);
}
```

#### 5.2.1. Create

When the order-api commit transaction is requested by the create method by default, we look at the logic

```go
func (l *CreateLogic) Create(in *pb.CreateReq) (*pb.CreateResp, error) {

   fmt.Printf("Create Order in : %+v \n", in)

    //barrier to prevent empty compensation, empty suspension, etc. specific look at the dtm official website can be, do not forget to add the barrier table in the current library, because the judgment compensation with the sql to be executed together with local transactions
   barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
   db, err := sqlx.NewMysql(l.svcCtx.Config.DB.DataSource).RawDB()
   if err != nil {
      //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
      return nil, status.Error(codes.Internal, err.Error())
   }
   if err := barrier.CallWithDB(db, func(tx *sql.Tx) error {

      order := new(model.Order)
      order.GoodsId = in.GoodsId
      order.Num = in.Num
      order.UserId = in.UserId

      _, err = l.svcCtx.OrderModel.Insert(tx, order)
      if err != nil {
         return fmt.Errorf("fail err : %v , order:%+v \n", err, order)
      }

      return nil
   }); err != nil {
     //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
      return nil, status.Error(codes.Internal, err.Error())
   }

   return &pb.CreateResp{}, nil
}
```

As you can see, once inside the method we use dtm's subtransaction barrier technology, as to why the use of subtransaction barriers because there may be duplicate requests or empty requests caused by dirty data, etc., where dtm automatically gives us the idempotent processing does not require us to do additional, while ensuring that his internal idempotent processing and our own implementation of the transaction to be in a transaction, so to use a session of the db link, at this point we have to first obtain

```go
   db, err := sqlx.NewMysql(l.svcCtx.Config.DB.DataSource).RawDB()
```

Then based on this db connection dtm does idempotent processing internally through sql execution, while we open transactions based on this db connection, so that we can ensure that dtm's internal subtransaction barrier in the execution of sql operations and our own business execution of sql operations in a transaction.

When dtm uses grpc to call our business, our grpc service returns to dtm server error, dtm will determine whether to perform rollback operation or keep retrying based on the grpc error code we return to it.

- Internal: dtm server will not call rollback, will always retry, each retry dtm database will add a retry count, you can monitor this retry count alarm, manual processing
- Aborted : dtm server will call all rollback requests and perform rollback operations

If dtm returns an error of nil when calling grpc, the call is considered successful

#### 5.2.2. CreateRollback

Aborted : dtm server will call all rollback operations when we call the order creation order or inventory deduction when the codes.Aborted is returned to dtm server, CreateRollback is the rollback operation corresponding to the order placed, the code is as follows

```go
func (l *CreateRollbackLogic) CreateRollback(in *pb.CreateReq) (*pb.CreateResp, error) {



 order, err := l.svcCtx.OrderModel.FindLastOneByUserIdGoodsId(in.UserId, in.GoodsId)
 if err != nil && err != model.ErrNotFound {
  //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
 }

 if order != nil {

  barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
  db, err := l.svcCtx.OrderModel.SqlDB()
  if err != nil {
   //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
   return nil, status.Error(codes.Internal, err.Error())
  }
  if err := barrier.CallWithDB(db, func(tx *sql.Tx) error {

   order.RowState = -1
   if err := l.svcCtx.OrderModel.Update(tx, order); err != nil {
    return fmt.Errorf("rollback fail  err : %v , userId:%d , goodsId:%d", err, in.UserId, in.GoodsId)
   }

   return nil
  }); err != nil {
   logx.Errorf("err : %v \n", err)

//!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
   return nil, status.Error(codes.Internal, err.Error())
  }

 }
 return &pb.CreateResp{}, nil
}

```

In fact, if the previous order was successful, the previous successful order to cancel is the corresponding order rollback operation

### 5.3. stock-srv

#### 5.3.1. Deduct

Deduct inventory, here and order Create the same, is the order transaction within the positive operation, deduction of inventory, the code is as follows

```go
func (l *DeductLogic) Deduct(in *pb.DecuctReq) (*pb.DeductResp, error) {

 stock, err := l.svcCtx.StockModel.FindOneByGoodsId(in.GoodsId)
 if err != nil && err != model.ErrNotFound {
  //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
 }
 if stock == nil || stock.Num < in.Num {
 //[Rollback] Insufficient inventory to determine the need for dtm direct rollback, return codes.Aborted, dtmcli.ResultFailure directly before you can rollback
  return nil, status.Error(codes.Aborted, dtmcli.ResultFailure)
 }

 //barrier to prevent empty compensation, empty suspension, etc. specific look at the dtm official website can be, do not forget to add the barrier table in the current library, because the judgment compensation with the sql to be executed together with local transactions
 barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
 db, err := l.svcCtx.StockModel.SqlDB()
 if err != nil {
  //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
 }
 if err := barrier.CallWithDB(db, func(tx *sql.Tx) error {
  sqlResult,err := l.svcCtx.StockModel.DecuctStock(tx, in.GoodsId, in.Num)
  if err != nil{
   //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
   return status.Error(codes.Internal, err.Error())
  }
  affected, err := sqlResult.RowsAffected()
  if err != nil{
   //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
   return status.Error(codes.Internal, err.Error())
  }

   // If it is affecting the number of lines to 0, directly tell dtm failed not to retry
  if affected <= 0 {
   return  status.Error(codes.Aborted,  dtmcli.ResultFailure)
  }

  //!!! Turn on the test!!! : Test order rollback change status is invalid, and the current library buckle failure does not need to be rolled back
  //return fmt.Errorf("Deductive inventory failed err : %v , in:%+v \n",err,in)

  return nil
 }); err != nil {
  //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil,err
 }

 return &pb.DeductResp{}, nil
}

```

It is worth noting here is that when only insufficient inventory, or in the deduction of inventory affects the number of lines to 0 (unsuccessful) only need to tell the dtm server to roll back, other cases are actually network jitter, hardware anomalies caused by, should let the dtm server has been retrying, of course, they should add a maximum number of retries to monitor the alarm, if the maximum number of times still unsuccessful can be achieved automatically send SMS, call the manual intervention. If it reaches the maximum number of unsuccessful attempts, it can automatically send SMS, call manual intervention.

#### 5.3.2. DeductRollback

Here is the rollback operation corresponding to the deduction of inventory

```go
func (l *DeductRollbackLogic) DeductRollback(in *pb.DecuctReq) (*pb.DeductResp, error) {

 fmt.Printf("Inventory in : %+v \n", in)

 barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
 db, err := l.svcCtx.StockModel.SqlDB()
 if err != nil {
  //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
 }
 if err := barrier.CallWithDB(db, func(tx *sql.Tx) error {
  if err := l.svcCtx.StockModel.AddStock(tx, in.GoodsId, in.Num); err != nil {
   return fmt.Errorf("rollback store fail err : %v ,goodsId:%d , num :%d", err, in.GoodsId, in.Num)
  }
  return nil
 }); err != nil {
  logx.Errorf("err : %v \n", err)
  //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
 }

 return &pb.DeductResp{}, nil
}

```

## 6. Sub-transaction barrier

This term is defined by the author of dtm, in fact, the subtransaction barrier code is not much, just look at the method barrier.CallWithDB.

```go
// CallWithDB the same as Call, but with *sql.DB
func (bb *BranchBarrier) CallWithDB(db *sql.DB, busiCall BarrierBusiFunc) error {
  tx, err := db.Begin()
  if err != nil {
    return err
  }
  return bb.Call(tx, busiCall)
}
```

As this method he opens the local transaction inside, it is in this transaction is executed sql operation, so when we execute our own business must use the same transaction with it, that is based on the same db connection open transaction, so ~ you know why we want to get db connection in advance, right, the purpose is to make it internal execution of sql operation with our sql operation in a transaction The purpose is to make the sql operation it executes internally and our sql operation under one transaction. As for why it executes its own sql operations internally, let's analyze it next.

Let's look at the method bb.Call

```go
// Call subtransaction barrier, see https://zhuanlan.zhihu.com/p/388444465 for details
// tx: transaction object of the local database, allowing the subtransaction barrier to perform transaction operations
// busiCall: business function to be called only when necessary
func (bb *BranchBarrier) Call(tx *sql.Tx, busiCall BarrierBusiFunc) (rerr error) {
 bb.BarrierID = bb.BarrierID + 1
 bid := fmt.Sprintf("%02d", bb.BarrierID)
 defer func() {
  // Logf("barrier call error is %v", rerr)
  if x := recover(); x != nil {
   tx.Rollback()
   panic(x)
  } else if rerr != nil {
   tx.Rollback()
  } else {
   tx.Commit()
  }
 }()
 ti := bb
 originType := map[string]string{
  BranchCancel:     BranchTry,
  BranchCompensate: BranchAction,
 }[ti.Op]

 originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.Op)
 currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.Op, bid, ti.Op)
 dtmimp.Logf("originAffected: %d currentAffected: %d", originAffected, currentAffected)
 if (ti.Op == BranchCancel || ti.Op == BranchCompensate) && originAffected > 0 || // This is empty compensation
  currentAffected == 0 { // This is a duplicate request or suspension
  return
 }
 rerr = busiCall(tx)
 return
}
```

The core is actually the following lines of code

```go
originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.Op)
currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.Op, bid, ti.Op)
dtmimp.Logf("originAffected: %d currentAffected: %d", originAffected, currentAffected)
if (ti.Op == BranchCancel || ti.Op == BranchCompensate) && originAffected > 0 || // This is empty compensation
  currentAffected == 0 { // This is a duplicate request or suspension
  return
}
rerr = busiCall(tx)
```

```go
func insertBarrier(tx DB, transType string, gid string, branchID string, op string, barrierID string, reason string) (int64, error) {  if op == "" {
   return 0, nil
 }
                                                                                                                                      sql := dtmimp.GetDBSpecial().GetInsertIgnoreTemplate("dtm_barrier.barrier(trans_type, gid, branch_id, op, barrier_id, reason) values(?,?,?,?,?,?)", "uniq_barrier")
                                                                                                                                     return dtmimp.DBExec(tx, sql, transType, gid, branchID, op, barrierID, reason)                                                                                                          }
```

Op default normal execution operation is action, so the normal first request is ti.Op value is action, that originType is ""

```go
 originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.Op)
```

Then the above sql will not be executed because ti.Op == "" in the insertBarrier directly return

```go
 currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.Op, bid, ti.Op)
```

Op of the second sql is action, so the subtransaction barrier table barrier will insert a data

Similarly, in the execution of the inventory will also insert a

> 1. Sub-transaction barrier where the whole transaction is successful

Op is action, so the originType is "" , so whether it is the barrier of the order or the barrier of the inventory deduction, the originAffected will be ignored when executing their two barrier inserts, because originType=="" will be directly returned without inserting data, so it seems that whether it is an order or deducted inventory, the second barrier insert data into effect, so the barrier data table will have two order data, one is the order and one is deducted inventory

![image-20220128131900853](../chinese/images/9/image-20220128131900853.png)

gid : dtm global transaction id

branch_id : the id of each operation under each global transaction id

op : operation, if it is a normal successful request is action

barrier_id : multiple openings under the same operation will increment

These four fields in the table is the joint unique so hidden, in insertBarrier time, dtm judgment if there is ignored not inserted

> 2. if the order success inventory is not enough to roll back the subtransaction barrier

We have only 10 inventory , we place an order for 20

(1) when the order is placed successfully, because the order is placed when the subsequent inventory is not known (instantly when the order is placed first to check the inventory that there will also be sufficient when the query, deducted when the shortage), so the order is placed successfully barrier table according to the following

So the success of the order barrier table in accordance with the logic previously sorted out will produce a correct data execution data in the barrier table

![image-20220128171950826](../chinese/images/9/image-20220128171950826.png)

2) Then execute the inventory deduction operation

```go

func (l *DeductLogic) Deduct(in *pb.DecuctReq) (*pb.DeductResp, error) {

 stock, err := l.svcCtx.StockModel.FindOneByGoodsId(in.GoodsId)
 if err != nil && err != model.ErrNotFound {
 //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
 }
 if stock == nil || stock.Num < in.Num {
    //[Rollback] Insufficient inventory to determine the need for dtm direct rollback, return codes.Aborted, dtmcli.ResultFailure directly before you can rollback
  return nil, status.Error(codes.Aborted, dtmcli.ResultFailure)
 }

  .......
}
```

Aborted will not go to the subtransaction barrier barrier here, so the barrier table will not insert data, but tell dtm to roll back

3) Call order rollback operation

Order rollback will open the barrier, which will then execute the barrier code (as follows), due to the rollback code ti.Op is compensate, orginType is action

```go
originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.Op)currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.Op, bid, ti.Op)dtmimp.Logf("originAffected: %d currentAffected: %d", originAffected, currentAffected)
if (ti.Op == BranchCancel || ti.Op == BranchCompensate) && originAffected > 0 ||  // This is empty compensation
currentAffected == 0 { // This is a duplicate request or suspension
  return
}
rerr = busiCall(tx)
```

Since our previous order was successful, the barrier table has a record action when the order was successful, so originAffected==0, so only one current rollback record will be inserted to continue to call busiCall(tx) to perform the subsequent rollback operation we wrote ourselves

![image-20220128172025283](../chinese/images/9/image-20220128172025283.png)

At this point, we should only have 2 pieces of data, an order creation record and an order rollback record

4) Inventory Rollback DeductRollback

After the order is successfully rolled back, it will continue to call the inventory rollback DeductRollback, the inventory rollback code is as follows

This is what the subtransaction barrier automatically helps us to determine, that is, the 2 core insert statements help us to determine, so that our business will not appear dirty data

Inventory rollback here is divided into 2 cases

- No successful rollback

- Rollback with a successful deduction

> rollback without successful deduction (our current example scenario is this one)

Op is compensate, orginType is action, the following 2 insert statements will be executed

```go
originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.Op)currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.Op, bid, ti.Op)dtmimp.Logf("originAffected: %d currentAffected: %d", originAffected, currentAffected)if (ti.Op == BranchCancel || ti.Op == BranchCompensate) && originAffected > 0 ||  currentAffected == 0 {  return}rerr = busiCall(tx)
```

Here combined with the judgment if the rollback, cancel the operation, originAffected > 0 current inserted successfully, before the corresponding positive deduction of inventory operations were not inserted successfully, indicating that the previous inventory was not deducted successfully, the direct return will not need to perform subsequent compensation. So at this time will be inserted in the barrier table 2 data directly return, will not perform our subsequent compensation operations

![image-20220128171756108](../chinese/images/9/image-20220128171756108.png)

At this point we have 4 entries in the barrier table

> deduction successfully rolled back (this situation yourself can try to simulate this scenario)

Op is compensate, orginType is action, continue to execute 2 insert statements if our previous step to deduct inventory successfully, in the implementation of this compensation ti.

```go
originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.Op)currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.Op, bid, ti.Op)dtmimp.Logf("originAffected: %d currentAffected: %d", originAffected, currentAffected)if (ti.Op == BranchCancel || ti.Op == BranchCompensate) && originAffected > 0 || currentAffected == 0 {  return}rerr = busiCall(tx)
```

Here combined with the judgment if the rollback, cancel the operation, originAffected == 0 current insertion ignored not inserted, indicating that the previous positive deduction of inventory inserted successfully, here only insert the second sql statement record can be, and then in the implementation of the subsequent business operations we compensate.

So, the overall analysis of the core statement is 2 insert, it helps us to solve the repeated rollback data, data power and other situations, I can only say that dtm author idea is really good, with a minimum of code to help us solve a very troublesome problem

## 7. Notes in go-zero docking

### 7.1. dtm's rollback compensation

When using dtm's grpc, when we use saga, tcc, etc., if the first step of the attempt or implementation failed, it is hoped that it can perform the rollback, the service in the grpc if an error occurs, must return: status.Error(codes.Aborted, dtmcli. ResultFailure), return other errors, will not perform your rollback operation, dtm will keep retrying, as follows.

```go
stock, err := l.svcCtx.StockModel.FindOneByGoodsId(in.GoodsId)
if err != nil && err != model.ErrNotFound {
    //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
}
if stock == nil || stock.Num < in.Num {
  //[Rollback] Insufficient inventory to determine the need for dtm direct rollback, return codes.Aborted, dtmcli.ResultFailure directly before you can rollback
  return nil, status.Error(codes.Aborted, dtmcli.ResultFailure)
}
```

### 7.2. barrier's empty compensation, suspension, etc

Before the preparation, we created the dtm_barrier library and the implementation of the barrier.mysql.sql, which is actually a check for our business services to prevent empty compensation, you can see the specific source code in the barrier.Call, not a few lines of code can be read.

If we use it online, each of your services interacting with the db as long as the barrier is used, the service uses to the mysql account, to assign him barrier library permissions, do not forget this

### 7.3, barrier in the rpc local transactions

In the rpc business, if the use of the barrier, then the interaction with the db in the model must be used when the transaction, and must be the same transaction with the barrier

logic

```go
barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
 db, err := sqlx.NewMysql(l.svcCtx.Config.DB.DataSource).RawDB()
 if err != nil {
  //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil, status.Error(codes.Internal, err.Error())
 }
 if err := barrier.CallWithDB(db, func(tx *sql.Tx) error {
  sqlResult,err := l.svcCtx.StockModel.DecuctStock(tx, in.GoodsId, in.Num)
  if err != nil{
   //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
   return status.Error(codes.Internal, err.Error())
  }
  affected, err := sqlResult.RowsAffected()
  if err != nil{
   //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
   return status.Error(codes.Internal, err.Error())
  }

  // If it is affecting the number of lines to 0, directly tell dtm failed not to retry
  if affected <= 0 {
   return  status.Error(codes.Aborted,  dtmcli.ResultFailure)
  }

  //!!! Turn on the test!!! : Test order rollback change status is invalid, and the current library buckle failure does not need to be rolled back
  //return fmt.Errorf("store fail err : %v , in:%+v \n",err,in)

  return nil
 }); err != nil {
   //!!! General database will not error does not need dtm rollback, let him keep retrying, this time do not return codes.Aborted, dtmcli.ResultFailure on it, the specific own control!!!!
  return nil,err
 }
```

model

```go
func (m *defaultStockModel) DecuctStock(tx *sql.Tx,goodsId , num int64) (sql.Result,error) {
 query := fmt.Sprintf("update %s set `num` = `num` - ? where `goods_id` = ? and num >= ?", m.table)
 return tx.Exec(query,num, goodsId,num)

}

func (m *defaultStockModel) AddStock(tx *sql.Tx,goodsId , num int64) error {
 query := fmt.Sprintf("update %s set `num` = `num` + ? where `goods_id` = ?", m.table)
 _, err :=tx.Exec(query, num, goodsId)
 return err
}
```

## 8. using go-zero http docking

This is basically not much difficulty, grpc will be this very simple, given that go in microservices to use http scenarios are not much, here do not go into detail, I have written a previous version of a simple, but not this perfect, interested to see, but that barrier is their own go-zero based on the sqlx, the official dtm will be modified, now do not need.

Project address: [https://github.com/Mikaelemmmm/dtmbarrier-go-zero](<https://github.com/Mikaelemmmm/dtmbarrier-go-zero>)<https://github.com/Mikaelemmmm/dtmbarrier-go-zero>)
