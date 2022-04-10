### dev port



#### service port

| service name | api service port(1xxx) | rpc service port(2xxx) | other service port(3xxx) |
| ------------ | ---------------------- | ---------------------- | ------------------------ |
| order        | 1001                   | 2001                   | mq-3001                  |
| payment      | 1002                   | 2002                   |                          |
| travel       | 1003                   | 2003                   |                          |
| usercenter   | 1004                   | 2004                   |                          |
| mqueue       | -                      | -                      | job-3002、schedule-3003  |



#### Prometheus Port

⚠️Online containers are separate, so online all set to the same port on it, local because in a container development, to prevent port conflicts

| service name     | prometheus port |
| ---------------- | --------------- |
| order-api        | 4001            |
| order-rpc        | 4002            |
| order-mq         | 4003            |
| payment-api      | 4004            |
| payment-rpc      | 4005            |
| travel-api       | 4006            |
| travel-rpc       | 4007            |
| usercenter-api   | 4008            |
| usercenter-rpc   | 4009            |
| mqueue-job       | 4010            |
| mqueue-scheduler | 4011            |

