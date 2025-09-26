#!/bin/bash

PROJECT_PATH="$(cd $(dirname $BASH_SOURCE[0]); pwd)"
. "$PROJECT_PATH"/.env

ssh -i "$SSH_PEM" "$SSH_ADDRESS"
