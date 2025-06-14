# WebShield
A DNS based web filtering service that allows users to block websites and apps by categories. It can block services permanently or allow temporary access as per user configuration. The objective behind this project is to allow users to reclaim their valuable time and focus on things that matter.

## Overview
![](./overview.png)

## How it works?
WebShield provides user with the 2 endpoints to choose from, these are:

1. DNS over TLS (for Android)
2. DNS over HTTPS (for browsers and major operating systems)

User needs to configure any one of the endpoints on their devices. Once configured all the DNS requests are routed via WebShield server where it's validated as per rules configured by the user which effectively enables blocking on the configured devices.

Following sequence diagram shows how DNS requests are processed on WebShield's server

```mermaid
sequenceDiagram
    participant User
    participant WebShield
    participant UpstreamDNS as Upstream DNS Server (8.8.8.8/1.1.1.1)

    User->>WebShield: 1. DNS Request
    Note over WebShield: 2. Check domain against blocking rules
    alt Domain is blocked
        WebShield->>User: 3. NXDOMAIN Response
    else Domain is allowed
        WebShield->>UpstreamDNS: 4a. Forward DNS Request
        UpstreamDNS->>WebShield: 4b. DNS Response
        WebShield->>User: 4c. Forward DNS Response to User
    end
```
Currently DNS caching is not implemented as it increases complexity significantly and causes no significant improvement in perfomance for single user usecase.

### Installation

1. Clone the repository using `git clone git@github.com:quaintdev/webshield.git`.
1. Create `blocklists` directory. The `blocklists` are available in [webshield-blocklists](https://github.com/quaintdev/webshield-blocklists) repository.
1. If you want DNS over TLS support you will have to provide TLS certs path via `config.json`.
1. Configure environment variables within `start.sh` as per your requirements. If you want DNS over TLS support then you will have to use `sudo` to run the script otherwise it's not required.
1. You can use Caddy or any other reverse proxy in front of this server for DNS over HTTPS support.

```
├── blocklists
│   ├── adult.txt
│   ├── ai.txt
│   ├── dating.txt
│   ├── entertainment.txt
│   ├── gambling.txt
│   ├── malware.txt
│   ├── news.txt
│   ├── shopping.txt
│   ├── social_media.txt
│   ├── sports.txt
│   └── streaming.txt
├── config.json
├── start.sh
├── static
   ├── guide.html
   ├── home.html
   └── styles
       └── home.css
```
### Screenshot of Webshield Panel

![WebShield Overview](./webshield.png)
