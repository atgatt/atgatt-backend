#!/usr/bin/env bash
set -x
set -e

sudo apt-get -y -qq install awscli python-pip
pip install awsebcli --upgrade --user
echo 'export PATH=$PATH:~/.local/bin' >> $BASH_ENV
source ~/.bashrc
chmod +x ./scripts/setup_aws_credentials.sh && ./scripts/setup_aws_credentials.sh
cp workspace/artifacts/api-artifacts.zip .
cp workspace/artifacts/worker-artifacts.zip .
mv api-artifacts.zip artifacts.zip
eb deploy api-$1 --label api-${CIRCLE_SHA1} --process --verbose
rm artifacts.zip && mv worker-artifacts.zip artifacts.zip
eb deploy worker-$1 --label worker-${CIRCLE_SHA1} --process --verbose