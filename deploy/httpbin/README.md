# Prepare

## 1. 创建命名空间：dubbo

```shell

$ kubectl create ns dubbo
$ kubectl apply -f dubboclient.yaml -n dubbo
$ kubectl apply -f dubbohttpbin.yaml -n dubbo
$ kubectl apply -f sleep.yaml -n dubbo

```

## 2. 创建命名空间：legacy

```shell

$ kubectl create ns legacy
$ kubectl apply -f dubboclient-legacy.yaml -n legacy
$ kubectl apply -f sleep-legacy.yaml -n legacy

```

## 3. 检查 PeerAuthentication, RequestAuthentication, AuthorizationPolicy 

```shell

```
# 测试 mtls & PeerAuthentication
## 默认

```shell
$ for from in "dubbo" "legacy"; do kubectl exec "$(kubectl get pod -l app=sleep -n ${from} -o jsonpath={.items..metadata.name})" -c sleep -n ${from} -- curl http://dubboclient.${from}.svc:9090/greet -s -o /dev/null -w "sleep.${from} to httpbin.${from}: %{http_code}\n"; done
sleep.foo to httpbin.foo: 200
sleep.foo to httpbin.bar: 200
sleep.bar to httpbin.foo: 200
sleep.bar to httpbin.bar: 200
sleep.legacy to httpbin.foo: 200
sleep.legacy to httpbin.bar: 200
```



# 测试 RequestAuthentication

```shell
SLEEP_POD=`kubectl get pod -l app=sleep -n dubbo -o jsonpath={.items..metadata.name}`
kubectl exec "${SLEEP_POD}" -c sleep -n dubbo -- curl http://dubboclient.dubbo.svc:9090/greet -s -H "x-hello-header: hello" | jq
```

```shell
kubectl apply -n dubbo -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
spec:
  mtls:
    mode: STRICT
EOF

kubectl apply -n dubbo -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
spec:
  mtls:
    mode: DISABLE
EOF

kubectl apply -n dubbo -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
spec:
  mtls:
    mode: PERMISSIVE
EOF
```





