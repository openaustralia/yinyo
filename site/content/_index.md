---
title: Yinyo
date: 2019-09-11T21:07:08.000+00:00
headline: |
  A wonderfully simple API to reliably execute many long running scrapers in a
  super scaleable way.
  Built on top of Kubernetes.
quotes:
  - name: Matthew Landauer
    title: cofounder
    company: OpenAustralia Foundation
    photo: matthew.jpg
    text: This is going to make morph.io so much better! I wish we had done this *ages* ago.
  - name: Matthew Landauer
    title: cofounder
    company: OpenAustralia Foundation
    photo: matthew.jpg
    text: This is going to make morph.io so much better! I wish we had done this *ages* ago.
  - name: Matthew Landauer
    title: cofounder
    company: OpenAustralia Foundation
    photo: matthew.jpg
    text: This is going to make morph.io so much better! I wish we had done this *ages* ago.
features:
  - heading: Easy and scaleable
    text: |
      Easily run as many scrapers as you like across a cluster of machines without
      having to sweat the details. Powered by [Kubernetes](https://kubernetes.io/).
  - heading: The languages you love
    text: |
      Use the language and libraries you love for writing scrapers.
      Supports Python, JavaScript, Ruby, PHP and Perl via Heroku Buildpacks.
  - heading: Flexible
    text: |
      Supports many different use cases through a simple, yet flexible API that can operate synchronously
      or asynchronously.
  - heading: Open source, liberally licensed
    text: |
      Made specifically for developers of scraper systems be it open source or
      commercial. No chance of vendor lock-in because it's open source,
      under the [Apache 2.0 license](https://github.com/openaustralia/yinyo/blob/master/LICENSE).
    draft: true
---

### Why we created Yinyo

Yinyo comes out of OpenAustralia Foundation’s many years of experience writing scrapers and _hosting the world’s largest site for open scrapers_, [morph.io](http://morph.io).

Originally we designed morph.io to run on top of Docker. As the site has continued to grow over the years we have hit the limitations of some of the early design decisions we made. That’s why we created Yinyo.

It is intended as a much improved, much more scaleable low level underpinnings of morph.io. But it’s a lot more than that.

It’s _intended as a foundation upon which other developers can build other scraper systems_ quite different from morph.io. Yinyo isn’t opinionated about how scrapers store their data or what languages they’re written in. Yet using it to develop your scraper system will _save you an enormous amount of effort_.
