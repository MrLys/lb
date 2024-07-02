# A very simple loadbalancer that should not be used in production
It has just been created for educational purpuses and I cannot guarantee that it will work

## How to use it
* Add a list of hosts to the `./conf/conf.yml` file.
* Create an endpoint `/health` (specified in the conf file) on your application servers that so the load balancer can check if the server is healthy.
* build with `go build`
* run with `./lb`
* That should be all you need.

### Misc
* you can test it with the included test server found in /test/main.go


