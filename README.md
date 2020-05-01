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

# using a template expression
$ terminus -t "{{.ip}}/{{.prefix}} ({{.network}} - {{.broadcast}})" tun0
10.197.63.254/11 (10.192.0.0 - 10.223.255.255)
```

## Template Language
When given a template expression, *terminus* evaluates it using [Go templates](https://golang.org/pkg/text/template/#pkg-overview).
Thus, all actions, expressions and functions available in a Go template can be used.

For example, the following statement checks if the environment variable `IP` is a loopback address.
If so, it replaces the IP with "`localhost`", and it yields a nice URL:
```shell script
$ export IP="127.100.100.1"
$ terminus -t "{{if .ip.IsLoopback}}http://localhost{{else}}https://{{.ip}}{{end}}:8080/" "${IP}"
http://localhost:8080/
```

### Network Interfaces
If an IP address or network interface is passed as a command line argument, it is set as the *default interface*.
In this way it can be accessed with `.` e.g., `{{.ip}}`.

In general, all interfaces can be addressed by name via `{{.interfaces.<NAME>}}` e.g., `{{.interfaces.tun0.ip}}`.
In other words, the following expressions are equivalent:
```shell script
$ terminus -t "{{.network}} - {{.broadcast}}" tun0
$ terminus -t "{{.interfaces.tun0.network}} - {{.interfaces.tun0.broadcast}}"
```

While the short syntax is preferred for retrieving multiple properties of a single network interface,
the long syntax can be used for retrieving properties of multiple network interfaces.
```shell script
$ terminus -t '{{.interfaces.eth0.ip}}{{"\n"}}{{.interfaces.eth1.ip}}'
172.16.57.200
192.168.100.1
```

#### Windows
Network adapters can contain whitespaces e.g., "`Local Area Connection`".
Therefore, it is essential to quote them in template expressions.
The Windows command interpreter *cmd* requires two double quotes to quote a literal string e.g., `""Local Area Connection""`.
Due to the nested quotes, the [`index` function](https://golang.org/pkg/text/template/#hdr-Functions) must be used to address properties.
This adds additional double quotes to the literals:
```
>terminus -t "{{index .interfaces """Local Area Connection""" """ip"""}} {{index .interfaces """Local Area Connection""" """netmask"""}}"
192.168.100.200 255.255.255.0
```

### Network Interface Properties
Every network interface has the following properties:
```
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
Note that values might be absent if an interface is not up.

### Functions
*terminus* comes with built-in functions for converting IP addresses and netmasks, and for formatting output.
The following functions are available:
* `toBinary`: converts an IP address (or netmask) to binary dot-decimal notation
* `toHex`:  converts a netmask (or IP address) to hexadecimal notation
* `toJson`: converts the input to a valid JSON object/array/string (if possible)

```shell script
$ terminus -t '{{.ip}} {{.ip | toBinary}}{{"\n"}}{{.netmask}} {{.netmask | toHex}}' eth0
172.16.57.200 10101100.00010000.00111001.11001000
255.255.254.0 0xfffffe00
```

The following example dumps all properties of the *eth0* interface as JSON object:
```shell script
$ terminus -t "{{. | toJson}}" eth0
{"broadcast":"172.16.57.255","first":"172.16.56.1","ip":"172.16.57.200","last":"172.16.57.254","name":"eth0","netmask":"255.255.254.0","network":"172.16.56.0","prefix":23,"size":512,"usable":510,"version":4,"wildcard":"0.0.1.255"}
```

The `toJson` function comes in handy when combined with other tools like *[jq](https://stedolan.github.io/jq/)*.
For example, a JSON array of all network interfaces and their properties can be created as follows:
```shell script
$ terminus -L | cut -f 1 | xargs -n1 terminus -t "{{. | toJson}}" | jq -rs "."
[
  {
    "broadcast": "172.16.57.255",
    "first": "172.16.56.1",
    "ip": "172.16.57.200",
    "last": "172.16.57.254",
    "name": "eth0",
    "netmask": "255.255.254.0",
    "network": "172.16.56.0",
    "prefix": 23,
    "size": 512,
    "usable": 510,
    "version": 4,
    "wildcard": "0.0.1.255"
  },
  ...
]
```

## Pipes & stdin
When using *terminus* in a pipeline, the output of the previous command is appended to the arguments passed to *terminus*.
Note that template expressions must be quoted to prevent word splitting.

The following example reads the template from a file and parametrizes it with the interface *eth0*:
```shell script
$ cat macOS_format.tmpl
-t "inet {{.ip}} netmask {{.netmask | toHex}} broadcast {{.broadcast}}"
$ cat macOS_format.tmpl | terminus eth0
inet 172.16.57.200 netmask 0xfffffe00 broadcast 172.16.57.255
# equivalent to
# terminus -t "inet {{.ip}} netmask {{.netmask | toHex}} broadcast {{.broadcast}}" eth0
````

Also *xargs* can be useful e.g., for iterating over all ethernet adapters and printing their IP addresses:
```shell script
$ terminus -L | cut -f 1 | grep -E "^eth" | xargs -n1 terminus -t '{{.name}}{{"\t"}}{{.ip}}'
eth0  172.16.57.200
eth0:0  127.16.57.100
eth1  192.168.100.1
```

## Roadmap
* IPv6 support (including conversions)
* subnet splitting

## Related Projects (in alphabetical order)
* ipcalc: https://gitlab.com/ipcalc/ipcalc
* netcalc: https://github.com/troglobit/netcalc
* sipcalc: https://github.com/sii/sipcalc
* subnetcalc: https://github.com/dreibh/subnetcalc
