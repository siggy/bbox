---
layout: default
title: "Beatboxer: A human-sized drum machine built with a Raspberry Pi, a Feather SCORPIO, NeoPixels, and Go"
permalink: /
---

<style>
  :root { color-scheme: dark; }
  html, body, h1, h2 { background:#202124; color:#619e4e; }
  a { color:#b88de3; }
  a:visited { color:#a774d6; }
  pre, code { background:#111; color:#619e4e; }
  hr { border-color:#333; }
</style>

# Beatboxer: A human-sized drum machine built with a Raspberry Pi, a Feather SCORPIO, NeoPixels, and Go

<iframe width="560" height="315" src="https://www.youtube.com/embed/MepepCV4EUw" frameborder="0" allowfullscreen></iframe>

# Prior Art

Inspired by Nine Inch Nails' Echoplex drum machine:

<iframe width="560" height="315" src="https://www.youtube.com/embed/6O_92BTrUcA" frameborder="0" allowfullscreen></iframe>
---
First implementation in JavaScript at [sig.gy/beatboxer](https://sig.gy/beatboxer):

<iframe width="900" height="315" src="https://sig.gy/beatboxer" frameborder="0" allowfullscreen></iframe>

# Hardware

- [Raspberry Pi 5](https://www.raspberrypi.com/products/raspberry-pi-5/)
- [Adafruit Feather RP2040 SCORPIO](https://www.adafruit.com/product/5650)
- 8x [Neopixel 144 strands](https://www.adafruit.com/product/2847)
- [Custom MacroPaw keyboard](https://github.com/kodachi614/macropaw) by [@kodachi614](https://github.com/kodachi614)

<a href="assets/images/bbox2025_fritzing.webp" data-lightbox="bbox2025_fritzing" data-title="Beatboxer 2025 Fritzing"><img src="assets/images/bbox2025_fritzing.webp" alt="Beatboxer 2025 Fritzing" class="thumbnail mid"></a>

<a href="assets/images/bbox2025_fritzing.fzz">Click here</a> to download the source Fritzing sketch file (requires the <a href="https://github.com/adafruit/Fritzing-Library">Adafruit Fritzing Library</a> and a <a href="https://github.com/siggy/macropaw/blob/main/Beatboxer/renders/macropaw_beatboxer_small.fzpz">custom MacroPaw part</a>).

# Software

- [main.go](https://github.com/siggy/bbox/blob/main/cmd/bbox/main.go)
- [generic program interface](https://github.com/siggy/bbox/blob/main/pkg/program/program.go#L16-L32)
- [SCORPIO integration](https://github.com/siggy/bbox/tree/main/scorpio)

# Build photos

<a
href="assets/images/bbox2025_board.webp" data-lightbox="bbox2025" data-title="System on a board"><img
  src="assets/images/bbox2025_board.webp" alt="System on a board" class="thumbnail mid"></a><a
href="assets/images/bbox2025_preboard.webp" data-lightbox="bbox2025" data-title="Pre-board installation"><img
  src="assets/images/bbox2025_preboard.webp" alt="Pre-board installation" class="thumbnail mid"></a><a
href="assets/images/bbox2025_scorpio.webp" data-lightbox="bbox2025" data-title="SCORPIO install"><img
  src="assets/images/bbox2025_scorpio.webp" alt="SCORPIO install" class="thumbnail mid"></a><a
href="assets/images/bbox2025_macropaw.webp" data-lightbox="bbox2025" data-title="MacroPaw install"><img
  src="assets/images/bbox2025_macropaw.webp" alt="MacroPaw install" class="thumbnail mid"></a><a
href="assets/images/bbox2025_playa.webp" data-lightbox="bbox2025" data-title="Complete on playa"><img
  src="assets/images/bbox2025_playa.webp" alt="Complete on playa" class="thumbnail mid"></a><a
href="assets/images/bbox2025_dust.webp" data-lightbox="bbox2025" data-title="After the dust storm"><img
  src="assets/images/bbox2025_dust.webp" alt="After the dust storm" class="thumbnail mid">
</a>

## Contact

<a href="https://github.com/siggy/bbox/issues">File an issue against this repo</a>, or find me at <a href="https://sig.gy">sig.gy</a>

This doc updated for 2025. For the original 2017-2018 writeup, see [Beatboxer 2017-2018](2017-2018).

<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.4/jquery.min.js"></script>
<script src="assets/js/lightbox.min.js"></script>

<script>
  (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
  (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
  m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
  })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');
  ga('create', 'UA-27834075-1', 'auto');
  ga('send', 'pageview');
</script>
