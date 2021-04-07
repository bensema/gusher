# gusher
`$ go build`
master mode:

`$ ./gusher mater -e master.env.example ` or `./gusher ma -e master.env.example -d`

slave mode:

`$ ./gusher slave -e slave.env.example` or `./gusher sl  --env-file slave.env.example -d`


流程
连接返回
pusher:connection_established
socket_id
activity_timeout

{event: "pusher:connection_established", data: "{"socket_id":"3469.6261593","activity_timeout":120}"}
data: "{"socket_id":"3469.6261593","activity_timeout":120}"
event: "pusher:connection_established"
