# The Zeta Machine

![Zeta Machine Bulb](http://zeta-machine.chriscowherd.com/public/bulb.png)

See it here: [http://zeta-machine.chriscowherd.com/](http://zeta-machine.chriscowherd.com/)

The [Zeta Machine](http://zeta-machine.chriscowherd.com) generates fractal-like 
iteration imagery of the [Riemann Zeta function](https://en.wikipedia.org/wiki/Riemann_zeta_function) 
using a distributed system of GPUs across cloud resources. 

![Riemann Zeta Function](http://zeta-machine.chriscowherd.com/public/riemann-zeta-function-sm.png)

The infrastructure and plumbing are all written in Go and 
the rendering is done in Cuda if you are running on an Cuda-enabled NVidia GPU. 
If you are not using a Cuda-enabled GPU, the algorithms are also written in Go 
and are rendered in software.

This project is based on the work of David Rainford's original Zeta Machine. His 
program was written in C and rendered a single image at a time at a specific
zoom level per pixel.

The Zeta Machine defines a patch to render based on starting and ending coordinates.
Based on these coordinates and the size of the patch, the algorithm iterates the 
Riemann zeta function at a given pixel until the value either blows-up or it 
converges on a fixed point within some tolerance. The number of iterations become 
an index into a color map that determines the color of the resulting pixel.

Patches are defined and dispatched to a message queue for the rendering farm to
pick up, render and place the iteration data back onto the queue for decoding and 
storage.

By rendering patches at different zoom levels we can then turn the whole thing into
a map that can be scrolled and zoomed.

Everything is hosted as a static website at [http://zeta-machine.chriscowherd.com](https://zeta-machine.chriscowherd.com)

## Building & Running

There are three programs that go into generating the tiled images. Once all the
tiles are generated, you can serve them up statically from an S3 bucket.

To run it locally, make a copy of the `.env-SAMPLE` file and call it `.env`. 
Have a look at the values in the file and adjust accordingly. You can ignore the
ones for the web server unless you want to go down that route.

Each can be built simply:

> go build -o build/request ./cmd/request

> go build -o build/generate ./cmd/generate

> go build -o build/store ./cmd/store

You could also just `go run` them.

### Request
The Request service (`zeta-machine/cmd/request`) will generate and publish messages
to the message queue (NSQ). The messages will contain the coordinates, zoom level and
other data to generate tile patches. You can specify the starting and ending zoom levels
as well as whether to generate tiles only for the bulb area.

### Generate
The Generate service (`zeta-machine/cmd/generate`) can be compiled to use an NVidia
GPU along with Cuda to very quickly render tiles. (see `pkg/zeta/cuda.go` comments
for build command) Building by default with no flags will just run on your CPU. If
you have multiple CPUs + cores it will divide the rendering work up over all of them.

Once generated, the data is sent back to the message queue for storage.

### Store
The Store service (`cmd/store`) pulls generated tile data from the message queue,
encodes it into a PNG and stores it to disk.

### Other
There are some other commands such as **lambda**, **seed** and **web** that aren't
actually used and were for some experiments.

**Lambda** - this was an experiment in having tiles generating as an AWS Lambda job.
In the end they weren't fast enough or cost effective so I went down another route.
I left this for your reference or entertainment.

**Seed** - this generates tiles in-process without the complexity message queueing
for testing and other experiments.

**Web** - this was an experiement to have tiles generated on the fly by a server 
farm when they were requested by the web server. It works but the tiles take a 
while to generate and I didn't want to have a farm large enough (spend the money) 
to have it near real time. Maybe in the future.

