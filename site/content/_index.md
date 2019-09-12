---
title: "Leila"
date: 2019-09-12T07:07:08+10:00
draft: true
---

* Made specifically for developers of scraper systems
* Easily run many many scrapers across a cluster without having to worry about managing machines and how things are scheduled
* Built on top of Kubernetes
* Supports Python, JavaScript, Ruby, PHP and Perl
* Language support uses Heroku Buildpacks
* Optionally logs scraper standard and error asynchronously to an http endpoint of your choice
* Choose between a synchronous or asynchronous API depending on your use case
* Open source. MIT licensed

*Put features in a block list with an icon for each feature block*

[Get Started](#get-started-in-the-documentation)

### What’s a scraper anyway?

A scraper is a program that collects data from the web (usually unstructured) and reformats into structured data. The input is the whole web (and optionally some state like the output of a previous run).The output is some data. Scrapers can often run a long time because they need to traverse a large number of web pages. It’s not unusual for scrapers to run for hours. This means that many technologies are not particularly well suited to running scrapers.

### Origins

Clay comes out of OpenAustralia Foundation’s many years of experience writing scrapers and hosting the world’s largest site for open scrapers, morph.io. Originally we designed morph.io to run on top of Docker. As the site has continued to grow over the years we have hit the limitations of some of the early design decisions we made. That’s why we created Clay. It is intended as a much improved, much more scaleable low level underpinnings of morph.io. But it’s a lot more than that. It’s intended as a foundation upon which other developers can build other scraper systems quite different from morph.io. Clay isn’t opinionated about how scrapers store their data or what languages they’re written in. Yet using it do develop your scraper system will save you an enormous amount of effort

### Quotes from people

*They can be some irreverent made up quotes to start with or quotes from current users of morph who appreciate where this project is coming from*

Quote
Name and title
Photograph

### More social validation

Used in production to run scrapers more than xxxxx times

### Why is it called Leila?

Matthew’s mother’s dog is called Leila. And no he doesn’t use that as a password security question.

### Contact

Email address
GitHub
Suggest a feature or report an issue


Made with love by the OpenAustralia Foundation for you
