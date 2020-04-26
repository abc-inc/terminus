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
	"github.com/abc-inc/terminus/iface"
	"github.com/c-robinson/iplib"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"
	"net"
	"os"
	"strings"
	"text/template"
)

var version = "0"

var rootCmd = &cobra.Command{
	Use: `terminus [flags] IP
  terminus [flags] IP/PREFIX_LEN
  terminus [flags] INTERFACE`,
	Short: "terminus is an IP subnet address calculator.",
	Long: `terminus is an IP subnet address calculator.
For a given IPv4 address (and optional prefix length), it calculates network address, broadcast address, maximum number of hosts, etc.`,
	Run: runRootCmd,
	Example: `  terminus -i eth0                # 172.16.57.200
  terminus -p 10.0.0.138          # 8
  terminus -b 192.168.100.1/24    # 192.168.100.255

  terminus -f -l lo
  # 127.0.0.1
  # 127.255.255.254

  terminus -t '{{.ip}}/{{.prefix}} ({{.network}} - {{.broadcast}})' tun0
  # 10.197.63.254/11 (10.192.0.0 - 10.223.255.255)`,
}

func main() {
	log.SetPrefix("terminus: ")
	log.SetFlags(0)
	Execute()
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().BoolP(iface.Broadcast, "b", false, "Show the broadcast address of the subnet")
	rootCmd.Flags().BoolP(iface.First, "f", false, "Show the first usable IP address of the subnet")
	rootCmd.Flags().BoolP("help", "h", false, "Print this help information and exit")
	rootCmd.Flags().BoolP(iface.IP, "i", false, "Show the IP address")
	rootCmd.Flags().BoolP(iface.Last, "l", false, "Show the last usable IP address of the subnet")
	rootCmd.Flags().BoolP(iface.NetMask, "m", false, "Show the subnet mask in dot-decimal notation")
	rootCmd.Flags().Bool(iface.Name, false, "Show the name of the network interface (if possible)")
	rootCmd.Flags().BoolP(iface.Network, "n", false, "Show the network address")
	rootCmd.Flags().BoolP(iface.Prefix, "p", false, "Show the prefix length")
	rootCmd.Flags().BoolP("range", "r", false, "Show the IP range of the subnet")
	rootCmd.Flags().BoolP(iface.Size, "s", false, "Count the total number of IPs of the subnet")
	rootCmd.Flags().StringP("template", "t", "", "Format the output with the given template expression")
	rootCmd.Flags().BoolP(iface.UsableSize, "u", false, "Count the number of hosts of the subnet")
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and exit")
	rootCmd.Flags().BoolP(iface.Wildcard, "w", false, "Show the wildcard mask of the subnet")
	rootCmd.Args = cobra.RangeArgs(0, 1)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runRootCmd(cmd *cobra.Command, args []string) {
	if cmd.Flag("version").Changed {
		_, _ = fmt.Fprintln(os.Stderr, "terminus version", version)
		os.Exit(0)
	} else if len(args) == 0 {
		_ = cmd.Usage()
		os.Exit(1)
	}

	arg := args[0]
	ip, n := determineIP(arg)
	data := iface.GetParams(arg, ip, n.Mask)
	s := &strings.Builder{}
	cmd.Flags().Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "range":
			_, _ = fmt.Fprintf(s, "%v - %v\n", data[iface.Network], data[iface.Broadcast])
		case "template":
			p, _ := cmd.Flags().GetString("template")
			if !strings.HasSuffix(p, "\n") {
				p += "\n"
			}
			t, err := template.New("tmpl").Parse(p)
			if err != nil {
				log.Fatal(err)
			}
			if err := t.Execute(s, data); err != nil {
				log.Fatal(err)
			}
		default:
			_, _ = fmt.Fprintln(s, data[f.Name])
		}
	})
	fmt.Print(s)
}

func determineIP(arg string) (net.IP, iplib.Net) {
	ip := net.ParseIP(arg)
	if ip != nil {
		size, _ := ip.DefaultMask().Size()
		return ip, iplib.NewNet(ip, size)
	}

	ip, ipNet, err := net.ParseCIDR(arg)
	if err == nil {
		size, _ := ipNet.Mask.Size()
		return ip, iplib.NewNet(ip, size)
	}

	ip, n, err := iface.GetAddr(arg)
	if err != nil {
		log.Fatal(err)
	}
	return ip, n
}
