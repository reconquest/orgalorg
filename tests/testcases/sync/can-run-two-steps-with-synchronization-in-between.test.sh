tests:put test-file <<EOF
line1
line2
EOF

tests:put sync <<EOF
read prefix message

nodes_count=\$(sed -un '/NODE/p;/START/q' | wc -l)

echo NODES COUNT: \$nodes_count

sleep \$((RANDOM%3))

echo PHASE 1

echo \$prefix SYNC 1

for (( i = 0; i < nodes_count; i++ )); do
    read message

    echo GOT: \$(cut -f2- -d' ' <<< \$message)
done

echo PHASE 2
EOF

containers:do :install-sync-command-into-container "sync"

# If synchronization fails, then there will be lines out of order.
# In correct ordering lines always will be sorted by phase number.
tests:runtime :orgalorg:with-key -e -r /home/orgalorg/ -S test-file \
    '|' grep -o 'PHASE .*' \
    '|' uniq \
    '|' wc -l

tests:assert-stdout '2'
