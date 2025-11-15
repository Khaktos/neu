#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

for file in ./tests/*_output.txt; do
  rm "$file"
done

go build

for file in ./tests/*.neu; do
  ./neu "$file" &> "$file"_output.txt
done

has_fail=false

for file in ./tests/*_output.txt; do
  if grep -q "panic" "$file"; then
    has_fail=true
    echo -e "$(basename "$file"): ${RED}failed with panic${NC}"
    cat "$file"
  elif cmp -s "$file" ./tests/expected/$(basename "$file"); then
    echo -e "$(basename "$file"): ${GREEN}passed${NC}"
  else
    has_fail=true
    echo -e "$(basename "$file"): ${RED}failed"
    diff "$file" ./tests/expected/$(basename "$file")
    echo -e ${NC}
  fi
done

if [ "$has_fail" = true ]; then
  echo "Tests done! With failures :("
else
  echo "Tests done!"
fi
rm neu
