#!/bin/bash
set -euf -o pipefail
# Symlink Dispatch

# Create Symlinks from multiple directories while keeping track of modification 
# times as not to duplicate symlinks should they be modified in the future.

# By @iotku
# License WTFPL (http://www.wtfpl.net/txt/copying/)

# Global Scope (Declare instead?)
STARTTIME="$(date --rfc-3339=seconds)"
export STARTTIME
export CONFDIR="${HOME}/.config/iotku/"
export CONFNAME="symdis-lastrun"
export CONFIGFILE="${CONFDIR}${CONFNAME}"

function setLastTime () {

	# If last time not saved yet (first run)
	
	# Create Dir
	if [ ! -d "$CONFDIR" ]; then
		echo "Making config dir: $CONFDIR"
		mkdir -p "$CONFDIR"
	fi

	# Create Config File	
	if [ ! -f "$CONFIGFILE" ]; then
		touch "$CONFIGFILE"
	fi

	# Save time to file 
	echo "$STARTTIME" > "$CONFIGFILE"
}

function getLastTime () {
	# General date Format rfc-3339
	# Example LASTTIME="2020-03-01 12:00:00-08:00"
	if [ ! -f "$CONFIGFILE" ]; then
		LASTTIME="none"
	else
		LASTTIME="$(< "$CONFIGFILE")"
	fi
	echo "$LASTTIME" # Return Last Time Run
}

function makeLinks () {
	LINKPATH="${@: -1}" # Final function argument
	LASTTIME="$(getLastTime)"

	for SRCDIR in "${@: 1: $# -1}"
	do
		if [ "$LASTTIME" == "none" ];
		then
			echo "First Run: Copying All Symlinks -> $LINKPATH"
			# Copy All Symlinks
			find "$SRCDIR" -maxdepth 1 -not -wholename "$SRCDIR" -exec ln -sv "{}" "$LINKPATH" \;
		else
			echo "Creating Symlinks from paths newer than $LASTTIME from $SRCDIR -> $LINKPATH"
			# Copy Newer Symlinks than last time
			find "$SRCDIR" -maxdepth 1 -newermt "$LASTTIME" -not -wholename "$SRCDIR" -exec ln -sv "{}" "$LINKPATH" \;
		fi
	done
}

function main () {
	# makeLinks takes "any" amount of paths, and then symlinks them to last path provided
	makeLinks "/path1/" "/path2/" "/symlink-destination/"

	# Save last Time for later
	setLastTime
}

# Main script
main
