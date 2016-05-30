tests:ensure :orgalorg-key -e -C echo 'two  spaces'

tests:assert-stdout "two  spaces"
