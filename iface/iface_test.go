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

package iface_test

import (
	"fmt"
	"github.com/abc-inc/terminus/iface"
	. "github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestGetAddr(t *testing.T) {
	addr, n, err := iface.GetAddr("lo")
	if err != nil {
		addr, n, err = iface.GetAddr("lo0")
	}
	NoError(t, err)
	Equal(t, "127.0.0.1", addr.To4().String())
	Equal(t, "ff000000", n.Mask.String())
}

func TestGetAddrInvalidName(t *testing.T) {
	_, _, err := iface.GetAddr("")
	EqualError(t, err, "invalid network interface name: ")
}

func TestGetParams(t *testing.T) {
	i := net.ParseIP("192.168.0.1")
	m := iface.GetParams("eth0", i.To4(), net.CIDRMask(24, 32))
	EqualValues(t, "eth0", fmt.Sprint(m[iface.Name]))
	EqualValues(t, "192.168.0.1", fmt.Sprint(m[iface.IP]))
	EqualValues(t, "255.255.255.0", fmt.Sprint(m[iface.NetMask]))
	EqualValues(t, 24, m[iface.Prefix])
	EqualValues(t, "0.0.0.255", fmt.Sprint(m[iface.Wildcard]))
	EqualValues(t, "192.168.0.255", fmt.Sprint(m[iface.Broadcast]))
	EqualValues(t, 256, m[iface.Size])
	EqualValues(t, 254, m[iface.UsableSize])
	EqualValues(t, "192.168.0.1", fmt.Sprint(m[iface.First]))
	EqualValues(t, "192.168.0.254", fmt.Sprint(m[iface.Last]))
	EqualValues(t, "192.168.0.0", fmt.Sprint(m[iface.Network]))
	EqualValues(t, "4", fmt.Sprint(m[iface.Version]))
}
