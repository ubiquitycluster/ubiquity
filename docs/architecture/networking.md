# Networking

As an example of how the networking works (because each system is different), this is an example of how you can
use cloudflared to setup a tunnel, that goes to and from the cluster.
```mermaid
flowchart TD
  subgraph LAN
    laptop/desktop/phone <--> LoadBalancer
    subgraph k8s[Kubernetes cluster]
      Pod --> Service
      Service --> Ingress

      LoadBalancer

      cloudflared
      cloudflared <--> Ingress
    end
    LoadBalancer <--> Ingress
  end

  cloudflared -- outbound --> Cloudflare
  Internet -- inbound --> Cloudflare
```

TODO
