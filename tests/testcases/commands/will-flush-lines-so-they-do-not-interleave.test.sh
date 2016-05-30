# checking, that lines from different remote sources do not interleave
# each other
tests:ensure :orgalorg-key -C \
    seq 1 10000 '|' awk '{print $1}' '|' sort '|' uniq '|' wc -l

tests:assert-no-diff "$(containers:count)" stdout
