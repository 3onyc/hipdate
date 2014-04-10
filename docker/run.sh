#!/bin/bash

export REDIS_URL="${REDIS_PORT_6379_TCP/tcp:/redis:}"
exec /usr/bin/hipdate
