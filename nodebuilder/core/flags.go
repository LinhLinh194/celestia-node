package core

import (
	"fmt"
	"net"
	"net/url"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var (
	coreFlag     = "core.ip"
	coreRPCFlag  = "core.rpc.port"
	coreGRPCFlag = "core.grpc.port"
)

// Flags gives a set of hardcoded Core flags.
func Flags() *flag.FlagSet {
	flags := &flag.FlagSet{}

	flags.String(
		coreFlag,
		"",
		"Indicates node to connect to the given core node. "+
			"Example: <ip>, 127.0.0.1. <dns>, subdomain.domain.tld "+
			"Assumes RPC port 26657 and gRPC port 9090 as default unless otherwise specified.",
	)
	flags.String(
		coreRPCFlag,
		"26657",
		"Set a custom RPC port for the core node connection. The --core.ip flag must also be provided.",
	)
	flags.String(
		coreGRPCFlag,
		"9090",
		"Set a custom gRPC port for the core node connection. The --core.ip flag must also be provided.",
	)
	return flags
}

// ParseFlags parses Core flags from the given cmd and saves them to the passed config.
func ParseFlags(cmd *cobra.Command, cfg *Config) error {
	coreIP := cmd.Flag(coreFlag).Value.String()
	if coreIP == "" {
		if cmd.Flag(coreGRPCFlag).Changed || cmd.Flag(coreRPCFlag).Changed {
			return fmt.Errorf("cannot specify RPC/gRPC ports without specifying an IP address for --core.ip")
		}
		return nil
	}

	ip := net.ParseIP(coreIP)
	if ip == nil {
		u, err := url.Parse(coreIP)
		if err != nil {
			return fmt.Errorf("failed to parse url: %w", err)
		}
		ips, err := net.LookupIP(u.Host)
		if err != nil {
			return fmt.Errorf("failed to resolve DNS record: %v", err)
		}
		if len(ips) == 0 {
			return fmt.Errorf("no IP addresses found for DNS record")
		}
		ip = ips[0]
	}

	rpc := cmd.Flag(coreRPCFlag).Value.String()
	grpc := cmd.Flag(coreGRPCFlag).Value.String()

	cfg.IP = ip.String()
	cfg.RPCPort = rpc
	cfg.GRPCPort = grpc
	return nil
}
