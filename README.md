# dubbo-mesh
## prebuild

build 项目之前，需要下载项目依赖库 istio 和 dubbo-go 特定分支到本地 external 目录
执行 make prebuild

```shell
make prebuild
```


## build images

```
  image-buildx-client  Build and push docker image for the dubbo client for cross-platform support
  image-buildx-httpbin  Build and push docker image for the dubbo httpbin for cross-platform support
  image-buildx-pilot-agent  Build and push docker image for the pilot agent for cross-platform support
  image-buildx-sleep  Build and push docker image for the sleep for cross-platform support
  prebuild         prebuild project

```

## 测试环境

![image](./deploy/httpbin/img.png)

[测试文档说明连接](./deploy/httpbin/README.md) 

# Reference 
- [xds-protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#xds-protocol)
- [filter chain match](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#envoy-v3-api-field-config-listener-v3-filterchain-filter-chain-match)
- [Istio 流量管理实现机制深度解析](https://cloudnative.to/blog/istio-traffic-management-impl-intro/)
- [component and port](https://tetrate.io/blog/istio-component-ports-and-functions-in-detail/)
- [Sidecar injection, transparent traffic hijacking, and routing process in Istio explained in detail](https://jimmysongio.medium.com/sidecar-injection-transparent-traffic-hijacking-and-routing-process-in-istio-explained-in-detail-d53e244e0348)
- https://www.modb.pro/db/128868