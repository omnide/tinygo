// Code generated by wit-bindgen-go. DO NOT EDIT.

//go:build !wasip1

// Package ipnamelookup represents the interface "wasi:sockets/ip-name-lookup@0.2.0".
package ipnamelookup

import (
	"github.com/ydnar/wasm-tools-go/cm"
	"internal/wasi/io/v0.2.0/poll"
	"internal/wasi/sockets/v0.2.0/network"
)

// ErrorCode represents the enum "wasi:sockets/network@0.2.0#error-code".
//
// See [network.ErrorCode] for more information.
type ErrorCode = network.ErrorCode

// IPAddress represents the variant "wasi:sockets/network@0.2.0#ip-address".
//
// See [network.IPAddress] for more information.
type IPAddress = network.IPAddress

// Network represents the resource "wasi:sockets/network@0.2.0#network".
//
// See [network.Network] for more information.
type Network = network.Network

// Pollable represents the resource "wasi:io/poll@0.2.0#pollable".
//
// See [poll.Pollable] for more information.
type Pollable = poll.Pollable

// ResolveAddressStream represents the resource "wasi:sockets/ip-name-lookup@0.2.0#resolve-address-stream".
//
//	resource resolve-address-stream
type ResolveAddressStream cm.Resource

// ResourceDrop represents the Canonical ABI function "resource-drop".
//
// Drops a resource handle.
//
//go:nosplit
func (self ResolveAddressStream) ResourceDrop() {
	self.wasmimport_ResourceDrop()
}

//go:wasmimport wasi:sockets/ip-name-lookup@0.2.0 [resource-drop]resolve-address-stream
//go:noescape
func (self ResolveAddressStream) wasmimport_ResourceDrop()

// ResolveNextAddress represents method "resolve-next-address".
//
// Returns the next address from the resolver.
//
// This function should be called multiple times. On each call, it will
// return the next address in connection order preference. If all
// addresses have been exhausted, this function returns `none`.
//
// This function never returns IPv4-mapped IPv6 addresses.
//
// # Typical errors
// - `name-unresolvable`:          Name does not exist or has no suitable associated
// IP addresses. (EAI_NONAME, EAI_NODATA, EAI_ADDRFAMILY)
// - `temporary-resolver-failure`: A temporary failure in name resolution occurred.
// (EAI_AGAIN)
// - `permanent-resolver-failure`: A permanent failure in name resolution occurred.
// (EAI_FAIL)
// - `would-block`:                A result is not available yet. (EWOULDBLOCK, EAGAIN)
//
//	resolve-next-address: func() -> result<option<ip-address>, error-code>
//
//go:nosplit
func (self ResolveAddressStream) ResolveNextAddress() cm.OKResult[cm.Option[IPAddress], ErrorCode] {
	var result cm.OKResult[cm.Option[IPAddress], ErrorCode]
	self.wasmimport_ResolveNextAddress(&result)
	return result
}

//go:wasmimport wasi:sockets/ip-name-lookup@0.2.0 [method]resolve-address-stream.resolve-next-address
//go:noescape
func (self ResolveAddressStream) wasmimport_ResolveNextAddress(result *cm.OKResult[cm.Option[IPAddress], ErrorCode])

// Subscribe represents method "subscribe".
//
// Create a `pollable` which will resolve once the stream is ready for I/O.
//
// Note: this function is here for WASI Preview2 only.
// It's planned to be removed when `future` is natively supported in Preview3.
//
//	subscribe: func() -> pollable
//
//go:nosplit
func (self ResolveAddressStream) Subscribe() Pollable {
	return self.wasmimport_Subscribe()
}

//go:wasmimport wasi:sockets/ip-name-lookup@0.2.0 [method]resolve-address-stream.subscribe
//go:noescape
func (self ResolveAddressStream) wasmimport_Subscribe() Pollable

// ResolveAddresses represents function "wasi:sockets/ip-name-lookup@0.2.0#resolve-addresses".
//
// Resolve an internet host name to a list of IP addresses.
//
// Unicode domain names are automatically converted to ASCII using IDNA encoding.
// If the input is an IP address string, the address is parsed and returned
// as-is without making any external requests.
//
// See the wasi-socket proposal README.md for a comparison with getaddrinfo.
//
// This function never blocks. It either immediately fails or immediately
// returns successfully with a `resolve-address-stream` that can be used
// to (asynchronously) fetch the results.
//
// # Typical errors
// - `invalid-argument`: `name` is a syntactically invalid domain name or IP address.
//
// # References:
// - <https://pubs.opengroup.org/onlinepubs/9699919799/functions/getaddrinfo.html>
// - <https://man7.org/linux/man-pages/man3/getaddrinfo.3.html>
// - <https://learn.microsoft.com/en-us/windows/win32/api/ws2tcpip/nf-ws2tcpip-getaddrinfo>
// - <https://man.freebsd.org/cgi/man.cgi?query=getaddrinfo&sektion=3>
//
//	resolve-addresses: func(network: borrow<network>, name: string) -> result<resolve-address-stream,
//	error-code>
//
//go:nosplit
func ResolveAddresses(network_ Network, name string) cm.OKResult[ResolveAddressStream, ErrorCode] {
	var result cm.OKResult[ResolveAddressStream, ErrorCode]
	wasmimport_ResolveAddresses(network_, name, &result)
	return result
}

//go:wasmimport wasi:sockets/ip-name-lookup@0.2.0 resolve-addresses
//go:noescape
func wasmimport_ResolveAddresses(network_ Network, name string, result *cm.OKResult[ResolveAddressStream, ErrorCode])
