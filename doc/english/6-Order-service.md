<h1>Table of Contents</h1>

- [VI. Order Service](#vi-order-service)
  - [1. Order service business architecture diagram](#1-order-service-business-architecture-diagram)
  - [2. Dependencies](#2-dependencies)
  - [3. Order examples](#3-order-examples)
    - [3.1. Placing an order](#31-placing-an-order)
    - [3.2. Order List](#32-order-list)
    - [3.3. Order details](#33-order-details)
  - [4. Closing](#4-closing)

# VI. Order Service

This project address : <https://github.com/Mikaelemmmm/go-zero-looklook>

## 1. Order service business architecture diagram

<img src="../chinese/images/6/image-20220428110910672.png" alt="image-20220213133955478" style="zoom:50%;" />

## 2. Dependencies

order-api (order-api) Dependencies order-rpc (order-rpc), payment-rpc (payment-rpc), travel-rpc (B&B rpc)

order-rpc (order-rpc) depend on travel-rpc (B&B rpc)

## 3. Order examples

### 3.1. Placing an order

1, the user in the travel service to browse the B&B homestay to choose the date to place an order, call the order api interface

app/order/cmd/api/desc/order.api

```protobuf
//order module v1 interface
@server(
   prefix: order/v1
   group: homestayOrder
)
service order {

   @doc "Create Homestay Order"
   @handler createHomestayOrder
   post /homestayOrder/createHomestayOrder (CreateHomestayOrderReq) returns (CreateHomestayOrderResp)

   .....
}
```

2, order-api call order-rpc

![image-20220120130235305](../chinese/images/6/image-20220120130235305.png)

3. After creating an order by checking the conditions in rpc, **Asynq will be called to create a message queue for delayed order closure**

go-zero-looklook/app/order/cmd/rpc/internal/logic/createHomestayOrderLogic.go

```go
// CreateHomestayOrder.
func (l *CreateHomestayOrderLogic) CreateHomestayOrder(in *pb.CreateHomestayOrderReq) (*pb.CreateHomestayOrderResp, error) {

 .....

 //2.Delayed closing of order tasks.
 payload, err := json.Marshal(jobtype.DeferCloseHomestayOrderPayload{Sn: order.Sn})
 if err != nil {
  logx.WithContext(l.ctx).Errorf("create defer close order task json Marshal fail err :%+v , sn : %s",err,order.Sn)
 }else{
  _, err = l.svcCtx.AsynqClient.Enqueue(asynq.NewTask(jobtype.DeferCloseHomestayOrder, payload), asynq.ProcessIn(CloseOrderTimeMinutes * time.Minute))
  if err != nil {
   logx.WithContext(l.ctx).Errorf("create defer close order task insert queue fail err :%+v , sn : %s",err,order.Sn)
  }
 }

 return &pb.CreateHomestayOrderResp{
  Sn: order.Sn,
 }, nil
}


```

4. go-zero-looklook/app/mqueue/cmd/job/internal/logic/closeOrder.go has a delayed close order task that defines asynq

```go

// defer  close no pay homestayOrder  : if return err != nil , asynq will retry
func (l *CloseHomestayOrderHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {

 var p jobtype.DeferCloseHomestayOrderPayload
 if err := json.Unmarshal(t.Payload(), &p); err != nil {
  return errors.Wrapf(ErrCloseOrderFal, "closeHomestayOrderStateMqHandler payload err:%v, payLoad:%+v", err, t.Payload())
 }

 resp, err := l.svcCtx.OrderRpc.HomestayOrderDetail(ctx, &order.HomestayOrderDetailReq{
  Sn: p.Sn,
 })
 if err != nil || resp.HomestayOrder == nil {
  return errors.Wrapf(ErrCloseOrderFal, "closeHomestayOrderStateMqHandler  get order fail or order no exists err:%v, sn:%s ,HomestayOrder : %+v", err, p.Sn, resp.HomestayOrder)
 }

 if resp.HomestayOrder.TradeState == model.HomestayOrderTradeStateWaitPay {
  _, err := l.svcCtx.OrderRpc.UpdateHomestayOrderTradeState(ctx, &order.UpdateHomestayOrderTradeStateReq{
   Sn:         p.Sn,
   TradeState: model.HomestayOrderTradeStateCancel,
  })
  if err != nil {
   return errors.Wrapf(ErrCloseOrderFal, "CloseHomestayOrderHandler close order fail  err:%v, sn:%s ", err, p.Sn)
  }
 }

 return nil
}


```

So we start this mqueue-job, asynq will be loaded, define the route, when we previously added the delay queue to 20 minutes, it will automatically perform the order closure logic, if the order is not paid, here will close the order, paid ignored, so that you can close the order without using the timed task rotation training, ha ha

### 3.2. Order List

There is no logic, just check out the display, just look at it yourself

```protobuf
//Order module v1 interface
@server(
   prefix: order/v1
   group: homestayOrder
)
service order {

   @doc "User order list"
   @handler userHomestayOrderList
   post /homestayOrder/userHomestayOrderList (UserHomestayOrderListReq) returns (UserHomestayOrderListResp)

}
```

### 3.3. Order details

There is no logic, just check out the display, just look at it yourself

```protobuf
//Order module v1 interface
@server(
 prefix: order/v1
 group: homestayOrder
)
service order {

 @doc "User order details"
 @handler userHomestayOrderDetail
 post /homestayOrder/userHomestayOrderDetail (UserHomestayOrderDetailReq) returns (UserHomestayOrderDetailResp)
}
```

## 4. Closing

After placing an order, of course we have to pay for it, so let's see the next payment service
