#!/bin/sh
! docker ps -aq --no-trunc | xargs docker rm --force
rm -f tmp/*