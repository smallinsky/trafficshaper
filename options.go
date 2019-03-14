package trafficshaper

type Opt func(*options)

func WithServerBandwith(limit Limit) Opt {
	return func(o *options) {
		o.serverLimit = &limit
	}
}

func WithCIDRBandiwth(cidr string, limit Limit) Opt {
	return func(o *options) {
		o.cidrs = append(o.cidrs, cidrLimit{
			cidr:  cidr,
			limit: limit,
		})
	}
}

type cidrLimit struct {
	cidr  string
	limit Limit
}

type Limit struct {
	UL int64
	DL int64
}

type options struct {
	cidrs       []cidrLimit
	serverLimit *Limit
}
