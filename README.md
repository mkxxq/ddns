## overview
A ddns tool that can monitor the local export IP changes and change the dns resolution, supports aws(route53) and alidns.

## dependency

go

## install


## usage

### docker-compose
```
version: "3"
services:
  ddns:
    entrypoint: ./ddns -t aws -d cloud.mkxxq.top
    image: xiaoqiang321/ddns
    restart: always
    environment:
      AWS_ACCESS_KEY_ID: {your aws key}
      AWS_SECRET_ACCESS_KEY:  {your aws secret}
```



