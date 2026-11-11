## overview
A ddns tool that can monitor the local export IP changes and change the dns resolution. Currently only Cloudflare is supported.

## dependency

go

## install


## usage

### docker-compose
```
services:
  ddns:
    entrypoint: ./ddns -t cloudflare -d cloud.mkxxq.top
    image: ddns-amd64
    restart: always
    env_file:
      - ddns.env
    environment:
      DDNS_DNS_SERVER: 1.1.1.1:53,223.5.5.5:53
```

`ddns.env` holds the credentials and the WAN IP source (`CLOUDFLARE_API_TOKEN`,
`DDNS_IP_PROVIDER`, `ROUTEROS_*`); keep it out of version control.

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

### dns servers
The domain's current record is read with an **explicit** resolver list, never with the system resolver:

```
DDNS_DNS_SERVER=1.1.1.1:53,223.5.5.5:53   # comma separated, tried in order
```

A bare IP gets `:53` appended. Both servers failing counts as one failed lookup for the round; the update
still proceeds and the record is only rewritten when the outer IP actually differs. Each query is bounded
to 5s (Go's resolver would otherwise retry for ~20s).

The loop retries after 5s once anything fails (router API, Cloudflare API), otherwise every 60s. After a
successful write it keeps the written address until the record's TTL expires, so a stale cached answer does
not trigger repeated PATCHes.

### routeros
when your export ip is on RouterOS, you can use RouterOS api provider.

```
./ddns -t cloudflare -d cloud.mkxxq.top -ip-provider routeros
```

env:

```
DDNS_IP_PROVIDER=routeros
ROUTEROS_ADDR=192.168.88.1:8728
ROUTEROS_USER=admin
ROUTEROS_PASS=your-password
ROUTEROS_INTERFACE=pppoe-out1
```
