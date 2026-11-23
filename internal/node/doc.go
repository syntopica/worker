// Package node is the inference node: it samples its host, reports, leases
// one job at a time and runs it on the local model server, releasing it the
// moment the host says so (spec section 7).
package node
