package utils

import (
	"os"
	"strings"
)

const (
	EnvServiceName = "SERVICE_NAME"
	EnvPodName     = "POD_NAME"
	EnvSubSystem   = "SUB_SYSTEM"
	EnvNameSpace   = "POD_NAMESPACE"
	EnvVersion     = "VERSION"
	EnvPODIP       = "INSTANCE_IP"
	EnvNodeName    = "NODE_NAME"
	DUBBOSERVERURL = "DUBBO_SERVER_URL"
)

func GetAllEnvs() map[string]string {
	allEnvs := make(map[string]string, 2)
	envs := os.Environ()
	for _, e := range envs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		} else {
			allEnvs[parts[0]] = parts[1]
		}
	}
	return allEnvs
}

func GetHostName() string {
	return GetStringEnv(EnvPodName, GetDefaultHostName())
}

func GetNodeName() string {
	return GetStringEnv(EnvNodeName, GetDefaultHostName())
}

func GetNameSpace() string {
	return GetStringEnv(EnvNameSpace, "")
}

func GetVersion() string {
	return GetStringEnv(EnvVersion, "")
}

func GetIP() string {
	return GetStringEnv(EnvPODIP, "")
}

func GetServiceName() string {
	return GetStringEnv(EnvServiceName, GetDefaultHostName())
}

func GetSubSystem() string {
	return GetStringEnv(EnvSubSystem, "")
}

func GetDefaultHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func GetDUBBOServerUrl() string {
	return GetStringEnv(DUBBOSERVERURL, "xds://httpbin.dubbo.svc.cluster.local:8000")
}

func GetStringEnv(name string, defvalue string) string {
	val, ex := os.LookupEnv(name)
	if ex {
		return val
	} else {
		return defvalue
	}
}

func FileExisted(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return true
	}
	return true
}

func ConvertAttachmentsToMap(attachments map[string]interface{}) map[string]string {
	dataMap := make(map[string]string, 0)
	for k, attachment := range attachments {
		if v, ok := attachment.([]string); ok {
			dataMap[k] = v[0]
		}
		if v, ok := attachment.(string); ok {
			dataMap[k] = v
		}
	}
	return dataMap
}
