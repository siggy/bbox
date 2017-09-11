**DRAFT**

# Beatboxer: Raspberry Pi, LEDs, and Go

This is the story of building Beatboxer, a human-sized beat machine:

(youtube)

## Prior Art

### Nine Inch Nails

Several years ago I saw Nine Inch Nails perform Echoplex using a beat machine that occupied the entire stage:

<iframe width="560" height="315" src="https://www.youtube.com/embed/6O_92BTrUcA" frameborder="0" allowfullscreen></iframe>

It occurred to me I'd love to try to build something like that.

### Beatboxer in JavaScript

About a year ago I attempted to replicate Nine Inch Nails' beat machine on a web page, using as little code as possible. The result was a JavaScript app called [Beatboxer](https://github.com/siggy/beatboxer):

<iframe width="560" height="315" src="https://sig.gy/beatboxer/" frameborder="0" allowfullscreen></iframe>

This was a fun little project, where I learned a bit about [AudioContext](https://developer.mozilla.org/en-US/docs/Web/API/AudioContext) and [requestAnimationFrame](https://developer.mozilla.org/en-US/docs/Web/API/window/requestAnimationFrame).

## A rewrite in Go

[Go](https://golang.org/) has recently become my favorite programming language, combining the simplicity of C with powerful concurrency primitives. Rewriting Beatboxer in Go was a great excuse to play with the language a bit more.

A bit surprisingly, reading and playing back audio files in Go turned out to be more work than in JavaScript, primarily due to the fact that you're no longer in a browser and need to interface more closely with the hardware. I eventually found a pair of excellent Go libraries, [youpy's go-wav](https://github.com/youpy/go-wav) for reading wavs, and [gordonklaus' Go bindings](https://github.com/gordonklaus/portaudio) for [PortAudio](http://portaudio.com/) to output sound.

## Back to the project

With a working prototype in Go, I returned to thinking about building a physical beat machine. [Burning Man](https://burningman.org/) seemed like a reasonable destination for a project like this, and also provided a concrete deadline for added motivation.

I began brainstorming materials with my good friend [@yet](https://twitter.com/yet). He quickly pointed out that a long, flat beat machine like the one used by Nine Inch Nails would not fare very well in the gale-force winds of the Black Rock Desert. He suggested a wrapping the beat machine into the shape of a phone booth. This seemed much more workable to build and transport.

### Adressable LEDs

As the physical form continued to coalesce, I knew I'd need buttons to enable beats, and LEDs to provide feedback. I began researching addressable LEDs. Like Go, this was another domain where I was looking for a personal project to enable learning more about. Arduino seemed to be the platform of choice for working with LEDs. I was not very familiar with Arduinos or Raspberry Pi, but I did not know a Raspberry Pi is basically just a Linux computer, a more familiar environment for me. I opted to build the project around a Raspberry Pi for this reason, though aware that I was going a bit against the grain for an addressable LED project.

Fortunately the amazing folks at [Adafruit](https://www.adafruit.com/) had a tutorial on [controlling NeoPixel LEDs with a Raspberry Pi](https://learn.adafruit.com/neopixels-on-raspberry-pi). Even more fortuitious, the library used in the tutorial, [jgarff's rpi_ws281x](https://github.com/jgarff/rpi_ws281x), came with a Go wrapper!

I cannot emphasize enough how helpful [Adafruit](https://www.adafruit.com/) is for projects such as these. Their tutorials and documentation are awesome, and they provide all the materials needed to follow along.

Controlling NeoPixels from a Raspberry Pi is not quite as simple as plugging in a USB device. It requires building a circuit board, something I had zero experience with. I again enlisted [@yet](https://twitter.com/yet)'s help to learn how to solder wires, chips, and resistors. In short order we had a Raspberry Pi controlling NeoPixels.

### Keyboard Hacking

With proof of concepts for LEDs and audio playback in place, next was user input. I wanted to use large buttons to enable/disable beats. Most Raspberry Pi tutorials document how to wire up a few buttons, I needed 64. The Raspberry Pi does not have enough pins to handle that many button inputs, so other chips/boards are required. An alternative was to simply hack apart a regular computer keyboard and wire 64 external buttons into it. This was a compelling option, as the software was already built to accept keyboard inputs, no additional coding or GPIO programming required. In hindsight this was probably the worst design decision of the entire project. Though it's a fun hack, wiring 64 buttons into a keyboard PCB yields this:

(image of wires)

https://www.instructables.com/id/Hacking-a-USB-Keyboard/
https://www.instructables.com/id/Create-External-Buttons-For-Your-Keyboard/

dustin keyboard

Jhon pyramid

johanna angles

jacobo plastic

brian support

nathan buttons

