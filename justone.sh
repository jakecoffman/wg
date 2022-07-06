#!/bin/bash

go build cmd/justone/justonemain.go
scp justonemain deploy@stldevs.com:~
ssh deploy@stldevs.com << EOF
  sudo service justone stop
  mv -f ~/justonemain /opt/justone/justone
  chmod +x /opt/justone/justone
  sudo servive justone start
EOF
