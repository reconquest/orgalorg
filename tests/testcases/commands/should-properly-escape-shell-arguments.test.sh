tests:ensure echo "'1'"
tests:ensure :orgalorg:with-key -e -C echo "'1'"

tests:assert-stdout-re "1$"

tests:ensure :orgalorg:with-key -e -C echo "\\'"

tests:assert-stdout-re "'$"
