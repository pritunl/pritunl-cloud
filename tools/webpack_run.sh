#!/bin/bash
set -e

npx webpack-cli --config webpack.dev.config --progress --color --watch
