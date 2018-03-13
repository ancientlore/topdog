# topdog

`topdog` is a simple Go application that is used to demonstrate some [Istio] features. See the full example at https://github.com/ancientlore/istio-talk.

`topdog` is designed to be run with three tiers. To build it locally:

    $ got get github.com/ancientlore/topdog

Run the application from the `topdog` folder where you built it. To start the backend tier:

    $ ./topdog -service_port 5002

To start the middle tier:

    $ ./topdog -service_port 5001 -backend http://localhost:5002

To start the UI:

    $ ./topdog -service_port 5000 -midtier http://localhost:5001

Then nagivate to http://localhost:5000/ to see the user interface.

Alternately, you can run it all in one step using:

    $ ./topdog -service_port 5000 -midtier http://localhost:5000 -backend http://localhost:5000

> This is the same as just running `topdog`, since those values are the defaults.

In this case, it will use the same process for all three.

When running the backend, you can set the `version` command-line argument (or the `VERSION` environment variable) to values from 1 to 3. This makes the service weigh its results differently.

