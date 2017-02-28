# aws-service-lookup

So .. what is this, that it is?  Qu'est-ce que c'est, mon ami?

This is a custom build of [CoreDNS](https://github.com/coredns/coredns) which
has two main goals:

- chopping out stuff I don't need - e.g. that hoofing great big kubernetes client library etc
- adding in its own extra little bit of middleware [(ec2tags)](./ec2tags)

The extra middleware will query an AWS account, currently just for EC2 instances.
Using this DNS server, instances can be resolved as `<name>.<region>.compute.internal`.
Additionally, if the instance has a "Services" tag then you can perform some simple
service-discovery. The "Services" tag on an instance should be a space-separated list
of services provided by that instance, and will be resolvable as `<service>.service.local`

Bit of a work in progress. Not sure if this is useful to anyone else, but it fills a
need for me.

## Building it

Simples.

```sh
make
```

This should do all the vendoring etc, produce a binary, run any tests (ahem, cough)
and run various linters.

The build I do is currently using go1.8, although it's also been built with go1.7.3.
