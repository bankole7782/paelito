#! /bin/bash

echo "Installing dependencies"

apt install libgtk-3-dev -y
apt install libwebkit2gtk-4.0-dev -y

echo "Fetching Assets"
mkdir -p /opt/saenuma/paelito
wget -q https://storage.googleapis.com/pandolee/paelito/20211126T034720WAT/paelito.tar.xz -O /opt/saenuma/paelito.tar.xz
tar -xf /opt/saenuma/paelito.tar.xz -C /opt/saenuma/paelito
cp /opt/saenuma/paelito/paelito.desktop /usr/share/applications
chmod +x /opt/saenuma/paelito/paelito

echo "All done."