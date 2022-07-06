#!/bin/bash

go build cmd/justone/justonemain.go
scp justonemain deploy@stldevs.com:~
ssh deploy@stldevs.com << EOF
  service justone stop
  mv -f ~/justonemain /opt/justone/justone
  chmod +x /opt/justone/justone
  servive justone start
EOF
