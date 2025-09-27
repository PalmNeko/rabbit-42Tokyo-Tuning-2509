#!/bin/bash

bash is_upped.sh || exit 0

cd webapp || exit 1


if [[ $HOSTNAME == ftt2508-* ]]; then
    HOSTNAME=$HOSTNAME docker compose down
else
	echo "コンテナを停止します。"
    docker compose -f docker-compose.local.yml down
fi
