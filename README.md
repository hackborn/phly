# phly
A tool for running pipeline operations.

WARNING: Currently in development, API will change.

DOUBLE WARNING: This is currently being rewritten, is totally torn up, and close to non-functional.

A pipeline processing framework. Phly has two main pieces:
* The phly framework that manages registering and instantiating nodes, and loading and running pipelines.
* The phly app, a lightweight application designed to make it easy to compile in additional nodes.

## Building ##
* Install Go 1.10.3 or later.
* From the command line, enter directory `phly\phly`.
* Type `go get` to get all dependencies.
* Type `go build` to build the app.

Alternatively, the phly library can be compiled into other Go apps.

## Use ##
The work so far has been on the framework. The actual application currently does nothing but scale images. To that end, running the app will load the `data/scale_image.json` pipeline, which loads an example image and scales it.

Examples (compiled for Windows):
* `phly.exe`. Run the app, which currently defaults to running the `data/scale_image.json` pipeline.
* `phly.exe -nodes`. Display all installed nodes.
* `phly.exe -markdown`. Generate markdown for all installed nodes.
* `phly.exe -vars`. Display all node-defined variables.

## Nodes ##
* **Batch** (phly/batch). Perform multiple actions in parallel.
* **Files** (phly/files). Create file lists from file names and folders. Produce a single doc with a single page.
    * cfg **value**. A value directly entered into the cfg file. Use this if no cla or env are present.
    * cfg **env**. A value from the environment variables. Use this if no cla is available.
    * cfg **cla**. A value from the command line arguments.
    * cfg **expand**. (true or false). When true, folders are expanded to the file contents.
    * output **out**. The file list.
* **Pipeline** (phly/pipeline). Run an internal pipeline.
* **Text** (phly/text). Acquire text from the cfg values. If a cla is available use that. If no cla, use the env. If no env, use the value.
    * cfg **value**. A value directly entered into the cfg file. Use this if no cla or env are present.
    * cfg **env**. A value from the environment variables. Use this if no cla is available.
    * cfg **cla**. A value from the command line arguments.
    * output **out**. The text output.