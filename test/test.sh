#!/bin/bash
#server="http://127.0.0.1:9999"
server="http://47.88.53.255:9999"
url='http://shmmsns.qpic.cn/mmsns/qbvaL9taELtDlcft4nFHb6y1icSq63iaozLrtPpZxY632WOLKhwvuDYIDjiblVurvvQSDdgOhFhTCw/0'
curl -vvv -H "Content-Type: application/json" -d '{"url": "http://shmmsns.qpic.cn/mmsns/qbvaL9taELtDlcft4nFHb6y1icSq63iaozLrtPpZxY632WOLKhwvuDYIDjiblVurvvQSDdgOhFhTCw/0"}' $server/limit
