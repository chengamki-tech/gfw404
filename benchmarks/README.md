# Authorized Performance Benchmarks

These benchmarks measure transport-related CPU and allocation costs. They do
not change packet contents, add camouflage, or attempt to avoid network
inspection. Run them only on systems and interfaces you are authorized to
operate.

Run the local baseline:

```bash
cd benchmarks
GOWORK=off go test -run '^$' -bench . -benchmem
```

For a Linux lab interface, the optional helper applies controlled delay and
loss. Record the interface name and remove the qdisc after the run:

```bash
sudo ./netem.sh apply eth0 80ms 1%
GOWORK=off go test -run '^$' -bench . -benchmem
sudo ./netem.sh clear eth0
```

The helper is intentionally explicit and does not select an interface or
modify the network automatically.
