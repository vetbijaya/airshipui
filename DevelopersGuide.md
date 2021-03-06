# Airship UI Developer's Guide

## Prerequisites

1. Airship UI needs to be pointed to a Kubernetes Cluster. For development we recommending [setting up Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
2. Install [Go](https://golang.org/dl/) v1.13 or newer
3. Install [Octant](https://github.com/vmware-tanzu/octant)

## Getting Started

Let's clone the Airship UI repository and build

    git clone https://opendev.org/airship/airshipui
    cd airshipui
    make build install-plugins

Now that Airship is built and we have a binary we can run it

    ./bin/airshipui

## Airship UI Plugins

- [Argo UI](./cmd/argoui/README.md)

## Developing Airship UI alongside Octant

### Running a local Octant binary

Instead of running your installed version of Octant you may want to run your development version that you built.
To do so you can append the build directory to your path.

Example (Adjust based on where your Octant repository is located):

    PATH=$PATH:/home/jordan/go/src/github.com/vmware-tanzu/octant/build ./bin/airshipui

### Using local Octant development

You may also want your Go code to depend on your local development version of Octant. To do that you can replace the Go dependency with your local version.

Example (Adjust based on where your Octant repository is located):

    go mod edit -replace github.com/vmware-tanzu/octant=/home/jordan/go/src/github.com/vmware-tanzu/octant

Don't forget to remove the replace before doing a `git review`

## Running on a remote host

If you're running Airship UI on a remote host and unable to open a browser on the machine you can tunnel the connection over ssh and use a browser on your local machine

    ssh -L 7777:localhost:7777 -L 8001:localhost:8001 <id>@<remote_host>

To avoid Airship UI from attempting to open the browser on startup run

    airshipui --disable-open-browser

## Working with the Airship UI

Airship UI utilizes the Octant plugin to drive some of the functionality.  Octant has a [plugin sample](https://github.com/vmware-tanzu/octant/blob/master/cmd/octant-sample-plugin/main.go) which is the basis of these instructions.

### Write a hello world plugin of your very own

1. Clone the Airship UI repository (if not already done previously)
    ```
    git clone https://opendev.org/airship/airshipui
    ```
2. cd airshipui
3. Make a hello-world package under the cmd sub folder to contain your code
    ```
    mkdir cmd/hello-world
    ```
4. Create / edit main.go under the cmd/hello-world directory.  This is modeled off the [Octant sample plugin](https://github.com/vmware-tanzu/octant/blob/master/cmd/octant-sample-plugin/main.go)
    ```
	package main

	import (
		"log"

		"github.com/vmware-tanzu/octant/pkg/navigation"
		"github.com/vmware-tanzu/octant/pkg/plugin"
		"github.com/vmware-tanzu/octant/pkg/plugin/service"
		"github.com/vmware-tanzu/octant/pkg/view/component"
	)

	var pluginName = "hello-world"

	// HelloWorldPlugin is a required struct to be an octant compliant plugin
	type HelloWorldPlugin struct{}

	// return a new hello world struct
	func newHelloWorldPlugin() *HelloWorldPlugin {
		return &HelloWorldPlugin{}
	}

	// This is a sample plugin showing the features of Octant's plugin API.
	func main() {
		// Remove the prefix from the go logger since Octant will print logs with timestamps.
		log.SetPrefix("")

		// Tell Octant to call this plugin when printing configuration or tabs for Pods
		capabilities := &plugin.Capabilities{
			IsModule: true,
		}

		hwp := newHelloWorldPlugin()

		// Set up what should happen when Octant calls this plugin.
		options := []service.PluginOption{
			service.WithNavigation(hwp.handleNavigation, hwp.initRoutes),
		}

		// Use the plugin service helper to register this plugin.
		p, err := service.Register(pluginName, "The very smallest thing you can do", capabilities, options...)
		if err != nil {
			log.Fatal(err)
		}

		// The plugin can log and the log messages will show up in Octant.
		log.Printf("hello-world-plugin is starting")
		p.Serve()
	}

	// handles the navigation pane interation
	func (hwp *HelloWorldPlugin) handleNavigation(request *service.NavigationRequest) (navigation.Navigation, error) {
		return navigation.Navigation{
			Title:    "Hello World Plugin",
			Path:     request.GeneratePath(),
			IconName: "folder",
		}, nil
	}

	// initRoutes routes for this plugin. In this example, there is a global catch all route
	// that will return the content for every single path.
	func (hwp *HelloWorldPlugin) initRoutes(router *service.Router) {
		router.HandleFunc("*", hwp.routeHandler)
	}

	func (hwp *HelloWorldPlugin) routeHandler(request service.Request) (component.ContentResponse, error) {
		contentResponse := component.NewContentResponse(component.TitleFromString("Hello World Title"))
		helloWorld := component.NewText("Hello World just some text on the page")
		contentResponse.Add(helloWorld)
		return *contentResponse, nil
	}
    ```
5. Build the hello-world plugin

    Build the single plugin
    ```
    go build -o ~/.config/octant/plugins/hello-world-plugin cmd/hello-world/main.go
    ```

    Or build all by issuing a make command.  Example output:
    ```
    $ make
    go build -o bin/airshipui -ldflags='-X opendev.org/airship/airshipui/internal/environment.version=7274339' cmd/airshipui/main.go
    go build -o bin/hello-world -ldflags='-X opendev.org/airship/airshipui/internal/environment.version=7274339' cmd/hello-world/main.go
    go build -o bin/argoui -ldflags='-X opendev.org/airship/airshipui/internal/environment.version=7274339' cmd/argoui/main.go
    ```
6. Restart the airshipgui with the -v flag so you can see the debug output (~/airshipui/bin/airshipui -v).  Open your browser to http://127.0.0.1:7777/#/configuration/plugins to make sure the hello-world plugin showed up.
    You will see example of the debug message added to the initRoutes function in the log:
    ```
    2020-01-15T19:00:11.530Z        DEBUG   hello-world-plugin      plugin/logger.go:33     2020/01/15 19:00:11 Sample debugging message: Hello World just some text on the page
    ```
7. You should be able to explore the basic plugin functionality by clicking on it in the navigation pane.
8. You can add other components for fun from the [Octant documentation](https://octant.dev/docs/master/plugins/reference/#component-text) for further enhancement to the page

## Appendix

### Minikube

[Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/) runs a single-node Kubernetes cluster for users looking to try out Kubernetes or develop with it day-to-day.  Installation instructions are available on the kubernetes website: https://kubernetes.io/docs/tasks/tools/install-minikube/).  If you are running behind a proxy it may be necessary to follow the steps outlined in the [How to use an HTTP/HTTPS proxy with minikube](https://minikube.sigs.k8s.io/docs/reference/networking/proxy/) website.

### Optional proxy settings

#### Environment settings for wget or curl

If your network has a proxy that prevents successful curls or wgets you may need to set the proxy environment variables.  The local ip is included in the no_proxy setting to prevent any local running process that may attempt api calls against it from being sent through the proxy for the request:

    ```
    export http_proxy=<proxy_host>:<proxy_port>
    export HTTP_PROXY=<proxy_host>:<proxy_port>
    export https_proxy=<proxy_host>:<proxy_port>
    export HTTPS_PROXY=<proxy_host>:<proxy_port>
    export no_proxy=localhost,127.0.0.1,<LOCAL_IP>
    export NO_PROXY=localhost,127.0.0.1,<LOCAL_IP>
    ```
