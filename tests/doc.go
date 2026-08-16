// Package tests boots the config plugin inside an Endure container and checks
// what the other plugins receive from it: the address the rpc plugin ends up
// binding, the values that reach a plugin through Unmarshal and Get, command
// line overrides, and environment variable expansion.
package tests
