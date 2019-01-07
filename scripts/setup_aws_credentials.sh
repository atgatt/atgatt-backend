#!/usr/bin/env bash
set -x
set -e

aws configure --profile eb-cli set aws_access_key_id $AWS_ACCESS_KEY_ID
aws configure --profile eb-cli set aws_secret_access_key $AWS_SECRET_ACCESS_KEY
