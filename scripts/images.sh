#!/bin/bash
rm -rf /tmp/board

mkdir /tmp/board
wget -qO /tmp/board/master.zip "https://github.com/battlesnakeio/board/archive/master.zip" 
unzip -q /tmp/board/master -d /tmp/board

rm -rf snake-images
mkdir -p snake-images/head
mkdir -p snake-images/tail
qlmanage -t -s 35 -o snake-images/head /tmp/board/board-master/public/images/snake/head/*.svg
qlmanage -t -s 35 -o snake-images/tail /tmp/board/board-master/public/images/snake/tail/*.svg