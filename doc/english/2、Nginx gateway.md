### 二、Using nginx as a gateway

Address of this project :  https://github.com/Mikaelemmmm/go-zero-looklook



#### 1、go-zero gateway concept

go-zero architecture to the big say there are mainly 2 parts, one is api, one is rpc. api is mainly http external access, rpc is mainly internal business interaction using protobuf + grpc, when our project volume is not large, we can use api to do a monolithic project, and after the subsequent volume up, you can split to rpc to do microservices, from a single body to microservices is very easy, much like the java springboot to springcloud, very convenient.

api is understood by many students as a gateway, in a practical sense when your project is using go-zero to do microservices, you api as a gateway is not a big problem, but this leads to the problem that an api corresponds to the back of multiple rpc, api acts as a gateway, so if I update the subsequent business code, update any business to change the For example, if I just change a small and insignificant service, I have to reconstruct the whole api, which is not very reasonable and very inefficient and inconvenient. So, we just treat the api as an aggregated service, can be split into multiple api, such as user services have user services rpc and api, order services, order services rpc and api, so that when I modify the user service, I only need to update the user rpc and api, all the api is only used to aggregate the back-end rpc business. Then some students will say, I can't resolve a domain name for each service corresponding to your api, of course not, this time there should be a gateway in front of the api, this gateway is the real sense of the gateway, such as we often say nginx, kong, apisix, many microservices have built-in gateway, such as springcloud provides springcloud-gateway, go-zero does not provide, and actually do not need to write a separate gateway, the gateway on the market has been enough, go-zero official in the dawn of blackboard nginx enough to use, of course, if you are more familiar with kong, apisix can be replaced, essentially nothing different, just a unified traffic Entrance, unified authentication, etc.



#### 2、nginx网关

[Note]: When looking here, it is recommended to look at the business architecture diagram in the previous section first

![nignx-svc](../chinese/images/2/nginx-gateway.jpg)

The actual project also uses nginx as a gateway, using the auth_request module of nginx as a unified authentication, business internal authentication is not done (designed to the assets of the best business internal authentication, the main extra layer of security), nignx gateway configuration in the project's deploy/nginx/conf.d/looklook- gateway.conf

```conf
server{
    listen 8081;
    access_log /var/log/nginx/looklook.com_access.log;
    error_log /var/log/nginx/looklook.com_error.log;

    location /auth {
	    internal;
        proxy_set_header X-Original-URI $request_uri;
	    proxy_pass_request_body off;
	    proxy_set_header Content-Length "";
	    proxy_pass http://looklook:8001/identity/v1/verify/token;
    }

    location ~ /usercenter/ {
       auth_request /auth;
       auth_request_set $user $upstream_http_x_user;
       proxy_set_header x-user $user;

       proxy_set_header Host $http_host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header REMOTE-HOST $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_pass http://looklook:8002;
   }

   location ~ /travel/ {
       auth_request /auth;
       auth_request_set $user $upstream_http_x_user;
       proxy_set_header x-user $user;

       proxy_set_header Host $http_host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header REMOTE-HOST $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_pass http://looklook:8003;
   }


    location ~ /order/ {
       auth_request /auth;
       auth_request_set $user $upstream_http_x_user;
       proxy_set_header x-user $user;

       proxy_set_header Host $http_host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header REMOTE-HOST $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_pass http://looklook:8004;
   }

    location ~ /payment/ {
       auth_request /auth;
       auth_request_set $user $upstream_http_x_user;
       proxy_set_header x-user $user;

       proxy_set_header Host $http_host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header REMOTE-HOST $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_pass http://looklook:8005;
   }

}
```

Container internal nginx port is 8081, use docker to expose out 8888 mapping port 8081, so that the external through 8888 to access the gateway, use location to match each service, of course, there will be people say, did not add an api service are to nignx configuration is too much trouble, you can also use confd unified configuration, self Baidu.



#### 3、Examples

When we access the user service, http://127.0.0.1:8888/usercenter/v1/user/detail , we access the external port 8888, which is then mapped to the nginx internal lookup gateway 8081, and then the location matches to /usercenter/ , where the module starts with a line auth_request /auth, so nginx will not request http://usercenter-api:8002 directly, but will first jump to the location /auth module, which will access http://identity-api:8001/ identity/v1/verify/token; ,identity-api is also our internal service, which is an authentication service written by ourselves, and actually uses go-zero's jwt

into the identity-api only do 2 things (see the identity-api code in the looklook project)

1, determine whether the currently accessed route (usercenter/v1/user/detail) needs to be logged in

Whether the route here needs to be logged in, can be configured in identity-api, the code has been implemented

​	![image-20220117162935341](../chinese/images/2/image-20220117162935341.png)



2、Parsing the passed token into the header

- If the currently accessed route requires login.

  - token parsing failure: it returns an http401 error code to the front-end.

  - The token is parsed successfully: the parsed userId will be put into the x-user of the header and returned to the auth module, which will pass the header to the corresponding service (usercenter), so that we can get the login user's id directly in usercenter

- If the currently accessed route does not require login：

  - The token is passed in the front-end header
    - If the token checksum fails: return http401.
    - If the token verification is successful: the parsed userId will be put into the x-user of the header and returned to the auth module, which will pass the header to the corresponding service (usercenter), so that we can get the login user's id directly in usercenter

  - No token passed in the front-end header: userid will pass 0 to the back-end service





#### 4、Summary

So that we can unify the entrance, unified authentication, but also unified collection of logs reported, used as error analysis, or access to the user's behavior analysis. Because our daily use of nginx more, and more familiar, if you students are more familiar with kong, apisix, in understanding the above go-zero use of the concept of gateway can be directly replaced is also the same.





