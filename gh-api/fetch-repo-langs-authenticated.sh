# fetch repo langs for REPO
# REPO is in the form of usr/repo
# USR = username
# TOKEN = api token
curl -u $USR:$TOKEN "https://api.github.com/repos/$REPO/languages"

