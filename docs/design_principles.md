# Design Principles

## Purpose of design principles

As we work on developing and improving things we make lots of design choices. Often those choices are informed by unspoken assumptions. Here we attempt to make those underlying assumptions explicit. Also if a few people are discussing a design it can be really useful to be able to point at a specific design principle and ask “how does it relate to this?” or “what does this tell us?”

## The design principles

- Your users are developers but don’t make the mistake that they like or want complexity. Far from it.
- Maintainability is crucial. This project should last and evolve long into the future.
- in early stages iterate quickly - don’t worry about making breaking changes - version these 0.x to make this explicit
- minimise [leaky abstractions](https://en.wikipedia.org/wiki/Leaky_abstraction)
- don’t be opinionated about how scrapers are written - you should be able to write them in different languages, use different databases. They might maintain state, they might not. All you can really say is that they’ll get stuff from the internet, probably run for between a minute and a day and produce some output. So, really they could be used for any kind of long running job.
- keep it simple - seems so obvious really you think it shouldn’t need to be said
- don’t try to do too much.
- the thing should feel like an integrated whole where what is does makes sense together but what it doesn’t do also feels clear. If this isn’t the case for whatever reason that probably means something is missing, something is there that shouldn’t be, or something fundamentally needs to be restructured or refactored
- the project should be very much a standalone project from morph.io. It’s about making something that is useful to larger set of people than those that might morph.io. We shouldn’t mention morph unless we’re talking about the history or motivation of the project.
- You should be able to be a user of the project without having to download it and get it running yourself
- Make it straightforward for people to move from being a user to a contributor
- When making changes always figure out how they can be broken into smaller changes. Get those changes deployed to production and used by other people as quickly as possible. Learn from that. Move to the next thing..
- Scaling and resource usage should be automatic and invisible as far as possible
