#!/bin/bash
set -eo pipefail
# convert-flac-to-opus.sh by iotku (https://github.com/iotku/)
# License: wtfpl (http://www.wtfpl.net/txt/copying/)
# Description: Copy Music folder while encoding flac to opus. Will maintain other files.
# Requirements:
#    - Bash (obviously)
#    - GNU coreutils, parallel, findutils
#    - opusenc

# Options
# Paths
SOURCE_DIR="${HOME}/Music/"
OUTPUT_DIR="${HOME}/Music-Converted/"
LOG_FILE="/tmp/log-convfto.txt"
ERROR_LOG_FILE="/tmp/err-convfto.txt"
SCRIPT_PATH=$(realpath "${BASH_SOURCE[0]}")
# Opus VBR Bitrate
BITRATE=192
# How many threads to use for opus encoding
THREADS=$(nproc) # All CPU Threads, or change to a number (i.e 6)
#THREADS=4 # or n threads

# Don't change these!
TRIM_PREFIX_AMT="${#SOURCE_DIR}"

# Verify commands
if ! command -v parallel >/dev/null 2>&1; then
    echo "Error: 'parallel' command not found." >&2
    exit 1
fi

# Check that it's GNU Parallel
if ! parallel --version 2>&1 | grep -q "GNU parallel"; then
    echo "Error: 'parallel' found but it is not GNU Parallel." >&2
    exit 1
fi

function main {
     # Read Command Line Arguments
    for i in "$@"; do
        case "$i" in
            encode) shift; _encode "$*"; exit;;
            *) ;; # Default, just continue.
        esac
    done
    # Functions: Can be commented out if you only want to do certain tasks.
    syncDirs      # Create directories matching SOURCE_DIR into OUTPUT_DIR
    transcodeFlac # Attempt To Transcode All flac (with .flac ext) files into opus using opusenc
    copyNonFlacs  # Copy all other files into the same locations inside OUTPUT_DIR

    # TODO rename directories to opus
}

function syncDirs {
    echo "Syncing directories"
    # Get Directories to copy to new location (incl. non-flac dirs to be copied into later)
    find "$SOURCE_DIR" -mindepth 1 -type d -print0 |
        while IFS= read -r -d '' line; do
            mkdir -vp "$OUTPUT_DIR${line:$TRIM_PREFIX_AMT}" ||
                logError "MKDIRFAIL" "$OUTPUT_DIR${line:$TRIM_PREFIX_AMT}"
        done
}

# Find all .flac files and use opusenc (via _encode) in parallel
function transcodeFlac {
    echo "Transcoding flacs from $INPUT_DIR to $OUTPUT_DIR"
    find "$SOURCE_DIR" -type f -name "*.flac" | parallel --progress -j "$THREADS" bash "$SCRIPT_PATH" encode '{}'
}

function _encode {
    INPUT_DIR="$*"
    OUTPUT_PATH="${INPUT_DIR:$TRIM_PREFIX_AMT}"                     # Remove Original DIR Prefix
    OUTPUT_PATH="$OUTPUT_DIR${OUTPUT_PATH::${#OUTPUT_PATH}-5}.opus" # .flac -> .opus ext
    if test -f "$OUTPUT_PATH"; then
        echo "File $OUTPUT_PATH exists, not encoding." >> "$LOG_FILE"
        exit
    fi
    echo "Encoding -> $OUTPUT_PATH"
    opusenc --music --bitrate $BITRATE --comp 10 "$INPUT_DIR" "${OUTPUT_PATH}" 2>&1 | grep Runtime -A1 | xargs &&
        logSuccess "$INPUT_DIR" ||
        logError "EF" "$INPUT_DIR"
}

function copyNonFlacs {
    echo "Copying non-flacs to $OUTPUT_DIR"
    find "$SOURCE_DIR" -type f -not -name "*.flac" -print0 |
        while IFS= read -r -d '' line; do
            if test -f "$OUTPUT_DIR${line:$TRIM_PREFIX_AMT}"; then
                # don't copy, file already exists.
                echo "Not copying, file exists: $line"
            else
                cp -v "$line" "$OUTPUT_DIR${line:$TRIM_PREFIX_AMT}" ||
                    logError "COPYFAIL" "$line"
            fi
        done
}

function renameDirectories {
    find "$OUTPUT_PATH" -mindepth 1 -type d |
        while read line; do
            currentDIR="$(basename "$line")"
        done
}

function logSuccess {
    echo -e "\033[0;32mEncode Success\033[0m"
    echo -e "ES: $1" >> "$LOG_FILE"
}

function logError {
    echo -e "\033[0;31m$1: $2\033[0m"
    echo -e "$1: $2" >> "$ERROR_LOG_FILE"
}

# actually run
main "$@"
