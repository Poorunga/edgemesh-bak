[![Version](https://img.shields.io/badge/EdgeMesh-0.1-orange)](https://hub.docker.com/r/poorunga/edgemesh)   [![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)



English | [简体中文](./README_ZH.md)



| ![notification](/images/bell-outline-badge.svg) What is NEW! |
| ------------------------------------------------------------ |
| April 20th, 2021. EdgeMesh v0.1.0 is **RELEASED**! Please check the [README](./README.md) for details. |



## Introduction

EdgeMesh is a type of service mesh, which is closely related to KubeEdge, and is applicable to the edge scenarios.

In recent years, as cloud native and microservice architectures have become more and more popular, the functions of edge nodes have become more and more complete. In this scenario, users can deploy their own applications on edge nodes, and these edge nodes must be accessed by external users through services. For this purpose, EdgeMesh provides users (users running applications in KubeEdge) with external access services running on different nodes, allowing them to access other edge nodes. In addition, EdgeMesh also provides capabilities of load balancing and traffic management for users.



## Advantage

EdgeMesh satisfies the new requirements in edge scenarios (e.g., limited edge resources, unstable edge cloud network, etc.), that is, high availability, high reliability, and extreme lightweight:

- **High availability**
  - Open up the network between edge nodes by using the edge cloud channel in KubeEdge
  - Divide the communication between edge nodes into intra-LAN and cross-LAN
    - Intra-LAN communication: direct access
    - Cross-LAN communication: forwarding through the cloud
- **High reliability (offline scenario)**
  - Both control plane and data plane traffic are delivered through the edge cloud channel
  - EdgeMesh internally implements a lightweight DNS server, thus no longer accessing the cloud DNS
- **Extreme lightweight**
  - Each node has one and only one EdgeMesh, which saves edge resources

#### User value

- For edge devices with limited resources, EdgeMesh provides a lightweight and highly integrated software with service discovery
- In the scene of Field Edge, compared to the mechanism of coredns + kube-proxy + cni service discovery , users only need to simply deploy an EdgeMesh to finish their goals



## Architecture

![](/images/em-principle.png)

To ensure the capability of service discovery in some edge devices with low-version kernels or low-version iptables, EdgeMesh adopts the userspace mode in its implementation. In addition, it also comes with a lightweight DNS server.

- Through the capability of list-watch on the edge of KubeEdge, EdgeMesh monitors the addition, deletion and modification of metadata (e.g., services and endpoints), and then creates iptables rules based on services and endpoints
- EdgeMesh uses domain names to access services, since fakeIP does NOT be exposed to users. FakeIP is similar to clusterIP, and its CIDR is between the network segment of 9.251.0.0/16 in each node (which will be unified into K8s service network in the future)
- When client's requests accessing a service reach a node with EdgeMesh, it will enter the kernel's iptables at first
- The iptables rules previously configured by EdgeMesh will redirect requests, and forward them all to the port 40001 which is occupied by the EdgeMesh process (data packets from kernel mode to user mode)
- After requests enter the EdgeMesh process, the EdgeMesh process completes the selection of backend Pods (load balancing occurs here), and then sends requests to the host where the Pod is located




## Functionality

|       Feature        |     Sub-Feature     | Edgemesh 0.1 |
| :------------------: | :-----------------: | :----------: |
|  Service Discovery   |                     |     `✓`      |
|  Traffic Governance  |        HTTP         |     `✓`      |
|                      |         TCP         |     `✓`      |
|                      |      Websocket      |     `✓`      |
|                      |        HTTPS        |     `✓`      |
|     Load Balance     |       Random        |     `✓`      |
|                      |     Round Robin     |     `✓`      |
|                      | Session Persistence |     `✓`      |
|   External Access    |                     |     `✓`      |
| Multi-NIC Monitoring |                     |     `✓`      |

**Noting:**

- `✓` Features supported by the EdgeMesh version 
- `+` Features not available in the EdgeMesh version, but will be supported in subsequent versions
- `-` Features not available in the EdgeMesh version, or deprecated features



## Operation Guidance

#### Deployment

​	At the edge node, close EdgeMesh, open metaserver, and restart edgecore

```shell
$ vim /etc/kubeedge/config/edgecore.yaml
modules:
  ..
  edgeMesh:
    enable: false
  metaManager:
    metaServer:
      enable: true
..
```

```shell
$ systemctl restart edgecore
```

​	On the cloud, open the dynamic controller module, and restart cloudcore

```shell
$ vim /etc/kubeedge/config/cloudcore.yaml
modules:
  ..
  dynamicController:
    enable: true
..
```

​	At the edge node, check if listwatch works

```shell
$ curl 127.0.0.1:10550/api/v1/services
{"apiVersion":"v1","items":[{"apiVersion":"v1","kind":"Service","metadata":{"creationTimestamp":"2021-04-14T06:30:05Z","labels":{"component":"apiserver","provider":"kubernetes"},"name":"kubernetes","namespace":"default","resourceVersion":"147","selfLink":"default/services/kubernetes","uid":"55eeebea-08cf-4d1a-8b04-e85f8ae112a9"},"spec":{"clusterIP":"10.96.0.1","ports":[{"name":"https","port":443,"protocol":"TCP","targetPort":6443}],"sessionAffinity":"None","type":"ClusterIP"},"status":{"loadBalancer":{}}},{"apiVersion":"v1","kind":"Service","metadata":{"annotations":{"prometheus.io/port":"9153","prometheus.io/scrape":"true"},"creationTimestamp":"2021-04-14T06:30:07Z","labels":{"k8s-app":"kube-dns","kubernetes.io/cluster-service":"true","kubernetes.io/name":"KubeDNS"},"name":"kube-dns","namespace":"kube-system","resourceVersion":"203","selfLink":"kube-system/services/kube-dns","uid":"c221ac20-cbfa-406b-812a-c44b9d82d6dc"},"spec":{"clusterIP":"10.96.0.10","ports":[{"name":"dns","port":53,"protocol":"UDP","targetPort":53},{"name":"dns-tcp","port":53,"protocol":"TCP","targetPort":53},{"name":"metrics","port":9153,"protocol":"TCP","targetPort":9153}],"selector":{"k8s-app":"kube-dns"},"sessionAffinity":"None","type":"ClusterIP"},"status":{"loadBalancer":{}}}],"kind":"ServiceList","metadata":{"resourceVersion":"377360","selfLink":"/api/v1/services"}}
```

​	Deploy configmap, and create Istio's crds

```shell
$ kubectl apply -f 03-configmap.yaml
configmap/edgemesh-cfg created
$ kubectl apply -f istio-crds-simple.yaml
customresourcedefinition.apiextensions.k8s.io/virtualservices.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/destinationrules.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/serviceentries.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/gateways.networking.istio.io created
```

​	Use daemonset to deploy EdgeMesh

```shell
$ kubectl apply -f 05-daemonset.yaml
daemonset.apps/edgemesh created
```



#### Test Case

​	**HTTP**

​	At the edge node, deploy a HTTP container application, and relevant service

```shell
$ kubectl apply -f hostname.yaml
```

​	Go to that edge node, use ‘curl’ to access the service, and print out the hostname of the container

```shell
$ curl hostname-lb-svc.edgemesh-test:12345
```



​	**TCP**

​	At the edge node 1, deploy a TCP container application, and relevant service	

```shell
$ kubectl apply -f tcp-echo-service.yaml
```

​	At the edge node 1, use ‘telnet’ to access the service		

```shell
$ telnet tcp-echo-service.edgemesh-test 2701
```



​	**Websocket**

​	At the edge node, deploy a websocket container application, and relevant service	

```shell
$ kubectl apply -f websocket-pod-svc.yaml
```

​	Enter the container, and use ./client to access the service

```shell
$ docker exec -it 2a6ae1a490ae bash
$ ./client --addr ws-svc.edgemesh-test:12348
```



​	**Load Balance**

​	Use the 'loadBalancer' in 'DestinationRule' to select LB modes	

```shell
$ vim edgemesh-gateway-dr.yaml
spec
..
  trafficPolicy:
    loadBalancer:
      simple: RANDOM
..    
```



## EdgeMesh Ingress Gateway

EdgeMesh ingress gateway provides a ability to access services in external edge nodes.

![image-20210414152916134](/images/em-ig.png)

#### Deployment

​	Create Istio's crds

```shell
$ kubectl apply -f istio-crds-simple.yaml
customresourcedefinition.apiextensions.k8s.io/virtualservices.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/destinationrules.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/serviceentries.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/gateways.networking.istio.io created
```

​	Configure configmap, and deploy edgemesh-gateway

```shell
$ kubectl apply -f 03-configmap.yaml 
configmap/edgemesh-gateway-cfg created
$ kubectl apply -f 04-deployment.yaml 
deployment.apps/edgemesh-gateway created
```

​	Create 'gateway' and 'Virtual Service'

```shell
$ kubectl apply -f edgemesh-gateway-gw-vsvc.yaml
gateway.networking.istio.io/edgemesh-gateway created
virtualservice.networking.istio.io/edgemesh-gateway-vsvc created
```

​	Check if the edgemesh-gateway is successfully deployed

```shell
$ kubectl get gw -n edgemesh-test
NAME               AGE
edgemesh-gateway   3m30s
```

​	Finally, use the IP and the port exposed by the Virtual Service to access

```shell
$ curl 192.168.0.211:23333
```



## Contact

If you need support, start with the 'Operation Guidance', and then follow the process that we've outlined

If you have further questions, feel free to reach out to us in the following ways:

​	[Bilibili KubeEdge](https://space.bilibili.com/448816706?from=search&seid=10057261257661405253)
