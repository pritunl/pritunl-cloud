#!/bin/bash
set -e

webpack-cli --config webpack.dev.config --progress --color --watch
