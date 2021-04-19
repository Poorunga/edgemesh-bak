[![Version](https://img.shields.io/badge/EdgeMesh-0.1-orange)](https://hub.docker.com/r/poorunga/edgemesh)[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)



[English](./README.md) | [简体中文]



| ![notification](docs/img/bell-outline-badge.svg) What is NEW! |
| ------------------------------------------------------------ |
| 2021年四月二十日. Edgemesh v0.1.0 **发布**了! 请查看 [README](./README.md) 了解更多细节. |



## 介绍

EdgeMesh是一种服务网格，它与KubeEdge紧密结合，并适用于边缘场景。

近年来，随着云原生和微服务架构的越来越流行，边缘节点的功能也越来越完善。在这个场景下，用户能够在边缘节点上部署自己的应用，而这些边缘节点必须通过服务被外部用户所访问。出于这个目的，EdgeMesh向用户（在KubeEdge内运行应用的用户）提供运行在不同节点上的外部访问服务，从而使他们访问其它边缘节点。此外，EdgeMesh还向用户供负载均衡和流量治理的能力。



## 优势

EdgeMesh满足边缘场景下的新需求（如边缘资源有限，边云网络不稳定等），即实现了高可用性，高可靠性，和极致轻量化：

- 高可用性
  - 利用KubeEdge中的边云通道，来打通边缘节点间的网络
  - 将边缘节点间的通信分为局域网内和跨局域网
    - 局域网内的通信：直接访问
    - 跨局域网的通信：通过云端转发
- 高可靠性 （离线场景）
  - 控制面和数据面流量都通过边云通道下发
  - EdgeMesh内部实现轻量级的DNS服务器，不再访问云端DNS
- 极致轻量化
  - 每个节点有且仅有一个EdgeMesh，节省边缘资源



## 架构

![Edgemesh架构图](.\em-arch.jpg)

如图所示，Edgemesh的工作流如下：

- 通过服务器域名，客户端向服务器发起请求
- 根据iptables规则，DNS被劫持到EdgeMesh，并返回一个虚假IP

- 根据iptables规则，请求也被劫持到EdgeMesh

- EdgeMesh解析请求，获得域名，协议和请求等信息，

- EdgeMesh的负载均衡：
  - 从域名中获得服务名和命名空间
  - 在MetaManager中，查询该服务的后端Pods
  - 基于负载均衡策略，选择一个后端Pod
- EdgeMesh向服务器传输请求，等待服务器的响应，并将响应返回给客户端



## 功能性

|    功能    |  子功能   | Edgemesh 0.1 |
| :--------: | :-------: | :----------: |
|  服务发现  |           |     `✓`      |
|  流量治理  |   HTTP    |     `✓`      |
|            |    TCP    |     `✓`      |
|            | Websocket |     `✓`      |
|            |   HTTPS   |     `✓`      |
|  负载均衡  |   随机    |     `✓`      |
|            |   轮询    |     `✓`      |
|            | 会话保持  |     `✓`      |
|  外部访问  |           |     `✓`      |
| 多网卡监听 |           |     `✓`      |

注：

- `✓` EdgeMesh版本所支持的功能
- `+` EdgeMesh版本不具备的功能，但在后续版本中会支持
- `-` EdgeMesh版本不具备的功能，或已弃用的功能



## 操作指导

#### 部署

​	在边缘节点，关闭edgemesh，打开metaserver，并重启edgecore

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

​	在云端，开启dynamic controller模块，并重启cloudcore

```shell
$ vim /etc/kubeedge/config/cloudcore.yaml
modules:
  ..
  dynamicController:
    enable: true
..
```

​	在边缘节点，查看listwatch是否开启

```shell
$ curl 127.0.0.1:10550/api/v1/services
{"apiVersion":"v1","items":[{"apiVersion":"v1","kind":"Service","metadata":{"creationTimestamp":"2021-04-14T06:30:05Z","labels":{"component":"apiserver","provider":"kubernetes"},"name":"kubernetes","namespace":"default","resourceVersion":"147","selfLink":"default/services/kubernetes","uid":"55eeebea-08cf-4d1a-8b04-e85f8ae112a9"},"spec":{"clusterIP":"10.96.0.1","ports":[{"name":"https","port":443,"protocol":"TCP","targetPort":6443}],"sessionAffinity":"None","type":"ClusterIP"},"status":{"loadBalancer":{}}},{"apiVersion":"v1","kind":"Service","metadata":{"annotations":{"prometheus.io/port":"9153","prometheus.io/scrape":"true"},"creationTimestamp":"2021-04-14T06:30:07Z","labels":{"k8s-app":"kube-dns","kubernetes.io/cluster-service":"true","kubernetes.io/name":"KubeDNS"},"name":"kube-dns","namespace":"kube-system","resourceVersion":"203","selfLink":"kube-system/services/kube-dns","uid":"c221ac20-cbfa-406b-812a-c44b9d82d6dc"},"spec":{"clusterIP":"10.96.0.10","ports":[{"name":"dns","port":53,"protocol":"UDP","targetPort":53},{"name":"dns-tcp","port":53,"protocol":"TCP","targetPort":53},{"name":"metrics","port":9153,"protocol":"TCP","targetPort":9153}],"selector":{"k8s-app":"kube-dns"},"sessionAffinity":"None","type":"ClusterIP"},"status":{"loadBalancer":{}}}],"kind":"ServiceList","metadata":{"resourceVersion":"377360","selfLink":"/api/v1/services"}}
```

​	部署configmap，并创建Istio的用户自定义资源

```shell
$ kubectl apply -f 03-configmap.yaml
configmap/edgemesh-cfg created
$ kubectl apply -f istio-crds-simple.yaml
customresourcedefinition.apiextensions.k8s.io/virtualservices.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/destinationrules.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/serviceentries.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/gateways.networking.istio.io created
```

​	使用daemonset的方式来部署edgemesh

```shell
$ kubectl apply -f 05-daemonset.yaml
daemonset.apps/edgemesh created
```



#### 测试样例

​	**HTTP协议**

​	在边缘节点上，部署支持http协议的容器应用和相关服务

```shell
$ kubectl apply -f hostname.yaml
```

​	到边缘节点上，使用curl去访问相关服务，打印出容器的hostname

```shell
$ curl hostname-lb-svc.edgemesh-test:12345
```



​	**TCP协议**

​	在边缘节点1，部署支持tcp协议的容器应用和相关服务

```shell
$ kubectl apply -f tcp-echo-service.yaml
```

​	在边缘节点2，使用telnet去访问相关服务

```shell
$ telnet tcp-echo-service.edgemesh-test 2701
```



​	**Websocket协议**

​	在边缘节点1，部署支持websocket协议的容器应用和相关服务

```shell
$ kubectl apply -f websocket-pod-svc.yaml
```

​	进入websocket的容器环境，并使用client去访问相关服务

```shell
$ docker exec -it 2a6ae1a490ae bash
$ ./client --addr ws-svc.edgemesh-test:12348
```



​	**负载均衡**

​	使用DestinationRule中的loadBalancer属性来选择不同的负载均衡模式

```shell
$ vim edgemesh-gateway-dr.yaml
spec
..
  trafficPolicy:
    loadBalancer:
      simple: RANDOM
..    
```



## Edgemesh Ingress Gateway

​	Edgemesh ingress gateway 提供了外部访问集群里服务的能力。

![image-20210414152916134](.\em-004.png)

#### 部署

​	创建istio的用户自定义资源

```shell
$ kubectl apply -f istio-crds-simple.yaml
customresourcedefinition.apiextensions.k8s.io/virtualservices.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/destinationrules.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/serviceentries.networking.istio.io created
customresourcedefinition.apiextensions.k8s.io/gateways.networking.istio.io created
```

​	配置configmap，并使用deployment来部署edgemesh-gateway

```shell
$ kubectl apply -f 03-configmap.yaml 
configmap/edgemesh-gateway-cfg created
$ kubectl apply -f 04-deployment.yaml 
deployment.apps/edgemesh-gateway created
```

​	创建gateway资源对象，和路由规则Virtual Service

```shell
$ kubectl apply -f edgemesh-gateway-gw-vsvc.yaml
gateway.networking.istio.io/edgemesh-gateway created
virtualservice.networking.istio.io/edgemesh-gateway-vsvc created
```

​	查看edgemesh-gateway是否部署成功

```shell
$ kubectl get gw -n edgemesh-test
NAME               AGE
edgemesh-gateway   3m30s
```

​	最后，使用IP和Virtual Service暴露的端口来进行访问

```shell
$ curl 192.168.0.211:23333
```



## 联系方式

如果您需要支持，请从 '操作指导' 开始，然后按照我们概述的流程进行操作。

如果您有任何疑问，请以下方式与我们联系：

​	[Bilibili KubeEdge](https://space.bilibili.com/448816706?from=search&seid=10057261257661405253)
