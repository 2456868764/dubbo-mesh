package app

import (
	"context"
	"dubbo-mesh/pilot/cmd/pilot-agent/config"
	"dubbo-mesh/pilot/cmd/pilot-agent/options"
	istio_agent "dubbo-mesh/pilot/pkg/istio-agent"
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"istio.io/api/annotation"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/util/network"
	"istio.io/istio/pkg/bootstrap"
	"istio.io/istio/pkg/config/constants"
	"istio.io/istio/pkg/envoy"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/util/sets"
	"istio.io/istio/security/pkg/stsservice/tokenmanager"
	"net"
	"net/netip"
	"strings"
)

const (
	localHostIPv4 = "127.0.0.1"
	localHostIPv6 = "::1"
)

var (
	loggingOptions = log.DefaultOptions()
	proxyArgs      options.ProxyArgs
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "pilot-agent",
		Short:        "Istio Pilot agent.",
		Long:         "Istio Pilot agent runs in the sidecar or gateway container and bootstraps Envoy.",
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	proxyCmd := newProxyCommand()
	addFlags(proxyCmd)
	rootCmd.AddCommand(proxyCmd)
	return rootCmd
}

func newProxyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "proxy",
		Short: "XDS proxy agent",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
		RunE: func(c *cobra.Command, args []string) error {
			proxy, err := initProxy(args)
			if err != nil {
				return err
			}
			proxyConfig, err := config.ConstructProxyConfig(proxyArgs.MeshConfigFile, proxyArgs.ServiceCluster, options.ProxyConfigEnv, proxyArgs.Concurrency)
			if err != nil {
				return fmt.Errorf("failed to get proxy config: %v", err)
			}

			secOpts, err := options.NewSecurityOptions(proxyConfig, proxyArgs.StsPort, proxyArgs.TokenManagerPlugin)
			if err != nil {
				return err
			}

			// If we are using a custom template file (for control plane proxy, for example), configure this.
			if proxyArgs.TemplateFile != "" && proxyConfig.CustomConfigFile == "" {
				proxyConfig.ProxyBootstrapTemplatePath = proxyArgs.TemplateFile
			}

			envoyOptions := envoy.ProxyConfig{
				LogLevel:          proxyArgs.ProxyLogLevel,
				ComponentLogLevel: proxyArgs.ProxyComponentLogLevel,
				LogAsJSON:         loggingOptions.JSONEncoding,
				NodeIPs:           proxy.IPAddresses,
				Sidecar:           proxy.Type == model.SidecarProxy,
				OutlierLogPath:    proxyArgs.OutlierLogPath,
			}
			agentOptions := options.NewAgentOptions(proxy, proxyConfig)
			agent := istio_agent.NewAgent(proxyConfig, agentOptions, secOpts, envoyOptions)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			defer agent.Close()

			// Start in process SDS, dns server, xds proxy, and Envoy.
			wait, err := agent.Run(ctx)
			if err != nil {
				return err
			}
			wait()
			return nil
		},
	}
}

func addFlags(proxyCmd *cobra.Command) {
	proxyArgs := options.NewProxyArgs()
	proxyCmd.PersistentFlags().StringVar(&proxyArgs.DNSDomain, "domain", "",
		"DNS domain suffix. If not provided uses ${POD_NAMESPACE}.svc.cluster.local")
	proxyCmd.PersistentFlags().StringVar(&proxyArgs.MeshConfigFile, "meshConfig", "./etc/istio/config/mesh",
		"File name for Istio mesh configuration. If not specified, a default mesh will be used. This may be overridden by "+
			"PROXY_CONFIG environment variable or proxy.istio.io/config annotation.")
	proxyCmd.PersistentFlags().IntVar(&proxyArgs.StsPort, "stsPort", 0,
		"HTTP Port on which to serve Security Token Service (STS). If zero, STS service will not be provided.")
	proxyCmd.PersistentFlags().StringVar(&proxyArgs.TokenManagerPlugin, "tokenManagerPlugin", tokenmanager.GoogleTokenExchange,
		"Token provider specific plugin name.")
	// DEPRECATED. Flags for proxy configuration
	//proxyCmd.PersistentFlags().StringVar(&proxyArgs.ServiceCluster, "serviceCluster", constants.ServiceClusterName, "Service cluster")
	// Log levels are provided by the library https://github.com/gabime/spdlog, used by Envoy.
	proxyCmd.PersistentFlags().StringVar(&proxyArgs.ProxyLogLevel, "proxyLogLevel", "warning,misc:error",
		fmt.Sprintf("The log level used to start the Envoy proxy (choose from {%s, %s, %s, %s, %s, %s, %s})."+
			"Level may also include one or more scopes, such as 'info,misc:error,upstream:debug'",
			"trace", "debug", "info", "warning", "error", "critical", "off"))
	proxyCmd.PersistentFlags().IntVar(&proxyArgs.Concurrency, "concurrency", 0, "number of worker threads to run")
	// See https://www.envoyproxy.io/docs/envoy/latest/operations/cli#cmdoption-component-log-level
	proxyCmd.PersistentFlags().StringVar(&proxyArgs.ProxyComponentLogLevel, "proxyComponentLogLevel", "",
		"The component log level used to start the Envoy proxy. Deprecated, use proxyLogLevel instead")
	proxyCmd.PersistentFlags().StringVar(&proxyArgs.TemplateFile, "templateFile", "",
		"Go template bootstrap config")
	proxyCmd.PersistentFlags().StringVar(&proxyArgs.OutlierLogPath, "outlierLogPath", "",
		"The log path for outlier detection")
	proxyCmd.PersistentFlags().BoolVar(&proxyArgs.EnableProfiling, "profiling", true,
		"Enable profiling via web interface host:port/debug/pprof/.")
}

