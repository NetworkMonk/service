# Service

This is a simple handler that allows a go application to be installed or executed as a service with minimal configuration

## Usage

Import into your app and simply call
```
service.Handle(serviceName, serviceTitle, execute)
```

serviceName - the system name for your service.

serviceTitle - A freindly name for your service, visible in the services.msc console.

execute - this is a reference to a function that is called when the service is started.

## Commands

The following commands can be run from a command prompt / terminal on the executable file.

`debug` - runs the service on the desktop and provides a visible console display. Errors message and other information that is logged is displayed here.

`install` - installs the service onto the system.

`remove` - uninstalls the service from the system.

`start` - starts the service if it is installed.

`stop` - stops any running service.

`pause` - pauses any running service.

`continue` - resumes any paused service.