# TINYDNS API

A ReST API written in Go to manage TinyDNS zone files.  TinyDNS is fronted by Unbound DNS caching, and the API has methods to manage cache flushing.  The API uses Consul as its persistent data store.

#### Quick Start
 * git clone https://github.com/briankohler/dnsapi
 * cd dnsapi
 * docker-compose up 

Once running, the API listens on localhost:9080 and Unbound listens on localhost:53.  The API only allows inserting records for the domain for which it's authoratative.  The default domain is local.docker.  
Insert an A record:

 ``` curl -XPOST http://localhost:9080/v2/dnsapi/A/host.local.docker/10.10.0.100/300 ```

Now verify the record:

``` dig @127.0.0.1 host.local.docker ```

#### API Methods

```
  get   '/dnsapi/data' - Returns the entire TinyDNS zonefile
  get   '/dnsapi/HEALTH' - Basic health check
  put   '/dnsapi/NS/:key/:value/:ttl/:newvalue/:newttl' - Update an NS record
  put   '/dnsapi/ALIAS/:key/:value/:ttl/:newvalue/:newttl' - Update an ALIAS record
  put   '/dnsapi/A/:key/:value/:ttl/:newvalue/:newttl' - Update an A record
  put   '/dnsapi/CNAME/:key/:value/:ttl/:newvalue/:newttl' - update a CNAME 
  post  '/dnsapi/CNAME/:key/:value/:ttl' - Create a CNAME
  post  '/dnsapi/A/:key/:value/:ttl' - Create an A record
  post  '/dnsapi/ALIAS/:key/:value/:ttl' - Create an ALIAS record
  post  '/dnsapi/NS/:key/:value/:ttl' - Create an NS record
  post  '/dnsapi/NS/:key  - Shorthand to create NS and SOA records
  post  '/dnsapi/SOA/:key/:value  - Create an SOA record
  del   '/dnsapi/:recordtype/:key/:value/:ttl' - Delete records

```

#### Extra features

By setting the environment variable S3_BUCKET on the api container and mounting in your aws credentials file to /root/.aws/credentials, the Tiny datafile will be versioned with Git and backed up to the S3 bucket specified.  

Setting USE_CONSULFS=true on the TinyDNS and API containers, the TinyDNS zonefile will be on a FUSE filesystem, backed by Consul.  THIS IS AN EXPERIMENT AND SHOULD NOT BE DONE EVER!!!!!  Privileged needs to be "true" for this to work.