func initProxy(args []string) (*model.Proxy, error) {
	proxy := &model.Proxy{
		Type: model.SidecarProxy,
	}
	if len(args) > 0 {
		proxy.Type = model.NodeType(args[0])
		if !model.IsApplicationNodeType(proxy.Type) {
			return nil, fmt.Errorf("Invalid proxy Type: " + string(proxy.Type))
		}
	}

	podIP, _ := netip.ParseAddr(options.InstanceIPVar.Get()) // protobuf encoding of IP_ADDRESS type
	if podIP.IsValid() {
		proxy.IPAddresses = []string{podIP.String()}
	}

	// Obtain all the IPs from the node
	if ipAddrs, ok := network.GetPrivateIPs(context.Background()); ok {
		if len(proxy.IPAddresses) == 1 {
			for _, ip := range ipAddrs {
				// prevent duplicate ips, the first one must be the pod ip
				// as we pick the first ip as pod ip in istiod
				if proxy.IPAddresses[0] != ip {
					proxy.IPAddresses = append(proxy.IPAddresses, ip)
				}
			}
		} else {
			proxy.IPAddresses = append(proxy.IPAddresses, ipAddrs...)
		}
	}

	// No IP addresses provided, append 127.0.0.1 for ipv4 and ::1 for ipv6
	if len(proxy.IPAddresses) == 0 {
		proxy.IPAddresses = append(proxy.IPAddresses, localHostIPv4, localHostIPv6)
	}

	// Apply exclusions from traffic.sidecar.istio.io/excludeInterfaces
	proxy.IPAddresses = applyExcludeInterfaces(proxy.IPAddresses)

	// After IP addresses are set, let us discover IPMode.
	proxy.DiscoverIPMode()

	// Extract pod variables.
	proxy.ID = proxyArgs.PodName + "." + proxyArgs.PodNamespace

	// If not set, set a default based on platform - podNamespace.svc.cluster.local for
	// K8S
	proxy.DNSDomain = getDNSDomain(proxyArgs.PodNamespace, proxyArgs.DNSDomain)
	log.WithLabels("ips", proxy.IPAddresses, "type", proxy.Type, "id", proxy.ID, "domain", proxy.DNSDomain).Info("Proxy role")

	return proxy, nil
}

func getDNSDomain(podNamespace, domain string) string {
	if len(domain) == 0 {
		domain = podNamespace + ".svc." + constants.DefaultClusterLocalDomain
	}
	return domain
}

func applyExcludeInterfaces(ifaces []string) []string {
	// Get list of excluded interfaces from pod annotation
	// TODO: Discuss other input methods such as env, flag (ssuvasanth)
	annotations, err := bootstrap.ReadPodAnnotations("")
	if err != nil {
		log.Debugf("Reading podInfoAnnotations file to get excludeInterfaces was unsuccessful. Continuing without exclusions. msg: %v", err)
		return ifaces
	}
	value, ok := annotations[annotation.SidecarTrafficExcludeInterfaces.Name]
	if !ok {
		log.Debugf("ExcludeInterfaces annotation is not present. Proxy IPAddresses: %v", ifaces)
		return ifaces
	}
	exclusions := strings.Split(value, ",")

	// Find IP addr of excluded interfaces and add to a map for instant lookup
	exclusionMap := sets.New[string]()
	for _, ifaceName := range exclusions {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Warnf("Unable to get interface %s: %v", ifaceName, err)
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Warnf("Unable to get IP addr(s) of interface %s: %v", ifaceName, err)
			continue
		}

		for _, addr := range addrs {
			// Get IP only
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			// handling ipv4 wrapping in ipv6
			ipAddr, okay := netip.AddrFromSlice(ip)
			if !okay {
				continue
			}
			unwrapAddr := ipAddr.Unmap()
			if !unwrapAddr.IsValid() || unwrapAddr.IsLoopback() || unwrapAddr.IsLinkLocalUnicast() || unwrapAddr.IsLinkLocalMulticast() || unwrapAddr.IsUnspecified() {
				continue
			}

			// Add to map
			exclusionMap.Insert(unwrapAddr.String())
		}
	}

	// Remove excluded IP addresses from the input IP addresses list.
	var selectedInterfaces []string
	for _, ip := range ifaces {
		if exclusionMap.Contains(ip) {
			log.Infof("Excluding ip %s from proxy IPaddresses list", ip)
			continue
		}
		selectedInterfaces = append(selectedInterfaces, ip)
	}

	return selectedInterfaces
}
