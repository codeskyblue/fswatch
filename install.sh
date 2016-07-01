PROJECT="fswatch"
LATEST_RELEASE_URL="https://api.github.com/repos/franciscocpg/$PROJECT/releases/latest"

testRoot() {
	if [ "$(id -u)" != "0" ]; then
		echo "You must be root to run this script"
		exit 1
	fi
}


initArch() {
	ARCH=$(uname -m)
	case $ARCH in
		arm*) ARCH="arm";;
		x86) ARCH="386";;
		x86_64) ARCH="amd64";;
	esac
}

initOS() {
    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')
}

downloadFile() {
	LATEST_RELEASE_JSON=$(curl -s "$LATEST_RELEASE_URL")
	TAG=$(echo "$LATEST_RELEASE_JSON" | grep 'tag_' | cut -d\" -f4)
	PROJECT_DIST="${PROJECT}_${OS}_${ARCH}"
	# || true forces this command to not catch error if grep does not find anything
	DOWNLOAD_URL=$(echo "$LATEST_RELEASE_JSON" | grep 'browser_' | cut -d\" -f4 | grep "$PROJECT_DIST") || true
	if [ -z "$DOWNLOAD_URL" ]; then
        echo "Sorry, we dont have a dist for your system: $OS $ARCH"
        exit 1
	else
		TMP_FILE="/tmp/$PROJECT_DIST"
        echo "Downloading $DOWNLOAD_URL"
        curl -L "$DOWNLOAD_URL" -o "$TMP_FILE"
	fi
}

installFile() {
	BIN_FILE="/usr/local/bin/$PROJECT"
	cp "$TMP_FILE" "$BIN_FILE"
	chmod +x "$BIN_FILE"
}

bye() {
	result=$?
	if [ "$result" != "0" ]; then
		echo "Fail to install $PROJECT"
	fi
	exit $result
}

# Execution

#Stop execution on any error
set -e
trap "bye" EXIT
testRoot
initArch
initOS
downloadFile
installFile
# Test if everything ok
PROJECT_VERSION=$($PROJECT -version)
echo "$PROJECT $PROJECT_VERSION installed succesfully"
