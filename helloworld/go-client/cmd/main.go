/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	greet "dubbo-mesh/helloworld/proto"
	"dubbo.apache.org/dubbo-go/v3/client"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"github.com/dubbogo/gost/log/logger"
)

func main() {
	//cli, err := client.NewClient(
	//	client.WithClientURL("127.0.0.1:20000"),
	//)

	url := "xds://httpbin.dubbo.svc.cluster.local:8000"
	if newUrl, ok := os.LookupEnv("DUBBO_SERVER_URL"); ok {
		url = newUrl
	}

	cli, err := client.NewClient(
		client.WithClientURL(url),
	)

	if err != nil {
		panic(err)
	}

	svc, err := greet.NewGreetService(cli)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
		resp, err := svc.Greet(context.Background(), &greet.GreetRequest{Name: "hello world"})
		if err != nil {
			logger.Error(err)
			w.Write([]byte(fmt.Sprintf("response error %v", err)))
			return
		}
		logger.Infof("Greet response: %s", resp.Greeting)
		w.Write([]byte(resp.Greeting))
	})

	httpSrv := &http.Server{Addr: ":9090", Handler: mux}
	httpSrv.ListenAndServe()

}
