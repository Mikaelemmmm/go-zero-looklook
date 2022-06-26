<h1>Table of Contents</h1>

- [X. Error handling](#x-error-handling)
  - [1. Overview](#1-overview)
  - [2. rpc error handling](#2-rpc-error-handling)
  - [3. api error](#3-api-error)
  - [4. the end](#4-the-end)

# X. Error handling

This project address : <https://github.com/Mikaelemmmm/go-zero-looklook>

## 1. Overview

We in the usual development time, the program in the error, hope that the error log can quickly locate the problem (then the parameters passed in, including the stack information must be printed to the log), but at the same time want to return to the front-end users more friendly, can understand the error tips, that these two points if only through a fmt. Error information is certainly impossible to do, unless the front-end error hints in the return of the place at the same time in the log, so that the log is flying, the code is difficult to see, not to mention that the log will also be very difficult to see.

Then we think about it, if there is a unified place to record logs, while in the business code only need a return err will be returned to the front-end error message, logging information to believe that separate tips and records, if you follow this idea, it is simply not too cool, yes go-zero-looklook is so handled, we look at the next.

## 2. rpc error handling

Under normal circumstances, go-zero's rpc service is based on grpc, the default error returned is grpc status.Error can't give us a custom error merge, and is not suitable for our custom error, its error code, error type are defined dead in the grpc package, ok, if we can use custom error return in the rpc, and then in the interceptor unified return when Error, then our rpc's err and api's err can be unified to manage our own errors?

Let's look at what is inside the code of grpc's status.Error

```go
package codes // import "google.golang.org/grpc/codes"

import (
 "fmt"
 "strconv"
)

// A Code is an unsigned 32-bit error code as defined in the gRPC spec.
type Code uint32
.......
```

The error code corresponding to grpc's err is actually a uint32, so we define our own error with uint32 and then convert it to grpc's err when the global interceptor of rpc returns, and that's it

So we define our own global error code in app/common/xerr

errCode.go

```go
package xerr

// Successful return
const OK uint32 = 200


/*(The first 3 digits represent the business, the last 3 digits represent the specific function)**/

//global error code
const SERVER_COMMON_ERROR uint32 = 100001
const REUQEST_PARAM_ERROR uint32 = 100002
const TOKEN_EXPIRE_ERROR uint32 = 100003
const TOKEN_GENERATE_ERROR uint32 = 100004
const DB_ERROR uint32 = 100005

//User Module
```

errMsg.go

```go
package xerr

var message map[uint32]string

func init() {
   message = make(map[uint32]string)
   message[OK] = "SUCCESS"
   message[SERVER_COMMON_ERROR] = "The server is deserted, try again later"
   message[REUQEST_PARAM_ERROR] = "Parameter error"
   message[TOKEN_EXPIRE_ERROR] = "token is invalid, please log in again"
   message[TOKEN_GENERATE_ERROR] = "Failed to generate a token"
   message[DB_ERROR] = "The database is busy, please try again later"
}

func MapErrMsg(errcode uint32) string {
   if msg, ok := message[errcode]; ok {
      return msg
   } else {
      return "The server is deserted, try again later"
   }
}

func IsCodeErr(errcode uint32) bool {
   if _, ok := message[errcode]; ok {
      return true
   } else {
      return false
   }
}
```

errors.go

```go
package xerr

import (
   "fmt"
)

/**
Common Common Fixed Errors
*/

type CodeError struct {
   errCode uint32
   errMsg  string
}

// Error code returned to the front-end
func (e *CodeError) GetErrCode() uint32 {
   return e.errCode
}

// Return error messages to the front-end display
func (e *CodeError) GetErrMsg() string {
   return e.errMsg
}

func (e *CodeError) Error() string {
   return fmt.Sprintf("ErrCode:%d，ErrMsg:%s", e.errCode, e.errMsg)
}

func NewErrCodeMsg(errCode uint32, errMsg string) *CodeError {
   return &CodeError{errCode: errCode, errMsg: errMsg}
}
func NewErrCode(errCode uint32) *CodeError {
   return &CodeError{errCode: errCode, errMsg: MapErrMsg(errCode)}
}

func NewErrMsg(errMsg string) *CodeError {
   return &CodeError{errCode: SERVER_COMMON_ERROR, errMsg: errMsg}
}
```

For example, our rpc code at the time of user registration

```go
package logic

import (
 "context"

 "looklook/app/identity/cmd/rpc/identity"
 "looklook/app/usercenter/cmd/rpc/internal/svc"
 "looklook/app/usercenter/cmd/rpc/usercenter"
 "looklook/app/usercenter/model"
 "looklook/common/xerr"

 "github.com/pkg/errors"
 "github.com/tal-tech/go-zero/core/logx"
 "github.com/tal-tech/go-zero/core/stores/sqlx"
)

var ErrUserAlreadyRegisterError = xerr.NewErrMsg("This user has been registered")

type RegisterLogic struct {
 ctx    context.Context
 svcCtx *svc.ServiceContext
 logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
 return &RegisterLogic{
  ctx:    ctx,
  svcCtx: svcCtx,
  Logger: logx.WithContext(ctx),
 }
}

func (l *RegisterLogic) Register(in *usercenter.RegisterReq) (*usercenter.RegisterResp, error) {

 user, err := l.svcCtx.UserModel.FindOneByMobile(in.Mobile)
 if err != nil && err != model.ErrNotFound {
  return nil, errors.Wrapf(xerr.ErrDBError, "mobile:%s,err:%v", in.Mobile, err)
 }

 if user != nil {
  return nil, errors.Wrapf(ErrUserAlreadyRegisterError, "User already exists mobile:%s,err:%v", in.Mobile, err)
 }

 var userId int64

 if err := l.svcCtx.UserModel.Trans(func(session sqlx.Session) error {

  user := new(model.User)
  user.Mobile = in.Mobile
  user.Nickname = in.Nickname
  insertResult, err := l.svcCtx.UserModel.Insert(session, user)
  if err != nil {
   return errors.Wrapf(xerr.ErrDBError, "err:%v,user:%+v", err, user)
  }
  lastId, err := insertResult.LastInsertId()
  if err != nil {
   return errors.Wrapf(xerr.ErrDBError, "insertResult.LastInsertId err:%v,user:%+v", err, user)
  }
  userId = lastId

  userAuth := new(model.UserAuth)
  userAuth.UserId = lastId
  userAuth.AuthKey = in.AuthKey
  userAuth.AuthType = in.AuthType
  if _, err := l.svcCtx.UserAuthModel.Insert(session, userAuth); err != nil {
   return errors.Wrapf(xerr.ErrDBError, "err:%v,userAuth:%v", err, userAuth)
  }
  return nil
 }); err != nil {
  return nil, err
 }

 //2.gen token.
 resp, err := l.svcCtx.IdentityRpc.GenerateToken(l.ctx, &identity.GenerateTokenReq{
  UserId: userId,
 })
 if err != nil {
  return nil, errors.Wrapf(ErrGenerateTokenError, "IdentityRpc.GenerateToken userId : %d , err:%+v", userId, err)
 }

 return &usercenter.RegisterResp{
  AccessToken:  resp.AccessToken,
  AccessExpire: resp.AccessExpire,
  RefreshAfter: resp.RefreshAfter,
 }, nil
}
```

```go
errors.Wrapf(ErrUserAlreadyRegisterError, "User already exists mobile:%s,err:%v", in.Mobile, err)
```

Wrapf (if you don't understand here, look up Wrap, Wrapf, etc. under go's errors package)

The first parameter, ErrUserAlreadyRegisterError, is defined above and is the use of xerr.NewErrMsg("The user has been registered"), which returns a friendly hint to the front end, remember that here we use the methods under the xerr package

The second parameter, which is recorded in the server log, can be written in detail it does not matter only recorded in the server will not be returned to the front-end

Then let's see why the first parameter can be returned to the front-end, the second parameter is the logging

⚠️  [Note] we in the rpc startup file main method, add the grpc global interceptor, this is very important, if not add this no way to achieve

```go
package main

......

func main() {

 ........

  //rpc log,global interceptor for grpc
 s.AddUnaryInterceptors(rpcserver.LoggerInterceptor)

 .......
}
```

Let's look at the specific implementation of rpcserver.LoggerInterceptor

```go
import(
  ...
   "github.com/pkg/errors"
)

func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

   resp, err = handler(ctx, req)
   if err != nil {
      causeErr := errors.Cause(err)                // err type
      if e, ok := causeErr.(*xerr.CodeError); ok { //Custom Error Types
         logx.WithContext(ctx).Errorf("【RPC-SRV-ERR】 %+v", err)

          //to grpc err
         err = status.Error(codes.Code(e.GetErrCode()), e.GetErrMsg())
      } else {
         logx.WithContext(ctx).Errorf("【RPC-SRV-ERR】 %+v", err)
      }

   }

   return resp, err
}
```

When there is a request into the rpc service, first enter the interceptor and then is the execution of the handler method, if you want to deal with certain things before entering you can write before the handler method, that we want to deal with is to return the results if there is an error, so we use github.com/pkg/errors below the handler this package, this package is often used in the go error is not the official errors package, but the design is very good, go official Wrap, Wrapf, etc. is borrowed from the idea of this package.

Because we grpc internal business in the return of errors when

1) If it is our own business error, we will unify the error generated with xerr, so that we can get the error information we defined, because the front of our own error is also used uint32, so here unified into grpc error err = status.Error(codes.Code(e. GetErrCode()), e.GetErrMsg()), that here to get, e.GetErrCode() is our definition of code, e.GetErrMsg() is our previous definition of the return of the error of the second parameter

   (2) but there is another situation is the bottom of the rpc service exception thrown out of the error, itself is grpc error, then this kind of we directly record the exception on the good

## 3. api error

When our api calls rpc's Register in the logic, rpc returns the error message in step 2 above Code is as follows

```go
......
func (l *RegisterLogic) Register(req types.RegisterReq) (*types.RegisterResp, error) {
 registerResp, err := l.svcCtx.UsercenterRpc.Register(l.ctx, &usercenter.RegisterReq{
  Mobile:   req.Mobile,
  Nickname: req.Nickname,
  AuthKey:  req.Mobile,
  AuthType: model.UserAuthTypeSystem,
 })
 if err != nil {
  return nil, errors.Wrapf(err, "req: %+v", req)
 }

 var resp types.RegisterResp
 _ = copier.Copy(&resp, registerResp)

 return &resp, nil
}
```

Wrapf , which means that all the errors returned by our business will be applied to the standard package of errors, but the internal parameters will use the errors defined by our xerr

There are 2 points to note here

1) api service wants to return the rpc to the front-end friendly error message, we want to return directly to the front-end without any processing (for example, the rpc has returned "user already exists", api does not want to do any processing, you want to return this error message directly to the front-end)

Wrapf the first parameter, but the second parameter is best to record the detailed logs you need to facilitate the follow-up in api China it view

(2) api service regardless of what error information is returned by the rpc, I would like to redefine the error information returned to the front-end (for example, the rpc has returned "user already exists", api want to call the rpc as long as there is an error I will return to the front-end "User registration failed")

For this case, write the following (of course you can put xerr.NewErrMsg("user registration failed") on top of the code and use a variable, it's okay to put the variable here)

```go
func (l *RegisterLogic) Register(req types.RegisterReq) (*types.RegisterResp, error) {
 .......
 if err != nil {
  return nil, errors.Wrapf(xerr.NewErrMsg("User registration failed"), "req: %+v,rpc err:%+v", req,err)
 }
 .....
}

```

Next we see how the final return to the front end is handled, we then look at app/usercenter/cmd/api/internal/handler/user/registerHandler.go

```go
func RegisterHandler(ctx *svc.ServiceContext) http.HandlerFunc {
 return func(w http.ResponseWriter, r *http.Request) {
  var req types.RegisterReq
  if err := httpx.Parse(r, &req); err != nil {
   result.ParamErrorResult(r,w,err)
   return
  }

  l := user.NewRegisterLogic(r.Context(), ctx)
  resp, err := l.Register(req)
  result.HttpResult(r, w, resp, err)
 }
}
```

Here you can see, go-zero-looklook generated handler code There are two places with the default official goctl generated code is not the same, is in the processing of error handling, here replaced with our own error handling, in common/result/httpResult.go

Note】Some people will say, every time you use goctl to manually change, that is not to trouble dead, here we use go-zero to provide us with the template template function (do not know this will have to go to the official documentation to learn a little), modify the handler to generate the template can be, the entire project template file placed under deploy/goctl. Here hanlder modify the template in deploy/goctl/1.2.3-cli/api/handler.tpl

ParamErrorResult is very simple, dedicated to handling parameter errors

```
//http Parameter error returned
func ParamErrorResult(r *http.Request, w http.ResponseWriter, err error) {
   errMsg := fmt.Sprintf("%s ,%s", xerr.MapErrMsg(xerr.REUQEST_PARAM_ERROR), err.Error())
   httpx.WriteJson(w, http.StatusBadRequest, Error(xerr.REUQEST_PARAM_ERROR, errMsg))
}
```

We will mainly look at the HttpResult, the error handling returned by the business

```go

//http return
func HttpResult(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {

 if err == nil {
  // Successful return
  r := Success(resp)
  httpx.WriteJson(w, http.StatusOK, r)
 } else {
   //Error return
  errcode := xerr.SERVER_COMMON_ERROR
  errmsg := "The server is deserted, try again later"

  causeErr := errors.Cause(err)                // err type
  if e, ok := causeErr.(*xerr.CodeError); ok { //Custom Error Types
   // Custom CodeError
   errcode = e.GetErrCode()
   errmsg = e.GetErrMsg()
  } else {
   if gstatus, ok := status.FromError(causeErr); ok { // grpc err error
    grpcCode := uint32(gstatus.Code())
    if xerr.IsCodeErr(grpcCode) { // Distinguish between custom errors and system underlying, db errors, etc., underlying, db errors cannot be returned to the front-end
     errcode = grpcCode
     errmsg = gstatus.Message()
    }
   }
  }

  logx.WithContext(r.Context()).Errorf("【API-ERR】 : %+v ", err)

  httpx.WriteJson(w, http.StatusBadRequest, Error(errcode, errmsg))
 }
}
```

err : The error to be logged

errcode : error code to return to the front-end

errmsg : friendly error message returned to the front-end

If we encounter an error, we also use github.com/pkg/errors to determine whether it is our own error (the error defined in the api directly uses our own xerr) or a grpc error (thrown by the rpc business), and if it is a grpc error, we convert it to our own error code via uint32. own error code, according to the error code and then go to our own definition of error information to find the definition of error information returned to the front end, if the api error directly back to the front end of our own definition of error information, can not find that the default error returned "the server deserted" ,

## 4. the end

Here the error handling has been clearly described in the message, the next we have to look at printing the server-side error log, we should collect how to view, it involves the log collection system.
