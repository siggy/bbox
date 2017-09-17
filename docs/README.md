---
layout: default
title: "Beatboxer: A human-sized beat machine built with a Raspberry Pi, LEDs, and Go"
permalink: /
---

**DRAFT**

# Beatboxer: A human-sized beat machine built with a Raspberry Pi, LEDs, and Go

This is the story of building Beatboxer:

<iframe width="560" height="315" src="https://www.youtube.com/watch?v=Pbo0mBi7B7w" frameborder="0" allowfullscreen></iframe>

## Prior Art

### Nine Inch Nails

Several years ago I saw Nine Inch Nails perform Echoplex using a beat machine that occupied the entire stage:

<iframe width="560" height="315" src="https://www.youtube.com/embed/6O_92BTrUcA" frameborder="0" allowfullscreen></iframe>

It occurred to me I'd love to try to build something like this.

### Beatboxer in JavaScript

About a year ago I attempted to replicate Nine Inch Nails' beat machine on a web page, using as little code as possible. The result was a JavaScript app called [Beatboxer](https://sig.gy/beatboxer/):

<iframe width="900" height="315" src="/beatboxer/" frameborder="0" allowfullscreen></iframe>

This was a fun little project, where I learned a bit about [AudioContext](https://developer.mozilla.org/en-US/docs/Web/API/AudioContext) and [requestAnimationFrame](https://developer.mozilla.org/en-US/docs/Web/API/window/requestAnimationFrame).

## A rewrite in Go

[Go](https://golang.org/) has recently become my favorite programming language. Rewriting Beatboxer in Go was a great excuse to play with the language a bit more.

The Beatboxer software requires input from a keyboard, which then transmits beat changes to a renderer and a drum loop. The Go language really shines here in allowing explicit definition of these concurrent threads, with communication via Go Channels. This whole architecture is summarized nicely in [BeatBoxer's main](https://github.com/siggy/bbox/blob/master/cmd/bbox.go).

Reading and playing back audio files in Go turned out to be more work than in JavaScript. I eventually found a pair of excellent Go libraries, [youpy's go-wav](https://github.com/youpy/go-wav) for reading wavs, and [gordonklaus' Go bindings](https://github.com/gordonklaus/portaudio) for [PortAudio](http://portaudio.com/) for outputting sound.

I also found myself wanting higher quality drum samples. I found an awesome collection at [99sounds.org](http://99sounds.org/drum-samples/). It's free for download, but definitely throw them a donation if you appreciate their work. I found their 808 beats to be just want I needed. I reached out to confirm they were ok with me using their samples in a public art project, and [Tomislav](https://twitter.com/bpblog) replied a few hours later with full support.

## Back to the project

With a working prototype in Go, I returned to thinking about building a physical beat machine. [Burning Man](https://burningman.org/) seemed like a reasonable destination for a project like this, and provided a concrete deadline for added motivation.

I began brainstorming a structure with my good friend [@yet](https://twitter.com/yet). He quickly pointed out that a long, flat structure like the one used by Nine Inch Nails would not fare very well in the gale-force winds of the Black Rock Desert. He suggested wrapping the beat machine into the shape of a phone booth, as a more compact structure would be much easier to build, transport, and support.

### Addressable LEDs

As the physical form continued to coalesce, I knew I'd need buttons to enable beats, and LEDs to provide feedback. I began researching addressable LEDs. Like Go, this was another domain where I was looking for a personal project to enable learning more about. Arduino seemed to be the platform of choice for working with LEDs. I was not very familiar with Arduinos or Raspberry Pi. I recalled a conversation I had with my friend [Nick](https://github.com/hebnern), who had related to me his experience building a project around Arduino, and the lack of library support being a pain point that would not exist on a Raspberry Pi, since it's basically just a Linux computer. I opted to build the project around a Raspberry Pi for this very reason, though aware that I was going a bit against the grain for an addressable LED project.

Fortunately the amazing folks at [Adafruit](https://www.adafruit.com/) had a tutorial on [controlling NeoPixel LEDs with a Raspberry Pi](https://learn.adafruit.com/neopixels-on-raspberry-pi). Even more fortuitious, the library used in the tutorial, [jgarff's rpi_ws281x](https://github.com/jgarff/rpi_ws281x), came with a Go wrapper!

I cannot emphasize enough how helpful [Adafruit](https://www.adafruit.com/) is for projects such as these. Their tutorials and documentation are awesome, and they provide all the materials needed to follow along.

Controlling [NeoPixels](https://www.adafruit.com/category/168) from a Raspberry Pi is not quite as simple as plugging in a USB device. It requires building a circuit board, something I had zero experience with. I again enlisted [@yet](https://twitter.com/yet)'s help to learn how to solder wires, chips, and resistors. In short order we had a Raspberry Pi controlling [NeoPixels](https://www.adafruit.com/category/168):

(led vid)

### Pulse width modulation and onboard audio

The [rpi_ws281x LED library](https://github.com/jgarff/rpi_ws281x) uses hardware PWM on the Raspberyy Pi to communicate with the LEDs. Unfortunately this [conflicts](https://github.com/jgarff/rpi_ws281x#pwm) with the Pi's onboard sound card. To get around this, I disabled the onboard sound and installed a [Plugable USB Audio Adapter](https://www.amazon.com/gp/product/B00NMXY2MO). For more details on the config changes, have a look at the `external sound card` section of [this repo's README.md](https://github.com/siggy/bbox#env--bootup).

### Keyboard hacking

With proof of concepts for LEDs and audio playback in place, next was user input. I wanted to use large buttons to toggle beats. Most Raspberry Pi tutorials document how to wire up a few buttons, I needed 64. The Raspberry Pi does not have enough pins to handle that many button inputs without other chips or boards. An alternative was to simply hack apart a regular computer keyboard and wire 64 external buttons into it, which I had read about on [Instructables](https://www.instructables.com/) in two articles: [Hacking a USB Keyboard](https://www.instructables.com/id/Hacking-a-USB-Keyboard/) and [Create External Buttons For Your Keyboard](https://www.instructables.com/id/Create-External-Buttons-For-Your-Keyboard/). This was a compelling option, as the software was already built to accept keyboard inputs, no additional coding or GPIO programming required. In hindsight this was probably the worst design decision of the entire project. Though it's a fun hack, wiring 64 buttons into a keyboard PCB yields this:

(image of wires)

Determined to hack this keyboard, I against enlisted much soldering help from [@yet](https://twitter.com/yet) to wire 30 pins from the keyboard pcb to an [Adafruit perma-proto board](https://www.adafruit.com/product/1606). From there we could wire up the 64 combinations of those 30 pins required to support 64 button presses. For those interested in the pinouts of an [AmazonBasics Keyboard](https://www.amazon.com/gp/product/B005EOWBHC/ref=oh_aui_search_detailpage?ie=UTF8&psc=1), have a look at this [code](https://github.com/siggy/bbox/blob/master/bbox/keys.go).

### A change in shape

While all this was going on, I described to my good friend [Jhon](https://www.instagram.com/yesjhon/) what I was working on. Jhon's background in electronics and industrial design exceeds anyone I know, and he immediately suggested I consider a pyramid shape rather than a phone booth. I loved the idea. At the time I did not realize the added complexity a shape like this would require, but in the end it was definitely worth it.

To build a pyramid, I naively thought I could simply design four triangles, do some linear calculations for the angles, and all would fit together. Fortunately my good friend and owner of Three Bears Furniture, [@jneaderhouser](https://twitter.com/jneaderhouser), pointed me to this excellent [compound miter calculator](http://www.pdxtex.com/canoe/compound.htm). This helped me determine that, based on triangles 6 feet tall by 2.5 feet wide, I needed 46.244° side bevels and 77.975° base bevels. I was able to confirm this all worked by building Beatboxer in SketchUp:

(triangle images)

For the full SketchUp files, (click here).

### Plastics

In deciding on the material for the exterior of the pyramid, my friends [@yet](https://twitter.com/yet) and Brian suggested plywood, but a desire for transparent material, to allow LEDs to shine through from the inside, led me to plastic. In the end we built the base and a halo out of plywood, as an homage to this initial design, and our shared fondness of [Tom Sachs' Love Letter to Plywood](https://vimeo.com/44947985).

I had no experience working with or manufacturing plastic. Fortunately, the good folks at [TAP Plastics SF](https://www.tapplastics.com/about/locations/detail/san_francisco_ca), and specifically my new friend at TAP, Jacobo, were extremely helpful in guiding me. I submitted the overall triangle dimensions for C&C manufacturing. The trickier part was producing those precise bevel angles so the triangles would fit together perfectly, and align with the ground. Jacobo took it upon himself to cut the bevels by hand. I had ordered up five triangles instead of four in case anything broke. I ended up with five perfectly shaped triangles. When put together, the four tops of the triangles fit with absolute precision, I cannot thank Jacobo and TAP enough for their work on this:

(jacobo photo)

### Construction

With software, electronics, and materials proven out, last was actual assembly. As with all the collaboration on this project, I simultaneously realized I needed help and found a friend extremely willing to collaborate. My friend Brian, who could build a house if needed, and had experience constructing things to withstand the elements at Burning Man, understood well before I did what was required to assemble Beatboxer. I showed up at his house expecting to ask some questions, and he had a complete internal skeleton designed, with materials sourced and ready to assemble. In one evening he built a complete internal structure out of PVC pipes, very strong and lightweight:

(PVC Photo)

Without Brian's help I would have assembled this thing with glue and tape and it probably would have collapsed on day one.

### Last minute freakout

While helping me with bevel angles, my woodworker friend [@jneaderhouser](https://twitter.com/jneaderhouser) also advised me to build a prototype. I forewent this step in favor of SketchUp, concerned I did not have time for additional construction. One week before the event this came back to bite me in a big way. I had purchased 64 [red arcade buttons](https://www.amazon.com/gp/product/B00V0OM7WO/ref=oh_aui_search_detailpage?ie=UTF8&psc=1) to be mounted in the pyramid sides, which would control the beats. As I assembled Beatboxer for the first time, I quickly realized that the buttons were so deep that they collided on the corners of the pyramid. I needed shallower buttons, but the holes in the pyramid had been cut specifically for the diameter of these larger buttons. Again a friend, [@oceanphoto](https://twitter.com/oceanphoto), swooped in to save the day. He quickly 3D printed mountings to allow smaller buttons to fit tightly in the pyramid holes:

(3d mounting photo)

## The big event

With all the pieces in place, my good friends Rob and [@cecbayan](https://www.instagram.com/cecbayan/) provided some space on their trailer next to their insane riding-lawnmower-turned-tank, [The Krawler](https://www.facebook.com/thekrawler/), to transport Beatboxer from the Bay Area to the Black Rock Desert.

Upon arrival setup went surprisingly smooth. The awesome folks at [The Artery](https://burningman.org/event/art-performance/artist-resources/) guided us to our placement location and let us do the honors of putting a stake in the ground:

(stake photo)

After a couple days of testing Beatboxer in our camp, we loaded it onto Brian's trailer and drove it out to its final location:

(trailer driving video)

The rest of the week was spent swapping and charging batteries and speakers, which proved a bit more challenging due to Beatboxer getting a lot more use than anticipated (an exciting problem to have). In the end it was fully functional roughly 90% of the week, with outages only due to batteries. The structure, buttons, and electronics held up to the elements, including a Mad Max-style dust and rain storm:

(photo of dust storm)

## Lessons learned

I originally set out to build Beatboxer on my own, partly to see if I could, but primarily because I did not want to burden friends who already had amazing projects on their plates. Though I quickly found that I was in way over my head, I fortunately also found my friends were extremely willing to get involved. Without everyone's help, there's no way this project would have happened. I cannot thank everyone enough.

## Epilogue

As I learned to work with LEDs, I found opportunities to work with other artists to help light their own projects. While Beatboxer used about 300 LEDs, these projects demanded a bit more. Applying the same powering methods to these larger projects resulted in some melted electronics, necessitating a power redesign. Fortunately my friend [Tom](http://www.tinsel.org/) quickly provided a wiring diagram to overcome these power issues:

(tom's wiring diagram)

With additional components and design complexity, my friend [@oceanphoto](https://twitter.com/oceanphoto) jumped back in and 3D printed some awesome enclosures to keep everything safe from the elements, and even included ethernet ports to be able to update the Raspberry Pi's without opening the enclosures:

(enclosure photo)

Another requirement for these projects was faster boot time. With vanilla Raspian I was seeing 30 second boot times before the LED programs would start. This was fine for Beatboxer where it was left running all night, but much too long for a mutant vehicle that should be illuminated as soon as possible when it's turned on. I switched to Raspbian Lite and disabled as many services as I could, eventually getting boot time down to around 7 seconds. For full details on these config changes, have a look at the [First Boot section of this repo's README.md](https://github.com/siggy/bbox#first-boot).

With power and boot time sorted, we were ready to apply these LEDs to three more projects.

### The Krawler

[The Krawler](https://www.facebook.com/thekrawler/) is Rob and [@cecbayan](https://www.instagram.com/cecbayan/)'s riding-lawnmower-turned-tank. It also shoots fire from eight poofers mounted on top. Here we added 480 LEDs along the bottom frame in a streak-like pattern, to highlight some of the amazing manufacturing and engineering that went into this vehicle.

(krawler pic/vid)

### Crane of Remembrance

The Crane of Remembrance is a 15-foot tall replica of a Port of Oakland crane, and a tribute to the victims of the [Ghost Ship Fire](https://en.wikipedia.org/wiki/2016_Oakland_warehouse_fire). Our friend Rob is an Oakland firefighter, and was onsite during the incident. He inspired our group to build the Crane. The lighting included 540 LEDs, 240 in a beating heart, and 300 in two strands extending across the top. The heart beats at 36 BPM, and the strands display 36 lights at a time, one for each person lost in the fire.

(crane pic/vid)

### The Fish

The Fish is Brian's mutant vehicle, built on top of an airport luggage carrier. This year was the Fish's tenth year at Burning Man. This turned out to be the most ambitious lighting project, involving 2040 LEDs, all mounted in its four fins. Each side fin alone has 720 LEDs, and the top and back fins have a combined 600. Painting with LEDs on a canvas like this turned out to be quite a thrill:

(fish pic/vid)
