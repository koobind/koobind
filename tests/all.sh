#!/usr/bin/env bash

fails=0

ctest users.yml | bash
((fails=fails+$?))

echo ""
ctest groups.yml | bash
((fails=fails+$?))

echo ""
ctest groupbindings.yml | bash
((fails=fails+$?))

echo ""
echo "Global result: $fails fail(s)"

