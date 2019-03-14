# Trafficshaper

Library that allows shape incomming and outgoing traffic per server or CIDR. Due to preformence penalty for production usage I would recomend
`tc` library from iproute2 package that allows creates rules for kernel packet scheduler.

## Example Usage

```
import (
	"net/http"

	ts "github.com/smallinsky/trafficshaper"
)

func main() {
	opts := []ts.Opt{
		ts.WithServerBandwith(ts.Limit{UL: ts.MBps * 10, DL: ts.MBps * 10}),
		ts.WithCIDRBandiwth("192.168.0.1/26", ts.Limit{UL: ts.MBps * 5, DL: ts.MBps * 10}),
		ts.WithCIDRBandiwth("192.168.1.1/26", ts.Limit{UL: ts.KBps * 64, DL: ts.KBps * 64}),
	}

	http.Handle("/", handler)
	l, e := ts.NewListener("tcp", ":1234", opts...)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	if err := http.Serve(l, nil); err != nil {
		log.Fatal(err)
	}
}
```

Server limit can be altered by calling `AddServerLimit` or `AddCIDRLimit` functions at runtime.

