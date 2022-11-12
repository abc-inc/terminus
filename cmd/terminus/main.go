// Copyright 2020 The Terminus authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/abc-inc/terminus/iface"
	"github.com/c-robinson/iplib"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var version = "0"

var rootCmd = &cobra.Command{
	Use: `terminus [flags] IP
  terminus [flags] IP/PREFIX_LEN
  terminus [flags] INTERFACE
  terminus [-L | --list-interfaces]`,
	Short: "terminus is an IP subnet address calculator.",
	Long: `terminus is an IP subnet address calculator.
For a given IPv4 address (and optional prefix length), ` +
		`it calculates network address, broadcast address, maximum number of hosts, etc.`,
	Run: runRootCmd,
	Example: `  terminus -i eth0                # 172.16.57.200
  terminus -p 10.0.0.138          # 8
  terminus -b 192.168.100.1/24    # 192.168.100.255

  terminus -f -l lo
  # 127.0.0.1
  # 127.255.255.254

  terminus -L
  # eth0    172.16.57.200   172.16.56.0     23
  # lo      127.0.0.1       127.0.0.0       8

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
	rootCmd.Flags().BoolP("list-interfaces", "L", false, "List all network interfaces")
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

	if args, err := readFromPipe(); err != nil {
		log.Fatal(err)
	} else if args != nil {
		rootCmd.SetArgs(append(os.Args[1:], args...))
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func readFromPipe() ([]string, error) {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Size() == 0 || fi.Mode()&os.ModeNamedPipe == 0 {
		// nothing available in pipe - continue
		return nil, nil
	}

	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	return shellquote.Split(strings.TrimRight(string(in), "\n"))
}

func runRootCmd(cmd *cobra.Command, args []string) {
	switch {
	case cmd.Flag("version").Changed:
		_, _ = fmt.Fprintln(os.Stderr, "terminus version", version)
		return
	case cmd.Flag("list-interfaces").Changed:
		fmt.Print(listInterfaces())
		return
	case strings.Contains(cmd.Flag("template").Value.String(), ".interfaces"):
		// if the template refers to interfaces by name, the positional argument is optional
	case len(args) == 0:
		_ = cmd.Usage()
		os.Exit(1)
	}

	data := map[string]interface{}{}
	if len(args) > 0 {
		arg := args[len(args)-1]
		ip, n, err := determineIP(arg)
		if err != nil {
			log.Fatal(err)
		}
		data = iface.GetParams(arg, ip, n.Mask)
	}

	s := &strings.Builder{}
	cmd.Flags().Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "range":
			_, _ = fmt.Fprintf(s, "%v - %v\n", data[iface.Network], data[iface.Broadcast])
		case "template":
			text, _ := cmd.Flags().GetString("template")
			printTemplate(text, s, data)
		default:
			_, _ = fmt.Fprintln(s, data[f.Name])
		}
	})
	fmt.Print(s)
}

func listInterfaces() string {
	is, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(is, func(i, j int) bool { return is[i].Name < is[j].Name })
	s := &strings.Builder{}
	for _, i := range is {
		if ip, n, err := determineIP(i.Name); err == nil {
			data := iface.GetParams(i.Name, ip, n.Mask)
			_, _ = fmt.Fprintf(s, "%s\t%v\t%v\t%v\n", data[iface.Name], data[iface.IP], data[iface.Network], data[iface.Prefix])
		}
	}
	return s.String()
}

func determineIP(arg string) (net.IP, iplib.Net, error) {
	ip := net.ParseIP(arg)
	if ip != nil {
		size, _ := ip.DefaultMask().Size()
		return ip, iplib.NewNet(ip, size), nil
	}

	ip, ipNet, err := net.ParseCIDR(arg)
	if err == nil {
		size, _ := ipNet.Mask.Size()
		return ip, iplib.NewNet(ip, size), nil
	}

	ip, n, err := iface.GetAddr(arg)
	if err != nil {
		return nil, n, err
	}
	return ip, n, nil
}

func printTemplate(text string, w io.Writer, data map[string]interface{}) {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}

	t, err := template.New("tmpl").
		Option("missingkey=zero").
		Funcs(template.FuncMap{
			"toBinary": toBinary,
			"toHex":    toHex,
			"toJson":   toJSON,
		}).Parse(text)

	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(text, ".interfaces") {
		ifByName := map[string]interface{}{}
		data["interfaces"] = ifByName

		is, _ := net.Interfaces()
		for _, i := range is {
			ip, n, _ := iface.GetAddr(i.Name)
			ifByName[i.Name] = iface.GetParams(i.Name, ip, n.Mask)
		}
	}

	if err := t.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

func toBinary(ip net.IP) string {
	ip = ip.To4()
	return fmt.Sprintf("%08b.%08b.%08b.%08b", ip[0], ip[1], ip[2], ip[3])
}

func toHex(ip net.IP) string {
	return "0x" + net.IPMask(ip.To4()).String()
}

func toJSON(i interface{}) string {
	j, err := json.Marshal(i)
	if err != nil {
		log.Fatal(err)
	}
	return string(j)
}
