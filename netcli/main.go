package main

import "flag"

// TODO: Write Command Line Tool

var setZone = flag.String("setZone", "", "Overwrite (permanently) the zone that this machine is in.")
var setDHCPIP = flag.String("setDHCPIP", "", "Overwrite (permanently) the DHCP hosting IP for this machine (or set it to empty to disable DHCP).")
var setDHCPNIC = flag.String("setDHCPNIC", "", "Overwrite (permanently) the DHCP hosting NIC name for this machine (or set it to empty to disable DHCP).")
var setDHCPSubnet = flag.String("setDHCPSubnet", "", "Overwrite (permanently) the DHCP subnet for this zone (requires setZone flag or it'll no-op).")
var setDHCPLeaseDuration = flag.String("setDHCPLeaseDuration", "", "Overwrite (permanently) the default DHCP lease duration for this zone (requires setZone flag or it'll no-op).")
var setDHCPTFTP = flag.String("setDHCPTFTP", "", "Overwrite (permanently) the DHCP TFTP Server Name for this machine (or set it to empty to disable DHCP).")
