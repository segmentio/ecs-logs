# ecs-logs [![CircleCI](https://circleci.com/gh/segmentio/ecs-logs.svg?style=shield)](https://circleci.com/gh/segmentio/ecs-logs)

ecs-logs is a log forwarder for services ran by ecs-agent.

## Quick Start

The simplest way to use ecs-logs is to run it as a docker container, you'll want
to configure the ecs-agent to forward services logs to journald first, here's
how to do it:

- run ecs-agent with `ECS_AVAILABLE_LOGGING_DRIVERS=["journald"]` set in the
environment variables.

- configure the task definitions of your services to use the journald driver by
adding the following:
```js
"logConfiguration": {
  "logDriver": "journald",
  "options": {
    "tag": "<your service name>"
  }
}
```
*Note: The tag is important here since it will be used as the group name for the
CloudWatch logs*

Once you have ecs-agent properly configured (you should be able to see your ECS
services logs in the journal), you can start ecs-logs this way:
```
docker run -t -i -v /run/logs/journal:/run/logs/journal:ro \
    segment/ecs-logs:latest -src journald -dst cloudwatchlogs
```
That's it! The services logs should now be showing up in CloudWatch Logs.

### Docker Image

- https://hub.docker.com/r/segment/ecs-logs

### Usage on OSX

If you're developing on OSX it may be inconvenient to not have the system
journal available for testing. One way that this can be worked around is using
the *stdin* source and piping your service's logs through [jq](https://stedolan.github.io/jq/)
to pack well formatted messages.  
Here's an example:
```shell
... | jq '. | {group: "<group>", stream: "<stream>", event: .}' | ecs-logs -src stdin -dst ...
```
*Note that This does require your service to output JSON formatted logs with a
structure that ecs-logs recognize.*
