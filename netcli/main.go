package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// TODO: Write Command Line Tool

var setZone = flag.String("setZone", "", "Overwrite (permanently) the zone that this machine is in.")
var setDHCPIP = flag.String("setDHCPIP", "", "Overwrite (permanently) the DHCP hosting IP for this machine (or set it to empty to disable DHCP).")
var setDHCPNIC = flag.String("setDHCPNIC", "", "Overwrite (permanently) the DHCP hosting NIC name for this machine (or set it to empty to disable DHCP).")
var setDHCPSubnet = flag.String("setDHCPSubnet", "", "Overwrite (permanently) the DHCP subnet for this zone (requires setZone flag or it'll no-op).")
var setDHCPLeaseDuration = flag.String("setDHCPLeaseDuration", "", "Overwrite (permanently) the default DHCP lease duration for this zone (requires setZone flag or it'll no-op).")
var setDHCPTFTP = flag.String("setDHCPTFTP", "", "Overwrite (permanently) the DHCP TFTP Server Name for this machine (or set it to empty to disable DHCP).")

var (
	commandSet = []string{"init"}
)

func usage(commands []string, fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])

	if commands != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("  %s", strings.Join(commands, "\n  ")))
	}

	if fs != nil {
		fs.PrintDefaults()
	}
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.Parse(os.Args)

	if len(os.Args) < 2 {
		// TODO: Display help
		usage(commandSet, fs)
		os.Exit(2)
	}

	subject := strings.ToLower(os.Args[0])
	args := os.Args[1:]
	switch subject {
	case "init", "initialize":
		initCommand(args)
	case "inst", "instance":
		instanceCommand(args)
	case "net", "network":
		networkCommand(args)
	case "prefix":
		prefixCommand(args)
	case "device":
		deviceCommand(args)
	case "type":
		typeCommand(args)
	case "pool":
		poolCommand(args)
	default:
		// TODO: Display Help
	}
}

func initCommand(args []string) {

}

func prefixCommand(args []string) {

}

func deviceCommand(args []string) {

}

func typeCommand(args []string) {

}

func poolCommand(args []string) {

}

func instanceCommand(args []string) {
	if len(args) < 1 {
		// TODO: Print help
		os.Exit(2)
	}

	subject := strings.ToLower(args[0])

	switch subject {
	case "pool":
	}
}

func networkCommand(args []string) {
	if len(args) < 1 {
		// TODO: Print help
		os.Exit(2)
	}

	subject := strings.ToLower(args[0])
	args = args[1:]

	// TODO: Parse flags
	switch subject {
	case "prefix":
		networkPrefixCommand(args)
	case "device":
		networkDeviceCommand(args)
	case "type":
		networkTypeCommand(args)
	case "pool":
		networkPoolCommand(args)
	case "create", "add":
	case "rm", "del", "remove", "delete":
	case "list":
	default:
		// TODO: Print help
		os.Exit(2)
	}
}

func networkPrefixCommand(args []string) {

}

func networkDeviceCommand(args []string) {

}

func networkTypeCommand(args []string) {

}

func networkPoolCommand(args []string) {
	if len(args) < 1 {
		// TODO: Print help
		os.Exit(2)
	}

	verb := strings.ToLower(args[0])

	switch verb {
	case "add":
		// Add network pool
	case "rm", "del", "remove", "delete":
		// Remove network pool
	case "list":
		// List network pools
	}
}
