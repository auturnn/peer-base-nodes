#!/bin/bash


echo "> create service"

cat <<EOF> /etc/systemd/system/kickshaw-coin.service
[Unit]
Description=kickshaw-coin base-node
After=syslog.target

[Service]
User=ec2-user
Group=ec2-user

ExecStart=/home/ec2-user/build/peer-base-nodes
WorkingDirectory=/home/ec2-user/build

RestartSec=10
Restart=always

[Install]
WantedBy=multi-user.target
EOF

echo "> start service(daemon)"

systemctl daemon-reload
systemctl restart kickshaw-coin
