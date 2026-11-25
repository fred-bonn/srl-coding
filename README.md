# srl-coding

**Spaced repetition learning for coding.**

## The idea

A user starts a learning process and is given a coding challenge for a
topic — for example `"printing"` or `"addition"`.

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

## API

The server lives in `cmd/server` and speaks plain JSON over HTTP, built on
the standard library. Set `DATABASE_URL` to connect to Postgres. `ADDR`
overrides the listen address (default `:8080`), `LESSONS_DIR` the lesson
content (default `data/lessons/python`), and `TEST_USERS` a comma-separated
list of emails whose accounts jump instantly to the next breakpoint instead
of waiting for the next calendar day.

Login issues a session token, valid for 30 days. Every learning endpoint
requires it in the `Authorization` header. Errors are JSON: every non-2xx
response carries `{"error": "<message>"}`.

A day's work is one lesson plus the queue of topics due that day. The
lesson's topic is introduced into the queue the first time the lesson is
shown. When the lesson is completed and the queue is empty, the next
breakpoint is reached: real users get the next day's lesson on the next
calendar day, test users immediately.

- POST /auth/register: Creates a user account.
  - Input: JSON with the following fields:
    ```
    {
        email: <String: user email>,
        password: <String: at least 8 characters>,
    }
    ```
  - Headers:
    - `Content-Type: application/json`
  - Response: `201` with the following fields:
    ```
    {
        id: <String: user ID>,
        email: <String: user email>,
    }
    ```
  - Response Headers:
    - `Content-Type: application/json`
  - Status Codes:
    - `201`: Account created.
    - `400`: Invalid JSON body, invalid email, or password too short.
    - `405`: Wrong method.
    - `409`: Email already registered.

- POST /auth/login: Logs in and returns a session token.
  - Input: JSON with the following fields:
    ```
    {
        email: <String: user email>,
        password: <String: the account password>,
    }
    ```
  - Headers:
    - `Content-Type: application/json`
  - Response: `200` with the following fields:
    ```
    {
        token: <String: session token, valid 30 days>,
    }
    ```
  - Response Headers:
    - `Content-Type: application/json`
  - Status Codes:
    - `200`: Login succeeded.
    - `400`: Invalid JSON body.
    - `401`: Wrong email or password.
    - `405`: Wrong method.

- GET /today: Returns the day's work: the current lesson, the queue of due
  topics, and any open attempts with their saved code.
  - Input: None.
  - Headers:
    - `Authorization: Bearer <token>`
  - Response: `200` with the following fields:
    ```
    {
        lesson: <Object|null: current lesson, null when today's lesson is
            already completed or the curriculum is done>,
            {
                id: <String: lesson ID, e.g. "python-001-printing">,
                number: <Number: position in the curriculum>,
                content: <String: lesson markdown>,
                assignment: <String: assignment text>,
                challenge_id: <String: intro challenge ID>,
            },
        queue: <Array: topics due today and not yet done>,
            {
                topic_id: <String: topic ID>,
                name: <String: topic name>,
            },
        open_attempts: <Array: started but not yet completed attempts>,
            {
                id: <String: attempt ID>,
                challenge_id: <String: challenge ID>,
                assignment: <String: assignment text>,
                code: <String: last saved editor state>,
            },
        done_for_today: <Boolean: no work left today>,
    }
    ```
  - Response Headers:
    - `Content-Type: application/json`
  - Status Codes:
    - `200`: Today's work returned.
    - `401`: Missing, invalid, or expired session token.
    - `405`: Wrong method.

- POST /attempts: Presents a challenge and starts an attempt. If an open
  attempt for the challenge already exists, it is returned unchanged, so
  this is also the resume entry point.
  - Input: JSON with the following fields:
    ```
    {
        challenge_id: <String: ID of an assigned challenge>,
    }
    ```
  - Headers:
    - `Content-Type: application/json`
    - `Authorization: Bearer <token>`
  - Response: `201` with the following fields:
    ```
    {
        id: <String: attempt ID>,
        challenge_id: <String: challenge ID>,
        assignment: <String: assignment text>,
        code: <String: current editor state, starts as the starter code>,
    }
    ```
  - Response Headers:
    - `Content-Type: application/json`
  - Status Codes:
    - `201`: Attempt created (or the open one returned).
    - `400`: Invalid JSON body or invalid challenge ID.
    - `401`: Missing, invalid, or expired session token.
    - `403`: Challenge is neither today's lesson's nor linked to a topic due
      today.
    - `404`: No challenge with this ID.
    - `405`: Wrong method.

- PUT /attempts/{id}/code: Saves the current editor state of an open
  attempt, enabling AFK resumes.
  - Input: JSON with the following fields:
    ```
    {
        code: <String: current editor state>,
    }
    ```
  - Headers:
    - `Content-Type: application/json`
    - `Authorization: Bearer <token>`
  - Response: `204`, no body.
  - Response Headers: None required.
  - Status Codes:
    - `204`: Code saved.
    - `400`: Invalid JSON body or invalid attempt ID.
    - `401`: Missing, invalid, or expired session token.
    - `404`: No such attempt for this user.
    - `405`: Wrong method.
    - `409`: Attempt already completed.

- POST /attempts/{id}/complete: Completes an attempt, records the review,
  and reschedules the challenge's topics.
  - Input: JSON with the following fields:
    ```
    {
        correct: <Boolean: whether the solution passed the tests>,
        duration_seconds: <Number: active solving time in seconds>,
    }
    ```
  - Headers:
    - `Content-Type: application/json`
    - `Authorization: Bearer <token>`
  - Response: `204`, no body.
  - Response Headers: None required.
  - Status Codes:
    - `204`: Attempt completed.
    - `400`: Invalid JSON body or invalid attempt ID.
    - `401`: Missing, invalid, or expired session token.
    - `404`: No such attempt for this user.
    - `405`: Wrong method.
    - `409`: Attempt already completed.

- POST /lessons/complete: Marks today's lesson as completed. One lesson
  per day; the next lesson becomes current tomorrow, or immediately for
  test users.
  - Input: None.
  - Headers:
    - `Authorization: Bearer <token>`
  - Response: `204`, no body.
  - Response Headers: None required.
  - Status Codes:
    - `204`: Lesson completed.
    - `401`: Missing, invalid, or expired session token.
    - `405`: Wrong method.
    - `409`: A lesson was already completed today, or the curriculum is
      done.

A full day's flow with `curl`:

```bash
curl -X POST localhost:8080/auth/register \
    -H 'Content-Type: application/json' \
    -d '{"email":"me@example.com","password":"password123"}'

TOKEN=$(curl -s -X POST localhost:8080/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"email":"me@example.com","password":"password123"}' | jq -r .token)

# fetch today's work and grab the lesson's intro challenge
curl -s localhost:8080/today -H "Authorization: Bearer $TOKEN"

# start the attempt, save some progress, then submit
ATTEMPT=$(curl -s -X POST localhost:8080/attempts \
    -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
    -d '{"challenge_id":"<challenge_id from /today>"}')
ID=$(echo "$ATTEMPT" | jq -r .id)

curl -X PUT localhost:8080/attempts/$ID/code \
    -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
    -d '{"code":"print(\"Hello, World!\")"}'

curl -X POST localhost:8080/attempts/$ID/complete \
    -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
    -d '{"correct":true,"duration_seconds":120}'

curl -X POST localhost:8080/lessons/complete \
    -H "Authorization: Bearer $TOKEN"

# the next breakpoint: /today now shows no lesson (or, for test users,
# the next lesson immediately)
curl -s localhost:8080/today -H "Authorization: Bearer $TOKEN"
```
