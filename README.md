# terminus
terminus is an IPv4 subnet address calculator

It can be used
* as a standalone tool to output human readable information about a network or address,
* as a tool suitable to be used by scripts or other programs via [template language](#Template-Language),
* as well as a library for applications written in Go.

## Why *terminus*?
In Roman religion, *Terminus* was the god who protected boundary stones.
His name was the Latin word for such a marker.
Therefore, it is the ideal name of an application, which calculates network boundaries.

Besides, there are a few command-line applications, which output information about IP addresses in human readable form.
However, they lack flexibility when it comes to output format.

## Features
For a given IPv4 address (and optional prefix length), *terminus* calculates
* IP / netmask in dot-decimal notation
* prefix length
* broadcast address / network address
* host address range of the subnet
* maximum number of hosts / IPs in the subnet
* wildcard mask in dot-decimal notation

## Examples
```shell script
$ terminus -i eth0
172.16.57.200

$ terminus -p 10.0.0.138
8

$ terminus -b 192.168.100.1/24
192.168.100.255

$ terminus -f -l lo
127.0.0.1
127.255.255.254

$ terminus -t '{{.ip}}/{{.prefix}} ({{.network}} - {{.broadcast}})' tun0
10.197.63.254/11 (10.192.0.0 - 10.223.255.255)
```

## Template Language
When given a template expression, *terminus* evaluates it using [Go templates](https://golang.org/pkg/text/template/#pkg-overview).
Thus, all actions, expression and functions available in a Go template can be used.
For example, the following statement checks if the environment variable `IP` is a loopback address.
If so, it replaces the IP with "`localhost`" and yields a nice URL:
```shell script
$ terminus -t '{{if .ip.IsLoopback}}http://localhost{{else}}https://{{.ip}}{{end}}:8080/' "${IP}"
```

The following variables can be used:
```gotemplate
Expression      Example         Type    Description
{{.broadcast}}  10.0.3.255      net.IP  broadcast address
{{.first}}      10.0.0.1        net.IP  first usable IP address of the subnet
{{.ip}}         10.0.0.42       net.IP  IP address
{{.last}}       10.0.3.254      net.IP  last usable IP address of the subnet
{{.name}}       eth0            string  name of the network interface
{{.netmask}}    255.255.252.0   net.IP  subnet mask
{{.network}}    10.0.0.0        net.IP  network address
{{.prefix}}     22              int     prefix length
{{.size}}       1024            int     size of the subnet
{{.usable}}     1022            int     usable size of the subnet (host count)
{{.wildcard}}   0.0.3.255       net.IP  wildcard mask
```

## Roadmap
* IPv6 support (including conversions)
* support for multiple interfaces in template expressions
* additional output formats (binary, hex) where it makes sense
* subnet splitting

## Related Projects (in alphabetical order)
* ipcalc: https://gitlab.com/ipcalc/ipcalc
* netcalc: https://github.com/troglobit/netcalc
* sipcalc: https://github.com/sii/sipcalc
* subnetcalc: https://github.com/dreibh/subnetcalc
