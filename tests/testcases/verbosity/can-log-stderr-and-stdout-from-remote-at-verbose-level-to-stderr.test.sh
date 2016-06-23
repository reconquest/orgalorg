# note! tests:ensure will eat '>&2' if it's passed without prefix
tests:ensure :orgalorg:with-key -v -C -- echo 1 \; echo err\>\&2

tests:assert-stderr-re '<stdout> \[.*\] 1'
tests:assert-stderr-re '<stderr> \[.*\] err'
