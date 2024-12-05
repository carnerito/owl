module owl

go 1.23.4

require github.com/vishvananda/netlink v1.3.0

replace github.com/vishvananda/netlink => github.com/carnerito/netlink v0.0.0-20241205082656-bd0169e1760b

require (
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
