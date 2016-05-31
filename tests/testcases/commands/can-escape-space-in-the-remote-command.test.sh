tests:ensure :orgalorg:with-key -e -C echo 'two  spaces'

tests:assert-stdout "two  spaces"
