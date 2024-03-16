FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update  \
        && apt-get install --no-install-recommends -y ca-certificates curl iptables iproute2 iputils-ping knot-dnsutils netcat tcpdump conntrack bsdmainutils net-tools lsof sudo  \
        && update-ca-certificates  \
        && apt-get upgrade -y  \
        && apt-get clean  \
        && rm -rf /var/log/*log /var/lib/apt/lists/* /var/log/apt/* /var/lib/dpkg/*-old /var/cache/debconf/*-old  \
        && update-alternatives --set iptables /usr/sbin/iptables-legacy  \
        && update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
RUN useradd -m --uid 1337 istio-proxy  \
        && echo "istio-proxy ALL=NOPASSWD: ALL" >> /etc/sudoers

ARG TARGETARCH
COPY ${TARGETARCH:-amd64}/pilot-agent /usr/local/bin/pilot-agent

# The pilot-agent will bootstrap Envoy.
ENTRYPOINT ["/usr/local/bin/pilot-agent"]