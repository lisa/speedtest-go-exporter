# Potential Future Roadmap Items

In the unlikely event this project is iterated on, here are some ideas in an unordered list.

* Proper YAML files for applying in a Kubernetes cluster (instead of in the [README.md](/README.md))
* Export geo coordinates for the test server so that it could be graphed on a map
* Support retrying in the event of invalid responses (may need longer timeouts in places)
* Tests of some kind
  * Will require mocking the tests, especially for timeouts
* Figure out a way to share the ca-certs layer across the various arch image builds
* Dockerfile should call Makefile to build the binary (`make build`) and that target should optionally build a static binary
