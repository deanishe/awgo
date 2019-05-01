# Workflow environment variables
# These variables create an Alfred-like environment
# root="$( dirname "$0" )"

root="$( git rev-parse --show-toplevel )"
testdir="${root}/testenv"

# Absolute bare-minimum for AwGo to function...
export alfred_workflow_bundleid="net.deanishe.awgo"
export alfred_workflow_data="${testdir}/data"
export alfred_workflow_cache="${testdir}/cache"
# export alfred_workflow_cache="$HOME/Library/Caches/com.runningwithcrayons.Alfred/Workflow Data/net.deanishe.awgo"
# export alfred_workflow_data="$HOME/Library/Application Support/Alfred/Workflow Data/net.deanishe.awgo"

test -f "$HOME/Library/Preferences/com.runningwithcrayons.Alfred.plist" || {
	export alfred_version="3.8.1"
	# export alfred_workflow_cache="$HOME/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data/net.deanishe.awgo"
	# export alfred_workflow_data="$HOME/Library/Application Support/Alfred 3/Workflow Data/net.deanishe.awgo"
}

# Expected by ExampleNew
export alfred_workflow_version="1.2.0"
export alfred_workflow_name="AwGo"

# Prevent random ID from being generated
export AW_SESSION_ID="test-session-id"

export GO111MODULE=on
