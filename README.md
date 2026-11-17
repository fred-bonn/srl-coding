# srl-coding

**Spaced repetition learning for coding.**

## The idea

A user starts a learning process and is given a coding challenge for a
topic/tag — for example `"printing"` or `"addition"`.

## The flow

```
start
  │
  ▼
topic introduced for the first time
  │
  ▼
static challenge for that topic          (always the same, first introduction only)
  │
  ▼
spaced repetition schedules the topic's next review
  │
  ▼
review challenge for the topic           (freshly generated
  │                                       or fetched from the
  └──────────────▶ back to scheduling    challenge database)
```

In short:

1. **Start** — the user begins the learning process.
2. **First introduction** — each topic has a static challenge that is always
   given when the topic is first introduced.
3. **Review** — spaced repetition decides when the topic is presented to the
   user again, as a generated challenge: either freshly generated, or fetched
   from the database of existing challenges.
