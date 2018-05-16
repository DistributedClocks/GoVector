#!/bin/bash
FILE=TSViz.log
echo '(?<timestamp>\d+) (?<host>\S*) (?<clock>{.*})\n(?<event>.*)' > $FILE
echo -e "\n" >> $FILE
cat *g.txt >> $FILE

