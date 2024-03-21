# dubbo-mesh

# tri client
1. dubbo-go/protocol/triple/client.go

```golang
func newClientManager(url *common.URL) (*clientManager, error) {

	
// todo(DMwangnima): support TLS in an ideal way
	var cfg *tls.Config
	var tlsFlag bool

	var transport http.RoundTripper
	callType := url.GetParam(constant.CallHTTPTypeKey, constant.CallHTTP2)
	switch callType {
	case constant.CallHTTP:
		transport = &http.Transport{
			TLSClientConfig: cfg,
		}
		cliOpts = append(cliOpts, tri.WithTriple())
	case constant.CallHTTP2:
		if tlsFlag {
			transport = &http2.Transport{
				TLSClientConfig: cfg,
			}
		} else {
			transport = &http2.Transport{
				DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
				AllowHTTP: true,
			}
		}
	default:
		panic(fmt.Sprintf("Unsupported callType: %s", callType))
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	
}

```
2. registry & serviceDiscovery

dubbo-go/registry 目录加 istio 目录


# docker 

```shell
cd pilot/cmd/pilot-agent
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o pilot-agent  main.go
cp pilot-agent ../docker/
cd ../docker
docker build -t registry.cn-hangzhou.aliyuncs.com/2456868764/pilot-agent:1.0.0 .


```



# Reference 
- [xds-protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#xds-protocol)
- [filter chain match](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#envoy-v3-api-field-config-listener-v3-filterchain-filter-chain-match)
- [Istio 流量管理实现机制深度解析](https://cloudnative.to/blog/istio-traffic-management-impl-intro/)
- [component and port](https://tetrate.io/blog/istio-component-ports-and-functions-in-detail/)
- [Sidecar injection, transparent traffic hijacking, and routing process in Istio explained in detail](https://jimmysongio.medium.com/sidecar-injection-transparent-traffic-hijacking-and-routing-process-in-istio-explained-in-detail-d53e244e0348)