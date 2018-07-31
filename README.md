# phly
A tool for running pipeline operations.

A pipeline processing framework. Phly has two main pieces:
* The phly framework that manages registering and instantiating nodes, and loading and running pipelines.
* The phly app, a lightweight application designed to make it easy to compile in additional nodes.

## Building ##
* Install Go 1.8.3 or later.
* From the command line, enter directory `phly\phly`.
* Type `go get` to get all dependencies.
* Type `go build` to build the app.

Alternatively, the phly library can be compiled into other Go apps.

## Use ##
The work so far has been on the framework. The actual application currently does nothing but scale images. To that end, running the app will load the `data/scale_image.json` pipeline, which loads an example image and scales it.

Examples (compiled for Windows):
* `phly.exe`. Run the app, which currently defaults to running the `data/scale_image.json` pipeline.
* `phly.exe nodes`. Display all installed nodes.
* `phly.exe markdown`. Generate markdown for all installed nodes.

## Nodes ##
* **String** (phly/string). Acquire text from the cfg values. If a cla is available use that. If no cla, use the env. If no env, use the value.
    * cfg **value**. A value directly entered into the cfg file. Use this if no cla or env are present.
    * cfg **env**. A value from the environment variables. Use this if no cla is available.
    * cfg **cla**. A value from the command line arguments.
    * output **0**. The single string output.
