## 部署 edgemesh到k8s集群
此方式将在容器中运行edgemesh

0. edgemesh需要先于edgecore部署
1. （临时）在运行edgemesh的节点上放置k8s集群的config文件，放置于/root/.kube/config
2. 编译容器镜像
```shell
cd /edgemesh-test # 进入项目目录
docker build -t edgemesh:0.1 -f build/Dockerfile  .
```
3. 执行部署
```shell 
# 请先按实际请求修改yaml里的节点信息
kubectl apply -f 03-configmap.yaml
kubectl apply -f 04-deployment.yaml
```
