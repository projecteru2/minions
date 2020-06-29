cp --parents /eru/minions.conf /etc/
cp eru-minions.service /usr/lib/systemd/system/
systemctl enable eru-minions.service
systemctl start eru-minions.service