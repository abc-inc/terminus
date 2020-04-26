/*
 * Copyright 2020 The terminus authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package iface

import (
	"errors"
	"github.com/c-robinson/iplib"
	"net"
	"strings"
)

const (
	// Broadcast address
	Broadcast = "broadcast"
	// First usable IP address of the subnet
	First = "first"
	// IP address
	IP = "ip"
	// Last usable IP address of the subnet
	Last = "last"
	// Name of the interface
	Name = "name"
	// NetMask of the subnet
	NetMask = "netmask"
	// Network address
	Network = "network"
	// Prefix in bits
	Prefix = "prefix"
	// Size of the subnet
	Size = "size"
	// UsableSize of the subnet
	UsableSize = "usable"
	// Version of the IP address
	Version = "version"
	// Wildcard mask
	Wildcard = "wildcard"
)

// GetAddr returns the first IPv4 unicast address for the interface specified by name.
func GetAddr(name string) (ip net.IP, n iplib.Net, err error) {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return ip, n, errors.Unwrap(err)
	}
	addrs, err := i.Addrs()
	if err != nil {
		return ip, n, errors.Unwrap(err)
	}
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok {
			if size, bits := n.Mask.Size(); bits == 32 {
				return n.IP, iplib.NewNet(n.IP, size), nil
			}
		}
	}
	return ip, n, errors.New("no IP address")
}

// GetParams returns the parameters for the specified IP.
func GetParams(name string, ip net.IP, mask net.IPMask) (m map[string]interface{}) {
	size, _ := mask.Size()
	n := iplib.NewNet(ip, size)

	m = make(map[string]interface{})
	m[Broadcast] = n.BroadcastAddress()
	m[First] = n.FirstAddress()
	m[Name] = name
	if ip.String() == strings.SplitN(name, "/", 2)[0] {
		m[Name] = findInterface(ip)
	}
	m[Network] = n.NetworkAddress()
	m[IP] = ip
	m[Last] = n.LastAddress()
	m[NetMask] = net.IP(mask)
	m[Prefix] = size
	m[Size] = int(n.Count4() + 2)
	m[UsableSize] = int(n.Count())
	m[Version] = n.Version()
	m[Wildcard] = net.IP(n.Wildcard())

	// special handling for /32 and /31
	if size == 32 {
		m[Size] = 1
	} else if size == 31 {
		m[Size] = 2
	}

	return m
}

func findInterface(ip net.IP) string {
	ifs, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, i := range ifs {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			if ia, ok := a.(*net.IPNet); ok && ip.Equal(ia.IP) {
				return i.Name
			}
		}
	}
	return ""
}
