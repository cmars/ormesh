#!/bin/bash

set -eu

RELEASE_VERSION=0.2.0

echo "deb http://deb.torproject.org/torproject.org xenial main" | sudo tee /etc/apt/sources.list.d/tor.list
echo "deb-src http://deb.torproject.org/torproject.org xenial main" | sudo tee -a /etc/apt/sources.list.d/tor.list
sudo apt-key adv --keyserver keys.gnupg.net --recv-keys A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89
sudo apt update
sudo apt upgrade -yy
sudo apt install -y tor deb.torproject.org-keyring

tmpdir=$(mktemp -d)
trap "rm -rf ${tmpdir}" EXIT

cd ${tmpdir}
wget -O ormesh.tar.gz https://github.com/cmars/ormesh/releases/download/v${RELEASE_VERSION}/ormesh_${RELEASE_VERSION}_linux_amd64.tar.gz
tar xf ormesh.tar.gz
sudo systemctl stop ormesh || true
sudo cp ormesh /usr/bin/ormesh
/usr/bin/ormesh agent privbind || echo "warning: setting privileged port bind capability failed"

if [ "$EUID" -eq 0 ]; then
	id -u ubuntu
	sudo su - ubuntu -c \
		'/usr/bin/ormesh agent systemd' | tee /etc/systemd/system/ormesh.service
else
	/usr/bin/ormesh agent systemd | sudo tee /etc/systemd/system/ormesh.service
fi

sudo systemctl enable ormesh
sudo systemctl start ormesh
