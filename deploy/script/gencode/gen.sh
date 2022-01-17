# 生成api ， 进入"服务/cmd"目录下，执行下面命令
# goctl api go -api ./api/desc/*.api -dir ./api -style=goZero

# 生成rpc ， 进入"服务/cmd"目录下，执行下面命令
# goctl rpc proto -src rpc/pb/*.proto -dir ./rpc -style=goZero
# 去除proto中的json的omitempty
# sed -i 's/,omitempty//g'  ./rpc/pb/*.pb.go

# docker + air hot deployment basic
# https://hub.docker.com/r/cosmtrek/air


# 创建kafka的topic
# kafka-topics.sh --create --zookeeper zookeeper:2181 --replication-factor 1 -partitions 1 --topic {topic}
# 查看消费者组情况
# kafka-consumer-groups.sh --bootstrap-server kafka:9092 --describe --group {group}
# 命令行消费
# ./kafka-console-consumer.sh  --bootstrap-server kafka:9092  --topic looklook-log   --from-beginning
# 命令生产
# ./kafka-console-producer.sh --bootstrap-server kafka:9092 --topic second
