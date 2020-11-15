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

package main

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/abc-inc/terminus/iface"
	. "github.com/stretchr/testify/require"
)

func TestPrintTemplate(t *testing.T) {
	tests := []struct {
		prop string
		want string
	}{
		{iface.Broadcast, "127.255.255.255"},
		{iface.First, "127.0.0.1"},
		{iface.IP, "127.255.255.255"},
		{iface.Last, "127.255.255.254"},
		{iface.Name, ""},
		{iface.NetMask, "255.0.0.0"},
		{iface.Network, "127.0.0.0"},
		{iface.Prefix, "8"},
		{iface.Size, strconv.Itoa(int(math.Pow(2, 24)))},
		{iface.UsableSize, strconv.Itoa(int(math.Pow(2, 24) - 2))},
		{iface.Version, "4"},
		{iface.Wildcard, "0.255.255.255"},
	}

	ip, n, _ := net.ParseCIDR("127.255.255.255/8")
	data := iface.GetParams(ip.String(), ip, n.Mask)

	for i := range tests {
		tt := tests[i]
		t.Run(tt.prop, func(t *testing.T) {
			s := &strings.Builder{}
			printTemplate("{{."+tt.prop+"}}", s, data)
			Equal(t, tt.want+"\n", s.String())
		})
	}
}

func TestPrintTemplateInterfaces(t *testing.T) {
	ip, n, _ := net.ParseCIDR("127.0.0.1/8")
	data := iface.GetParams(ip.String(), ip, n.Mask)
	s := &strings.Builder{}
	printTemplate(fmt.Sprintf("{{.interfaces.%s.ip}}", data[iface.Name]), s, data)
	Equal(t, ip.String()+"\n", s.String())
}

func TestPrintTemplateFunctions(t *testing.T) {
	tests := []struct {
		tmpl string
		want string
	}{
		{"{{.ip | toBinary}}", "01111111.00000000.00000000.00000001"},
		{"{{.netmask | toHex}}", "0xffffff00"},
		{"{{.prefix | toJson}}", "24"},
		{"{{.ip | toJson}}", "\"127.0.0.1\""},
		{"{{.ip | toBinary | toJson}}", "\"01111111.00000000.00000000.00000001\""},
	}

	ip, n, _ := net.ParseCIDR("127.0.0.1/24")
	data := iface.GetParams(ip.String(), ip, n.Mask)

	for i := range tests {
		tt := tests[i]
		t.Run(tt.tmpl, func(t *testing.T) {
			s := &strings.Builder{}
			printTemplate(tt.tmpl, s, data)
			Equal(t, tt.want+"\n", s.String())
		})
	}
}

func TestPrintTemplateNoData(t *testing.T) {
	data := map[string]interface{}{}
	s := &strings.Builder{}
	printTemplate("{{.name}}", s, data)
	Equal(t, "<no value>\n", s.String())
}

func TestListInterfaces(t *testing.T) {
	is, err := net.Interfaces()
	NoError(t, err)
	NotEmpty(t, is)

	s := listInterfaces()
	Contains(t, s, "127.0.0.1")

	for _, i := range is {
		if ip, _, err := iface.GetAddr(i.Name); err == nil {
			Contains(t, s, i.Name)
			Contains(t, s, ip.String())
		}
	}
}

func TestDetermineIP(t *testing.T) {
	ip, n, err := determineIP("127.0.100.1")
	Equal(t, "127.0.100.1", ip.String())
	Equal(t, "127.0.0.0", n.IP.String())
	Equal(t, "ff000000", n.Mask.String())
	NoError(t, err)
}

func TestDetermineIPCIDR(t *testing.T) {
	ip, n, err := determineIP("127.0.100.1/24")
	Equal(t, "127.0.100.1", ip.String())
	Equal(t, "127.0.100.0", n.IP.String())
	Equal(t, "ffffff00", n.Mask.String())
	NoError(t, err)
}
