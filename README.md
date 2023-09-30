# SDS Configuration

**This module provides packages to configure the service in the *dev* context.**

The SDS framework doesn't provide a configuration for the entire application.
For decentralization, there is no shared data by all services. 
That includes the configurations as well.

The SDS services created with [service-lib](https://github.com/ahmetson/service-lib), can access the command line arguments only.

Everything else must be retrieved from the configuration.
Even the environment variables too.

The configuration is the key-value database.
This database is running on a separated thread. 
As such, it's wrapped by the [handler](https://github.com/ahmetson/handler-lib).

> The *handler* code is defined in the `handler` package.

To access the configuration, other threads must use a client.
The *client* code is defined in the `client` package.

The configuration data could be grouped into two categories.
The first group is [custom](#custom-data) data.
The second group is the [service](#service-meta) configuration itself.

> **todo** 
>
> In the API reference group the routes by custom data and service meta

## Custom Data
There is only one way to set the custom parameters to use in the service.
Only by providing an environment variables or `.env` files.

For example, `PRIVATE_KEY=0xdead` environment variable will be available by `PRIVATE_KEY` key.  

It's possible to pass the group of the environment variables as the `.env` files.
The `.env` in the same directory as the binary is loaded automatically.
To pass the other `.env` files pass them as the command line arguments.

Example of passing environment files

```shell
/bin/sds-app --flag --flag2=value ./dev.env ./env_file "C:/Program Files/shared/app"
```

### Engine
To turn the environment variables into the configuration parameters, this module uses [spf13/viper](https://github.com/spf13/viper).
It's defined in the `engine` package.

> Contributing
> 
> To include the configuration from .ini, .toml or .json edit the `engine`.

## Service meta
The configuration is also responsible for generation, storage of the service parameters.
The service parameters include the meta parameters such as a list of the handlers and their exposed port.

The services in the distributed systems must know the other services as well.
In SDS Framework, the services nearby services parameters only.

Nearby services are the extensions and proxies.

The following diagram shows the configuration:

![User and Handler diagram](_assets/ServiceConfiguration.jpg "Handler diagram")
*The green box is a service with the proxies and extensions.*
*The white rectangular is the part stored in the configuration for the service*

### Proxy relationship
If there is a proxy chain, then the service knows the information about the last proxy.
All front proxies are unknown for the service.

### Extension relationship
If there is an extension, then the service has access to the front of the extension.
Extension maybe behind the proxies. In this case, the service has access to the first proxy.
Everything after the proxy is unknown.
