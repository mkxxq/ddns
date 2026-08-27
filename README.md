## overview
A ddns tool that can monitor the local export IP changes and change the dns resolution, supports aws(route53) and cloudflare.

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
    image: ddn-amd64
    restart: always
    environment:
      AWS_ACCESS_KEY_ID: {your aws key}
      AWS_SECRET_ACCESS_KEY:  {your aws secret}
```

### docker-compose with routeros
```
version: "3"
services:
  ddns:
    entrypoint: ./ddns -t aws -d cloud.mkxxq.top
    image: ddns-amd64
    restart: always
    environment:
      AWS_ACCESS_KEY_ID: {your aws key}
      AWS_SECRET_ACCESS_KEY: {your aws secret}
      DDNS_IP_PROVIDER: routeros
      ROUTEROS_ADDR: 192.168.88.1:8728
      ROUTEROS_USER: admin
      ROUTEROS_PASS: {your routeros password}
      ROUTEROS_INTERFACE: pppoe-out1
```

### docker-compose with cloudflare
```
version: "3"
services:
  ddns:
    entrypoint: ./ddns -t cloudflare -d cloud.mkxxq.top
    image: ddns-amd64
    restart: always
    environment:
      CLOUDFLARE_API_TOKEN: {your cloudflare api token}
      CLOUDFLARE_ZONE_ID: {your cloudflare zone id}
```

### cloudflare
```
./ddns -t cloudflare -d cloud.mkxxq.top
```

env:

```
CLOUDFLARE_API_TOKEN=your-cloudflare-api-token
CLOUDFLARE_ZONE_ID=your-cloudflare-zone-id
```

`CLOUDFLARE_ZONE_ID` is optional. When it is not set, the provider will look up the active zone from the domain.

### routeros
when your export ip is on RouterOS, you can use RouterOS api provider.

```
./ddns -t aws -d cloud.mkxxq.top -ip-provider routeros
```

env:

```
DDNS_IP_PROVIDER=routeros
ROUTEROS_ADDR=192.168.88.1:8728
ROUTEROS_USER=admin
ROUTEROS_PASS=your-password
ROUTEROS_INTERFACE=pppoe-out1
```
