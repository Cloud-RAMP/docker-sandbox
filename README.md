# Cloud RAMP Docker Sandbox

The purpose of this package is to evaluate the performance of a Docker-based user code sandbox instead of WebAssembly.

### System Architecture

* Upon initialization, we will not start any Docker containers
  * This is because we don't know which services will have requests, so we can't pre-warm functions
* Upon request, we will:
  * Fetch the user's code from external store (for testing purposes, the local fs)
  * Create a new container to run the service
  * We will support only a certain number of containers on one machine, to mimic the same behavior as our WASM system. However, we will likely be able to support a fewer number of containers than WASM modules, due to their size
  * Execute the function and return a response
    * How do we do communication? Custom protocol defined over unix domain sockets
* Benchmark this process
  * One benchmark where we don't hit the container limit, no swapping necessary
  * One benchmark where we have one over the limit, swapping is necessary every n requests
  * For our tests, we will run the same on both systems:
    * We will call the onMessage method for both, having defined an onMessage within the user code that calls an external function and waits for a response
    * Knowing that the functionality is fully complete in the WASM version, we can simulate this in the Docker version by simply sending a payload back and forth between container / coordinator
* Main issue
  * Need to dynamically execute user code. Because we know it's safe, we could just use the `eval` function to parse it into functions and execute those


### Test Architecture

* Define some TypeScript code we will run in response to the conventional WebSocket events
  * onJoin
  * onLeave
  * onMessage
  * onError
* Define a set of instances whose code we will run, likely identical but with different names so we can differentiate between them
* Set up a pool of Docker containers, where we have one per instance of code running
  * We could opt for a scheduler based approach where code is dynamically executed on a predefined size set of containers, but that would require the use of a scheduler. We don't use a scheduler for WASM, so no need to use one here
* Execute a benchmarked test where we create a zpif distribution of which modules receive requests

### Communicaiton Protocol

We will use a very basic protocol to communicate between the two.
* First byte will denote what the incoming message represents
  * 0 = sending initial code that the container needs to execute
  * 1 = sending request (coordinator -> container)
  * 2 = sending resposne (container -> coordinator)
  * 3 = sending request (container -> coordinator)
  * 4 = sending response (coordinator -> container)
  * 5 = sending error (either way)
  * 6 = container done processing (container -> coordinator)