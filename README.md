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

There's currently no health-checking involved. This is most useful if you don't want
to run extra infrastructure components for a simple deployment where you have just a few
which are not redundantly-provisioned.

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

You can also build an RPM package which has some config files and init scripts.
The init scripts should support upstart (for EL6) and systemd (for EL7), although
they've mostly been used on EL6 so far.

```sh
make rpm
```

## Configuration

If you're running this on AWS, the easiest thing to do is ensure that you have
an EC2 instance profile which gives you readonly access to the EC2 API. If that's
in place then the accesskey/secret/token will be read from the instance metadata
and you don't need to do anything else.

Otherwise, you'll need to pass in am AWS access key/secret, which can be done
either on the command-line or via the usual environment variables.

## Running it

There's fairly decent help available on the command line.

```sh
aws-service-lookup --help
```
