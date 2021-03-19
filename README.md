# goreverse
Automatically detect running docker containers and redirect http traffic to them depending on the url.

I have a pi that runs pihole and other non-public services. I want to set nice names for these services in the pihole dns and just start docker containers. I guess [jwilder/nginx-proxy](https://hub.docker.com/r/jwilder/nginx-proxy) does something similar, although I haven't really looked at it.

## idea
* search all running containers for the tag `goreverse` and parse the value as a url
* get the public port that is mapped to private port 80
* proxy requests sent to the public port 80 of the machine to the private port of the matching container
* profit

## notes
* the resolved url is set in the [host](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Host) header
* [nginx](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/)
* [reverse proxy](https://golang.org/pkg/net/http/httputil/#ReverseProxy) already sets `X-Forwarded-For`
* [Ealenn/Echo-Server](https://github.com/Ealenn/Echo-Server) is nice for manual testing: `docker run --rm -p 80 -l goreverse=http://kek.com -d ealen/echo-server`

## todo
* accept https and terminate ssl
* listen on [docker event](https://pkg.go.dev/github.com/docker/docker@v20.10.5+incompatible/client#Client.Events) channel and update on demand
* tests lul
