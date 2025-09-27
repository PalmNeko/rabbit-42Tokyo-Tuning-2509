
default: help

up:
	bash is_upped.sh || bash restore_and_migration.sh

restart:
	bash restore_and_migration.sh

down:
	bash stop.sh

help:
	echo "up		コンテナを起動します。(データベースの中身も消えます)"
	echo "restart	コンテナを再起動します。(データベースの中身も消えます)"
	echo "down		db, backendコンテナを停止します。"

"PATH": "${env:HOME}/bin:${env:PATH}"
