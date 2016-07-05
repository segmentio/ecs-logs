# ecs-logs [![CircleCI](https://circleci.com/gh/segmentio/ecs-logs.svg?style=shield)](https://circleci.com/gh/segmentio/ecs-logs) [![GoDoc](https://godoc.org/github.com/segmentio/ecs-logs?status.svg)](https://godoc.org/github.com/segmentio/ecs-logs)

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

### Sources

Sources are log streams from which events are read by ecs-logs and forwarded to
the destinations.  
The log events can be JSON formatted with the following structure:
```js
{
  "level": "<debug | info | notice | warn | error | crit | alert | emerg>",
  "time": "<iso8601 time representation>",
  "info": {
    "host": "<hostname>",
    "source": "<file:line:function>",
    "errors": [
      {
        "type": "<error type>",
        "error": "<error message>",
        "errno": <errno value>,
        "stack": [...]
      },
      ...
    ]
  },
  "data": {
    ...
  },
  "message": "<log message>"
}
```
All fields are optional and ecs-logs will assume defaults if some are missing.

- **stdin**

The default source that ecs-logs uses is *stdin*, in most cases this is not what
will be used in production but can be useful for development and testing
purposes.  
Because no metadata can be passed to log messages read from *stdin*, this source
expects a to read a stream of JSON-formatted objects with this following
structure:
```js
{
  "group": "<log group>",
  "stream": "<log stream>",
  "event": {
    ...
  }
}
```
Where *group* and *stream* will be used to identify where the log event belong
and *event* must be a JSON object with the structure defined above.

- **journald**

This journald source is what is usually used for production deployments since
ECS can be easily configured to send docker containers logs to the journal which
then acts as a centralize logging system on host.

The journald source expects to find the *CONTAINER_TAG* and *CONTAINER_NAME*
metadata on log events, which it uses to set the group and stream to which the
log event will be sent.

The log message can be easier plain text or JSON formatted. When ecs-logs fails
to parse a JSON message, either because the content is not JSON or because the
format is not something it understands, it will generate a log event where the
*message* field is set to the full log message, for example:

*log message:*
```
2016-07-05T09:08:12.123Z - INFO - hello!
```
*log event:*
```js
{
  "level": "NONE",
  "time": "<time at wich the message was sent to the journal>",
  "info": {
    "host": "<hostname>"
  },
  data: { },
  "message": "2016-07-05T09:08:12.123Z - INFO - hello!"
}
```

### Usage on OSX

If you're developing on OSX it may be inconvenient to not have the system
journal available for testing. One way that this can be worked around is using
the *stdin* source and piping your service's logs through [jq](https://stedolan.github.io/jq/)
to pack well formatted messages.  
Here's an example:
```shell
... | jq '. | {group: "<group>", stream: "<stream>", event: .}' | ecs-logs -src stdin -dst ...
```
*Note that it requires your service to output JSON formatted logs with a
structure that ecs-logs recognize.*
