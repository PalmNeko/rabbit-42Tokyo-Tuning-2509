#!/bin/bash

cd webapp || exit 1

if [[ $HOSTNAME == ftt2508-* ]]; then
	if [[ "$(docker compose ps | wc -l)" > 5 ]]; then
		echo "OK"
		exit 0
	else
		echo "NG"
		exit 1
	fi
else
	if [[ "$(docker compose -f docker-compose.local.yml ps | wc -l)" > 5 ]]; then
		echo "OK"
		exit 0
	else
		echo "NG"
		exit 1
	fi
fi
