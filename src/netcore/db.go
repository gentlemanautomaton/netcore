package main

type DB interface {
	ConfigProvider
	DHCPDB
	DNSDB
}
